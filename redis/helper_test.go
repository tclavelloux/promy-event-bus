//nolint:all // Test file
package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// defaultTestRedisDSN matches this repo's own docker-compose.yaml, which maps
// host port 6389 to the redis container's 6379 — deliberately not the Redis
// default port, so it doesn't collide with an unrelated Redis instance a
// developer may already have running on 6379 for another project.
const defaultTestRedisDSN = "redis://localhost:6389/15"

// testDSN returns the Redis DSN to use for this test run: REDIS_TEST_DSN if
// set (CI sets it to point at its own service container), otherwise
// defaultTestRedisDSN. It skips the test if nothing is reachable at that
// address, and registers a cleanup that flushes only the DB it used.
func testDSN(t *testing.T) string {
	t.Helper()

	dsn := defaultTestRedisDSN
	if v := os.Getenv("REDIS_TEST_DSN"); v != "" {
		dsn = v
	}

	opts, err := goredis.ParseURL(dsn)
	if err != nil {
		t.Fatalf("invalid REDIS_TEST_DSN/default %q: %v", dsn, err)
	}

	client := goredis.NewClient(opts)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("no Redis reachable at %s (set REDIS_TEST_DSN to override): %v", dsn, err)
	}

	t.Cleanup(func() {
		cleanupClient := goredis.NewClient(opts)
		defer cleanupClient.Close()

		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()

		_ = cleanupClient.FlushDB(cleanupCtx).Err()
	})

	return dsn
}
