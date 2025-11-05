package events

// Event type constants define all possible event types in the system.
const (
	// Promotion events
	EventPromotionCreated = "promotion.created"
	EventPromotionUpdated = "promotion.updated"
	EventPromotionDeleted = "promotion.deleted"

	// Product events
	EventProductIdentified = "product.identified"

	// User events
	EventUserRegistered         = "user.registered"
	EventUserPreferencesUpdated = "user.preferences.updated"
	EventUserLocationUpdated    = "user.location.updated"
)

// Stream name constants define the Redis Stream names for each domain.
const (
	// StreamPromotions is the stream for all promotion-related events.
	StreamPromotions = "events:promotions"

	// StreamProducts is the stream for all product-related events.
	StreamProducts = "events:products"

	// StreamUsers is the stream for all user-related events.
	StreamUsers = "events:users"
)
