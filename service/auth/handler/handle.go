package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/auth/service"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Auth AuthHandleGrpc
}

// NewHandler sets up the handler for the authentication service.
//
// It takes a pointer to a Deps struct, which contains all the dependencies required
// to set up the handler.
//
// The returned Handler contains the gRPC handler for the authentication service.
func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Auth: NewAuthHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
