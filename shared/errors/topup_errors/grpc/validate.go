package topupgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateTopup = errors.NewGrpcError("Invalid input for create topup", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateTopup = errors.NewGrpcError("Invalid input for update topup", int(codes.InvalidArgument))
)
