package events

import (
	"encoding/json"
	"time"
)

// Envelope is the standardized event wrapper sent through Kafka.
// All Phase 3 events MUST use this envelope so consumers can
// deduplicate by event_id.
type Envelope struct {
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	EventVersion int32           `json:"event_version"`
	AggregateID  string          `json:"aggregate_id"`
	OccurredAt   time.Time       `json:"occurred_at"`
	Payload      json.RawMessage `json:"payload"`
}

const (
	// Financial event types
	EventTopupCreated       = "topup.created"
	EventTransactionCreated = "transaction.created"
	EventTransferCreated    = "transfer.created"
	EventWithdrawCreated    = "withdraw.created"
	EventSaldoChanged       = "saldo.changed"

	// Stats event types (domain-specific)
	EventStatsTransaction = "stats.transaction"
	EventStatsTopup       = "stats.topup"
	EventStatsTransfer    = "stats.transfer"
	EventStatsWithdraw    = "stats.withdraw"
	EventStatsSaldo       = "stats.saldo"
)

// Marshal wraps a domain event payload into the standard Envelope.
func Marshal(eventID, eventType string, version int32, aggregateID string, payload interface{}) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := Envelope{
		EventID:      eventID,
		EventType:    eventType,
		EventVersion: version,
		AggregateID:  aggregateID,
		OccurredAt:   time.Now(),
		Payload:      payloadBytes,
	}
	return json.Marshal(env)
}
