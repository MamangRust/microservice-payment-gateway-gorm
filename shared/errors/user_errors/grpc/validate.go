package usergrpcerrors

import (
	"net/http"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

var (
	// ErrGrpcValidateCreateUser is returned when a create user request validation fails.
	ErrGrpcValidateCreateUser = errors.NewGrpcError("Validation failed: invalid create user request", http.StatusBadRequest)

	// ErrGrpcValidateUpdateUser is returned when an update user request validation fails.
	ErrGrpcValidateUpdateUser = errors.NewGrpcError("Validation failed: invalid update user request", http.StatusBadRequest)
)
