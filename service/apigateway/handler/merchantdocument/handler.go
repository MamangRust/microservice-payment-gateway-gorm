package merchantdocumenthandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant_document"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	merchantdocument_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/merchantdocument"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	merchantdocumentapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/merchantdocument"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsMerchantDocument struct {
	Client *grpc.ClientConn

	E *echo.Echo

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler errors.ApiHandler
}

func RegisterMerchantDocumentHandler(deps *DepsMerchantDocument) {
	mapper := merchantdocumentapimapper.NewMerchantDocumentResponseMapper()
	cache := merchantdocument_cache.NewMerchantDocumentMencache(deps.Cache)

	handlers := []func(){
		setupMerchantDocumentQueryHandler(deps, mapper.QueryMapper(), cache),
		setupMerchantDocumentCommandHandler(deps, mapper.CommandMapper(), cache),
	}

	for _, h := range handlers {
		h()
	}
}

func setupMerchantDocumentQueryHandler(deps *DepsMerchantDocument, mapper merchantdocumentapimapper.MerchantDocumentQueryResponseMapper, cache merchantdocument_cache.MerchantDocumentMencache) func() {
	return func() {
		NewMerchantQueryDocumentHandler(&merchantDocumentQueryDocumentHandleDeps{
			client:     pb.NewMerchantDocumentQueryServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupMerchantDocumentCommandHandler(deps *DepsMerchantDocument, mapper merchantdocumentapimapper.MerchantDocumentCommandResponseMapper, cache merchantdocument_cache.MerchantDocumentMencache) func() {
	return func() {
		NewMerchantCommandDocumentHandler(&merchantCommandDocumentHandleDeps{
			client:     pb.NewMerchantDocumentCommandServiceClient(deps.Client),
			router:     deps.E,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}
