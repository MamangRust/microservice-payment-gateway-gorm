package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"
)

type Handler interface {
	CardQueryService
	CardCommandService
}

type handler struct {
	CardQueryService
	CardCommandService
}

// NewHandler creates a new CardHandler instance.
func NewHandler(service service.Service) Handler {
	return &handler{
		CardQueryService:     NewCardQueryHandleGrpc(service),
		CardCommandService:   NewCardCommandHandleGrpc(service),
	}
}
