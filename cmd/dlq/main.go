package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tclavelloux/promy-event-bus/streams"
)

type dlqPayload struct {
	OriginalStream    string    `json:"original_stream"`
	OriginalEventID   string    `json:"original_event_id"`
	OriginalEventType string    `json:"original_event_type"`
	OriginalPayload   string    `json:"original_payload"`
	FailureReason     string    `json:"failure_reason"`
	FailedAt          time.Time `json:"failed_at"`
	FailedService     string    `json:"failed_service"`
	AttemptsExhausted int       `json:"attempts_exhausted"`
}

type replayOpts struct {
	stream string
	typ    string
	id     string
	all    bool
	dryRun bool
	limit  int64
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	var err error

	switch subcommand {
	case "inspect":
		err = runInspect(os.Args[2:])
	case "replay":
		err = runReplay(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: dlq <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  inspect   Show DLQ statistics")
	fmt.Println("  replay    Re-publish DLQ entries to their original streams")
}

func redisDefault() string {
	if v := os.Getenv("REDIS_URL"); v != "" {
		return v
	}

	return "redis://localhost:6379/0"
}

func newRedisClient(dsn string) (*redis.Client, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis DSN %q: %w", dsn, err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cannot connect to Redis: %w", err)
	}

	return client, nil
}

// --- inspect ---

func runInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	redisDSN := fs.String("redis", redisDefault(), "Redis DSN")
	limit := fs.Int64("limit", 10000, "Max entries to scan")
	_ = fs.Parse(args)

	client, err := newRedisClient(*redisDSN)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck // best-effort cleanup

	ctx := context.Background()

	length, err := client.XLen(ctx, streams.StreamDLQ).Result()
	if err != nil {
		return fmt.Errorf("reading DLQ length: %w", err)
	}

	fmt.Printf("%s length: %d\n", streams.StreamDLQ, length)

	if length == 0 {
		return nil
	}

	msgs, err := client.XRangeN(ctx, streams.StreamDLQ, "-", "+", *limit).Result()
	if err != nil {
		return fmt.Errorf("reading DLQ entries: %w", err)
	}

	printInspectReport(msgs)

	return nil
}

func printInspectReport(msgs []redis.XMessage) {
	type stats struct {
		count  int
		oldest time.Time
	}

	byStream := make(map[string]*stats)
	byType := make(map[string]int)
	byService := make(map[string]int)

	for _, msg := range msgs {
		entry, err := parseDLQMessage(msg)
		if err != nil {
			continue
		}

		s, ok := byStream[entry.OriginalStream]
		if !ok {
			s = &stats{oldest: entry.FailedAt}
			byStream[entry.OriginalStream] = s
		}

		s.count++

		if entry.FailedAt.Before(s.oldest) {
			s.oldest = entry.FailedAt
		}

		byType[entry.OriginalEventType]++
		byService[entry.FailedService]++
	}

	fmt.Println("\nBy original stream:")

	for stream, s := range byStream {
		age := time.Since(s.oldest).Truncate(time.Minute)
		fmt.Printf("  %-30s %d entries (oldest: %s ago)\n", stream+":", s.count, formatDuration(age))
	}

	fmt.Println("\nBy event type:")

	for typ, count := range byType {
		fmt.Printf("  %-30s %d\n", typ+":", count)
	}

	fmt.Println("\nBy service:")

	for svc, count := range byService {
		fmt.Printf("  %-30s %d\n", svc+":", count)
	}
}

// --- replay ---

func runReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	opts := replayOpts{}
	redisDSN := fs.String("redis", redisDefault(), "Redis DSN")
	fs.StringVar(&opts.stream, "stream", "", "Filter by original stream")
	fs.StringVar(&opts.typ, "type", "", "Filter by original event type")
	fs.StringVar(&opts.id, "id", "", "Replay a single entry by DLQ message ID")
	fs.BoolVar(&opts.all, "all", false, "Replay all entries")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "List without replaying")
	fs.Int64Var(&opts.limit, "limit", 1000, "Max entries to process")
	_ = fs.Parse(args)

	if !opts.all && opts.stream == "" && opts.typ == "" && opts.id == "" {
		fmt.Fprintln(os.Stderr, "Error: specify at least one filter (-stream, -type, -id) or -all")
		fs.Usage()

		return fmt.Errorf("no filter specified")
	}

	client, err := newRedisClient(*redisDSN)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck // best-effort cleanup

	ctx := context.Background()

	msgs, err := fetchDLQMessages(ctx, client, opts)
	if err != nil {
		return err
	}

	replayed, skipped, failed := processReplayMessages(ctx, client, msgs, opts)

	label := "Replayed"
	if opts.dryRun {
		label = "Would replay"
	}

	fmt.Printf("\n%s: %d, Skipped: %d, Failed: %d\n", label, replayed, skipped, failed)

	return nil
}

func fetchDLQMessages(ctx context.Context, client *redis.Client, opts replayOpts) ([]redis.XMessage, error) {
	var (
		msgs []redis.XMessage
		err  error
	)

	if opts.id != "" {
		msgs, err = client.XRangeN(ctx, streams.StreamDLQ, opts.id, opts.id, 1).Result()
	} else {
		msgs, err = client.XRangeN(ctx, streams.StreamDLQ, "-", "+", opts.limit).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("reading DLQ: %w", err)
	}

	return msgs, nil
}

func processReplayMessages(ctx context.Context, client *redis.Client, msgs []redis.XMessage, opts replayOpts) (replayed, skipped, failed int) {
	for _, msg := range msgs {
		entry, err := parseDLQMessage(msg)
		if err != nil {
			skipped++

			continue
		}

		if !matchesFilter(entry, opts) {
			skipped++

			continue
		}

		if opts.dryRun {
			fmt.Printf("[dry-run] %s | stream=%s type=%s service=%s reason=%q\n",
				msg.ID, entry.OriginalStream, entry.OriginalEventType, entry.FailedService, entry.FailureReason)
			replayed++

			continue
		}

		if err := replayEntry(ctx, client, msg.ID, entry); err != nil {
			fmt.Fprintf(os.Stderr, "  FAIL %s: %v\n", msg.ID, err)
			failed++

			continue
		}

		replayed++
	}

	return replayed, skipped, failed
}

func matchesFilter(entry *dlqPayload, opts replayOpts) bool {
	if opts.all {
		return true
	}

	if opts.id != "" {
		return true
	}

	if opts.stream != "" && entry.OriginalStream != opts.stream {
		return false
	}

	if opts.typ != "" && entry.OriginalEventType != opts.typ {
		return false
	}

	return true
}

func replayEntry(ctx context.Context, client *redis.Client, msgID string, entry *dlqPayload) error {
	metadata := map[string]any{
		"id":        entry.OriginalEventID,
		"type":      entry.OriginalEventType,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0",
		"attempt":   1,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = client.XAdd(ctx, &redis.XAddArgs{
		Stream: entry.OriginalStream,
		Values: map[string]any{
			"metadata": string(metadataJSON),
			"payload":  entry.OriginalPayload,
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("XADD to %s: %w", entry.OriginalStream, err)
	}

	_, err = client.XDel(ctx, streams.StreamDLQ, msgID).Result()
	if err != nil {
		return fmt.Errorf("XDEL %s from DLQ: %w", msgID, err)
	}

	return nil
}

// --- helpers ---

func parseDLQMessage(msg redis.XMessage) (*dlqPayload, error) {
	payloadStr, ok := msg.Values["payload"].(string)
	if !ok {
		return nil, fmt.Errorf("missing payload field in message %s", msg.ID)
	}

	var entry dlqPayload
	if err := json.Unmarshal([]byte(payloadStr), &entry); err != nil {
		return nil, fmt.Errorf("unmarshal payload in message %s: %w", msg.ID, err)
	}

	return &entry, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}
