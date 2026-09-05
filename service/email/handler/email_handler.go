package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	traceunic "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/trace_unic"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/email/mailer"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/email/metrics"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotent_consumer"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EventEnvelope mirrors the outbox envelope for dedup purposes.
type eventEnvelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

// emailHandler implements sarama.ConsumerGroupHandler with Phase 3.1 dedup.
type emailHandler struct {
	ctx             context.Context
	trace           trace.Tracer
	logger          logger.LoggerInterface
	Mailer          mailer.MailerInterface
	dedup           *idempotent_consumer.Dedup
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewEmailHandler(ctx context.Context, logger logger.LoggerInterface, mailer mailer.MailerInterface) *emailHandler {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "email_service_requests_total",
			Help: "Total number of requests to the EmailService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_service_request_duration_seconds",
			Help:    "Histogram of request durations for the EmailService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &emailHandler{
		ctx:             ctx,
		logger:          logger,
		Mailer:          mailer,
		dedup:           idempotent_consumer.New(24 * time.Hour),
		trace:           otel.Tracer("email-handler"),
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

func (h *emailHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *emailHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *emailHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	start := time.Now()
	status := "success"
	skipped := 0

	defer func() {
		h.recordMetrics("ConsumeClaim", status, start)
	}()

	_, span := h.trace.Start(h.ctx, "ConsumeClaim")
	defer span.End()

	for msg := range claim.Messages() {
		// Try envelope (Phase 3 format) first; if it carries an event_id
		// that we have already processed, skip this message entirely.
		var inner map[string]interface{}
		if env := h.tryUnwrap(msg.Value); env != nil {
			if h.dedup.IsDuplicate(env.EventID) {
				skipped++
				sess.MarkMessage(msg, "")
				continue
			}
			// Unmarshal the inner payload from the envelope.
			if err := json.Unmarshal(env.Payload, &inner); err != nil {
				traceID := traceunic.GenerateTraceID("FAILED_UNMARSHAL_ENVELOPE_PAYLOAD")
				h.logger.Error("Failed to unmarshal envelope payload", zap.Error(err))
				span.SetAttributes(attribute.String("trace.id", traceID))
				span.RecordError(err)
				span.SetStatus(codes.Error, "Failed to unmarshal envelope payload")
				status = "failed_unmarshal_envelope_payload"
				continue
			}
		} else {
			// Fallback: legacy raw payload (pre-Phase 3 services).
			if err := json.Unmarshal(msg.Value, &inner); err != nil {
				traceID := traceunic.GenerateTraceID("FAILED_UNMARSHAL_MESSAGE")
				h.logger.Error("Failed to unmarshal message", zap.Error(err))
				span.SetAttributes(attribute.String("trace.id", traceID))
				span.RecordError(err)
				span.SetStatus(codes.Error, "Failed to unmarshal message")
				status = "failed_unmarshal_message"
				continue
			}
		}

		email, _ := inner["email"].(string)
		subject, _ := inner["subject"].(string)
		body, _ := inner["body"].(string)

		err := h.Mailer.Send(email, subject, body)
		if err != nil {
			traceID := traceunic.GenerateTraceID("FAILED_SEND_EMAIL")
			h.logger.Error("Failed to send email", zap.Error(err))
			span.SetAttributes(attribute.String("trace.id", traceID))
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to send email")
			status = "failed_send_email"
			metrics.EmailFailed.Inc()
		} else {
			metrics.EmailSent.Inc()
		}

		sess.MarkMessage(msg, "")
	}

	h.logger.Info("ConsumeClaim finished",
		zap.Int("messages", len(claim.Messages())),
		zap.Int("skipped_duplicates", skipped),
		zap.Int("dedup_size", h.dedup.Size()),
	)

	return nil
}

// tryUnwrap attempts to parse a Phase 3 outbox envelope. Returns nil if
// the message is in an older format.
func (h *emailHandler) tryUnwrap(raw []byte) *eventEnvelope {
	var env eventEnvelope
	if err := json.Unmarshal(raw, &env); err != nil || env.EventID == "" {
		return nil
	}
	return &env
}

func (s *emailHandler) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method, status).Observe(time.Since(start).Seconds())
}
