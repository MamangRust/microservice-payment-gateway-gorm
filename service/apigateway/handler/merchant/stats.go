package merchanthandler

import (
	"fmt"
	"net/http"
	"strconv"

	pbMerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbMerchantStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	stats_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	merchantapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/merchant"
	"github.com/labstack/echo/v4"
)

type merchantStatsHandleApi struct {
	logger logger.LoggerInterface
	cache  stats_cache.StatsCache

	apiHandler errors.ApiHandler

	// Merchant Clients
	merchantTransaction pbMerchant.MerchantTransactionServiceClient
	merchantAmount      pbMerchantStats.MerchantStatsAmountServiceClient
	merchantTotal       pbMerchantStats.MerchantStatsTotalAmountServiceClient
	merchantMethod      pbMerchantStats.MerchantStatsMethodServiceClient
}

func setupMerchantStatsHandler(deps *DepsMerchant) func() {
	return func() {
		statsCache := stats_cache.NewStatsCache(deps.Cache)
		h := &merchantStatsHandleApi{
			logger:     deps.Logger,
			cache:      statsCache,
			apiHandler: deps.ApiHandler,

			merchantTransaction: pbMerchant.NewMerchantTransactionServiceClient(deps.StatsClient),
			merchantAmount:      pbMerchantStats.NewMerchantStatsAmountServiceClient(deps.StatsClient),
			merchantTotal:       pbMerchantStats.NewMerchantStatsTotalAmountServiceClient(deps.StatsClient),
			merchantMethod:      pbMerchantStats.NewMerchantStatsMethodServiceClient(deps.StatsClient),
		}

		g := deps.E.Group("/api/merchant/stats")

		// Transaction Handlers
		g.GET("/transactions", h.apiHandler.Handle("find-all-transaction-merchant", h.FindAllTransactionMerchant))
		g.GET("/transactions/apikey", h.apiHandler.Handle("find-all-transaction-by-apikey", h.FindAllTransactionByApikey))
		g.GET("/transactions/id/:id", h.apiHandler.Handle("find-all-transaction-by-merchant", h.FindAllTransactionByMerchant))

		// Amount Handlers
		g.GET("/amount/monthly", h.apiHandler.Handle("find-monthly-merchant-amount", h.FindMonthlyMerchantAmount))
		g.GET("/amount/yearly", h.apiHandler.Handle("find-yearly-merchant-amount", h.FindYearlyMerchantAmount))
		g.GET("/amount/monthly/id/:id", h.apiHandler.Handle("find-monthly-amount-by-merchants", h.FindMonthlyAmountByMerchants))
		g.GET("/amount/yearly/id/:id", h.apiHandler.Handle("find-yearly-amount-by-merchants", h.FindYearlyAmountByMerchants))
		g.GET("/amount/monthly/apikey", h.apiHandler.Handle("find-monthly-amount-by-apikey", h.FindMonthlyAmountByApikey))
		g.GET("/amount/yearly/apikey", h.apiHandler.Handle("find-yearly-amount-by-apikey", h.FindYearlyAmountByApikey))

		// Total Amount Handlers
		g.GET("/total-amount/monthly", h.apiHandler.Handle("find-monthly-merchant-total-amount", h.FindMonthlyTotalAmountMerchant))
		g.GET("/total-amount/yearly", h.apiHandler.Handle("find-yearly-merchant-total-amount", h.FindYearlyTotalAmountMerchant))
		g.GET("/total-amount/monthly/id/:id", h.apiHandler.Handle("find-monthly-total-amount-by-merchants", h.FindMonthlyTotalAmountByMerchants))
		g.GET("/total-amount/yearly/id/:id", h.apiHandler.Handle("find-yearly-total-amount-by-merchants", h.FindYearlyTotalAmountByMerchants))
		g.GET("/total-amount/monthly/apikey", h.apiHandler.Handle("find-monthly-total-amount-by-apikey", h.FindMonthlyTotalAmountByApikey))
		g.GET("/total-amount/yearly/apikey", h.apiHandler.Handle("find-yearly-total-amount-by-apikey", h.FindYearlyTotalAmountByApikey))

		// Method Handlers
		g.GET("/method/monthly", h.apiHandler.Handle("find-monthly-merchant-method", h.FindMonthlyMerchantMethod))
		g.GET("/method/yearly", h.apiHandler.Handle("find-yearly-merchant-method", h.FindYearlyMerchantMethod))
		g.GET("/method/monthly/id/:id", h.apiHandler.Handle("find-monthly-payment-method-by-merchants", h.FindMonthlyMethodByMerchants))
		g.GET("/method/yearly/id/:id", h.apiHandler.Handle("find-yearly-payment-method-by-merchants", h.FindYearlyMethodByMerchants))
		g.GET("/method/monthly/apikey", h.apiHandler.Handle("find-monthly-payment-method-by-apikey", h.FindMonthlyMethodByApikey))
		g.GET("/method/yearly/apikey", h.apiHandler.Handle("find-yearly-payment-method-by-apikey", h.FindYearlyMethodByApikey))
	}
}

