package saldogrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcSaldoInvalidID         = errors.NewGrpcError("Invalid Saldo ID", int(codes.InvalidArgument))
	ErrGrpcSaldoInvalidCardNumber = errors.NewGrpcError("Invalid Saldo Card Number", int(codes.InvalidArgument))
	ErrGrpcSaldoInvalidMonth      = errors.NewGrpcError("Invalid Saldo Month", int(codes.InvalidArgument))
	ErrGrpcSaldoInvalidYear       = errors.NewGrpcError("Invalid Saldo Year", int(codes.InvalidArgument))
)
