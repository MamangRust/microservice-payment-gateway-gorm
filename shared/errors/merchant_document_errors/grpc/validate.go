package merchantdocumentgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateMerchantDocument = errors.NewGrpcError("Invalid input for create merchant document", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateMerchantDocument = errors.NewGrpcError("Invalid input for update merchant document", int(codes.InvalidArgument))
)
