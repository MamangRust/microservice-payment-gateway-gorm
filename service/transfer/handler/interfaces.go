package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
)

type TransferQueryHandleGrpc interface {
	pb.TransferQueryServiceServer
}

type TransferCommandHandleGrpc interface {
	pb.TransferCommandServiceServer
}
