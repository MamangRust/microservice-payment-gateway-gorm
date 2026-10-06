package errors

import (
	"errors"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/observability"
	"github.com/labstack/echo/v4"
)

// ErrorResponse is the standard JSON body returned for mapped API errors.
type ErrorResponse struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Type        ErrorType         `json:"type,omitempty"`
	Code        int               `json:"code"`
	TraceID     string            `json:"trace_id,omitempty"`
	Retryable   bool              `json:"retryable,omitempty"`
	Validations []ValidationError `json:"validations,omitempty"`
}

// ApiHandler wraps an echo handler with shared cross-cutting behavior
// (observability, logging, error mapping) for the API gateway routes.
type ApiHandler struct {
	obs    observability.TraceLoggerObservability
	logger logger.LoggerInterface
}

// NewApiHandler constructs an ApiHandler.
func NewApiHandler(obs observability.TraceLoggerObservability, logger logger.LoggerInterface) ApiHandler {
	return ApiHandler{
		obs:    obs,
		logger: logger,
	}
}

// Handle returns an echo handler that delegates to h. The name is kept for
// trace/metric labelling. Any *AppError returned by h is mapped to its HTTP
// status code so structured gRPC errors surface as 400/401/403/404/409/...
// instead of collapsing to Echo's default 500.
func (a ApiHandler) Handle(name string, h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := h(c)
		if err == nil {
			return nil
		}

		var appErr *AppError
		if errors.As(err, &appErr) {
			return c.JSON(appErr.Code, ErrorResponse{
				Status:      "error",
				Message:     appErr.Message,
				Type:        appErr.Type,
				Code:        appErr.Code,
				TraceID:     c.Response().Header().Get(echo.HeaderXRequestID),
				Retryable:   appErr.Retryable,
				Validations: appErr.Validations,
			})
		}

		return err
	}
}
