package outbox

import "context"

// Store is the shared outbox interface used by services to enqueue events.
// R is the service's own outbox record type (each owning service has an
// identical outbox_events table); the publisher reads from the same table
// and publishes to Kafka.
type Store[R any] interface {
	// Insert enqueues a single event into the outbox. The event will be
	// picked up by the Publisher worker and sent to Kafka.
	Insert(ctx context.Context, record R) error
}

// NewStore creates a Store backed by an insert function bound to the service's
// own GORM-based outbox store (see service/*/repository/outboxGormStore.go).
func NewStore[R any](insert func(context.Context, R) error) Store[R] {
	return &store[R]{insert: insert}
}

type store[R any] struct {
	insert func(context.Context, R) error
}

func (s *store[R]) Insert(ctx context.Context, record R) error {
	return s.insert(ctx, record)
}
