package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
)

type UserQueryHandleGrpc interface {
	pb.UserQueryServiceServer
}

type UserCommandHandleGrpc interface {
	pb.UserCommandServiceServer
}
