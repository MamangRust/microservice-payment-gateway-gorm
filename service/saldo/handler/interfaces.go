package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
)

type SaldoQueryHandleGrpc interface {
	pb.SaldoQueryServiceServer
}

type SaldoCommandHandleGrpc interface {
	pb.SaldoCommandServiceServer
}
