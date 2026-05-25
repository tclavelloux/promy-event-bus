package streams

// Stream name constants define the Redis Stream names for each domain.
// These are the only domain-aware constants that promy-event-bus owns.
// Event type constants and payload structs belong in each producing service.
const (
	StreamUsers           = "events:users"
	StreamSubscriptions   = "events:subscriptions"
	StreamPromotions      = "events:promotions"
	StreamProducts        = "events:products"
	StreamIdentifications = "events:identifications"
)

// StreamDLQ is the dead-letter queue stream.
// Unlike all other streams, StreamDLQ has no single owner — any service may
// publish to it when a Tier 1 event handler exhausts its retries.
// Subscribers on StreamDLQ are ops tooling or automated replay workers.
// Never publish business events directly to StreamDLQ; use it only as a failure sink.
const StreamDLQ = "events:dlq"
