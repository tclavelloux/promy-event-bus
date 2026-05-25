package streams

// Stream name constants define the Redis Stream names for each domain.
// These are the only domain-aware constants that promy-event-bus owns.
// Event type constants and payload structs belong in each producing service.
const (
	StreamPromotions = "events:promotions"
	StreamProducts   = "events:products"
	StreamUsers      = "events:users"
)