// FindAllTransactionMerchant godoc
// @Summary Get all merchant transactions
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponsesMerchant
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/transactions [get]
func (h *merchantStatsHandleApi) FindAllTransactionMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	cacheKey := "stats:merchant:transactions:all"

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTransaction.FindAllTransactionMerchant(ctx, &pbMerchant.FindAllMerchantRequest{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantTransactionResponseMapper()
	apiRes := mapper.ToApiResponseMerchantsTransactionResponse(res)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindAllTransactionByMerchant godoc
// @Summary Get all transactions for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} response.ApiResponsesMerchant
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/transactions/id/{id} [get]
func (h *merchantStatsHandleApi) FindAllTransactionByMerchant(c echo.Context) error {
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:transactions:id:%d", merchantID)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTransaction.FindAllTransactionByMerchant(ctx, &pbMerchant.FindAllMerchantTransaction{MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantTransactionResponseMapper()
	apiRes := mapper.ToApiResponseMerchantsTransactionResponse(res)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindAllTransactionByApikey godoc
// @Summary Get all transactions for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Success 200 {object} response.ApiResponsesMerchant
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/transactions/apikey [get]
func (h *merchantStatsHandleApi) FindAllTransactionByApikey(c echo.Context) error {
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:transactions:apikey:%s", apiKey)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTransaction.FindAllTransactionByApikey(ctx, &pbMerchant.FindAllMerchantApikey{ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantTransactionResponseMapper()
	apiRes := mapper.ToApiResponseMerchantsTransactionResponse(res)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyMerchantAmount godoc
// @Summary Get monthly merchant amounts
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/monthly [get]
func (h *merchantStatsHandleApi) FindMonthlyMerchantAmount(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindMonthlyAmountMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyMerchantAmount godoc
// @Summary Get yearly merchant amounts
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/yearly [get]
func (h *merchantStatsHandleApi) FindYearlyMerchantAmount(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindYearlyAmountMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyAmountByMerchants godoc
// @Summary Get monthly amounts for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/monthly/id/{id} [get]
func (h *merchantStatsHandleApi) FindMonthlyAmountByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:monthly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindMonthlyAmountByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyAmountByMerchants godoc
// @Summary Get yearly amounts for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/yearly/id/{id} [get]
func (h *merchantStatsHandleApi) FindYearlyAmountByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:yearly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindYearlyAmountByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyAmountByApikey godoc
// @Summary Get monthly amounts for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/monthly/apikey [get]
func (h *merchantStatsHandleApi) FindMonthlyAmountByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:monthly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindMonthlyAmountByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyAmountByApikey godoc
// @Summary Get yearly amounts for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/amount/yearly/apikey [get]
func (h *merchantStatsHandleApi) FindYearlyAmountByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:yearly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantAmount.FindYearlyAmountByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTotalAmountMerchant godoc
// @Summary Get yearly total amounts across all merchants
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/yearly [get]
func (h *merchantStatsHandleApi) FindYearlyTotalAmountMerchant(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindYearlyTotalAmountMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTotalAmountMerchant godoc
// @Summary Get monthly total amounts across all merchants
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/monthly [get]
func (h *merchantStatsHandleApi) FindMonthlyTotalAmountMerchant(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindMonthlyTotalAmountMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTotalAmountByMerchants godoc
// @Summary Get monthly total amounts for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/monthly/id/{id} [get]
func (h *merchantStatsHandleApi) FindMonthlyTotalAmountByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:monthly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindMonthlyTotalAmountByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTotalAmountByMerchants godoc
// @Summary Get yearly total amounts for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/yearly/id/{id} [get]
func (h *merchantStatsHandleApi) FindYearlyTotalAmountByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:yearly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindYearlyTotalAmountByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTotalAmountByApikey godoc
// @Summary Get monthly total amounts for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/monthly/apikey [get]
func (h *merchantStatsHandleApi) FindMonthlyTotalAmountByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:monthly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindMonthlyTotalAmountByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTotalAmountByApikey godoc
// @Summary Get yearly total amounts for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyTotalAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/total-amount/yearly/apikey [get]
func (h *merchantStatsHandleApi) FindYearlyTotalAmountByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:amount:total:yearly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantTotal.FindYearlyTotalAmountByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsTotalAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyTotalAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyMerchantMethod godoc
// @Summary Get monthly merchant payment methods
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/monthly [get]
func (h *merchantStatsHandleApi) FindMonthlyMerchantMethod(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindMonthlyPaymentMethodsMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyMerchantMethod godoc
// @Summary Get yearly merchant payment methods
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/yearly [get]
func (h *merchantStatsHandleApi) FindYearlyMerchantMethod(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindYearlyPaymentMethodMerchant(ctx, &pbMerchant.FindYearMerchant{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseYearlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyMethodByMerchants godoc
// @Summary Get monthly payment methods for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/monthly/id/{id} [get]
func (h *merchantStatsHandleApi) FindMonthlyMethodByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:monthly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindMonthlyPaymentMethodByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyMethodByMerchants godoc
// @Summary Get yearly payment methods for a specific merchant
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/yearly/id/{id} [get]
func (h *merchantStatsHandleApi) FindYearlyMethodByMerchants(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	merchantID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:yearly:id:%d:%d", merchantID, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindYearlyPaymentMethodByMerchants(ctx, &pbMerchant.FindYearMerchantById{Year: int32(year), MerchantId: int32(merchantID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseYearlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyMethodByApikey godoc
// @Summary Get monthly payment methods for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantMonthlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/monthly/apikey [get]
func (h *merchantStatsHandleApi) FindMonthlyMethodByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:monthly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindMonthlyPaymentMethodByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyMethodByApikey godoc
// @Summary Get yearly payment methods for a specific API Key
// @Tags Stats Merchant
// @Accept json
// @Produce json
// @Param apikey query string true "API Key"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMerchantYearlyPaymentMethod
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/merchant/stats/method/yearly/apikey [get]
func (h *merchantStatsHandleApi) FindYearlyMethodByApikey(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	apiKey := c.QueryParam("apikey")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:merchant:method:yearly:apikey:%s:%d", apiKey, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.merchantMethod.FindYearlyPaymentMethodByApikey(ctx, &pbMerchant.FindYearMerchantByApikey{Year: int32(year), ApiKey: apiKey})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := merchantapimapper.NewMerchantStatsMethodResponseMapper()
	apiRes := mapper.ToApiResponseYearlyPaymentMethods(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}
