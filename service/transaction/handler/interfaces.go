package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
)

type TransactionQueryHandleGrpc interface {
	pb.TransactionQueryServiceServer
}

type TransactionCommandHandleGrpc interface {
	pb.TransactionCommandServiceServer
}
