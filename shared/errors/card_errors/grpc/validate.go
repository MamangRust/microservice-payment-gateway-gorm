package cardgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateCardRequest = errors.NewGrpcError("Invalid input for create card", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateCardRequest = errors.NewGrpcError("Invalid input for update card", int(codes.InvalidArgument))
	ErrGrpcValidateToggleCardStatus  = errors.NewGrpcError("Invalid input for toggle card status", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateCreditLimit = errors.NewGrpcError("Invalid input for update credit limit", int(codes.InvalidArgument))
	ErrGrpcValidateRedeemPoints      = errors.NewGrpcError("Invalid input for redeem points", int(codes.InvalidArgument))
)
