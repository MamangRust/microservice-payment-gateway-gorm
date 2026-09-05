package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/service"
)

type Handler interface {
	TopupQueryHandleGrpc
	TopupCommandHandleGrpc
}

type handler struct {
	TopupQueryHandleGrpc
	TopupCommandHandleGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		TopupQueryHandleGrpc:   NewTopupQueryHandleGrpc(service),
		TopupCommandHandleGrpc: NewTopupCommandHandleGrpc(service),
	}
}
