package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
)

type WithdrawQueryHandlerGrpc interface {
	pb.WithdrawQueryServiceServer
}

type WithdrawCommandHandlerGrpc interface {
	pb.WithdrawCommandServiceServer
}
