package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
)

type TopupQueryHandleGrpc interface {
	pb.TopupQueryServiceServer
}

type TopupCommandHandleGrpc interface {
	pb.TopupCommandServiceServer
}
