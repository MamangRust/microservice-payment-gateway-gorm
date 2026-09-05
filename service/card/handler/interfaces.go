package handler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
)

type CardQueryService interface {
	pb.CardQueryServiceServer
}

type CardCommandService interface {
	pb.CardCommandServiceServer
}

