package usergrpcerrors

import (
	"net/http"

	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

// ErrGrpcUserInvalidId is returned when an invalid user ID is provided.
var ErrGrpcUserInvalidId = errors.NewGrpcError("Invalid user ID", http.StatusBadRequest)
