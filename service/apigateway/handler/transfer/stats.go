package transferhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	pbTransferStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	stats_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	transferapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/transfer"
	"github.com/labstack/echo/v4"
)

type transferStatsHandleApi struct {
	logger logger.LoggerInterface
	cache  stats_cache.StatsCache

	apiHandler errors.ApiHandler

	// Transfer Clients
	transferAmount pbTransferStats.TransferStatsAmountServiceClient
	transferStatus pbTransferStats.TransferStatsStatusServiceClient
}

func setupTransferStatsHandler(deps *DepsTransfer) func() {
	return func() {
		statsCache := stats_cache.NewStatsCache(deps.Cache)
		h := &transferStatsHandleApi{
			logger:     deps.Logger,
			cache:      statsCache,
			apiHandler: deps.ApiHandler,

			transferAmount: pbTransferStats.NewTransferStatsAmountServiceClient(deps.StatsClient),
			transferStatus: pbTransferStats.NewTransferStatsStatusServiceClient(deps.StatsClient),
		}

		g := deps.E.Group("/api/transfer/stats")

		// Amount Handlers
		g.GET("/amount/monthly", h.apiHandler.Handle("find-monthly-transfer-amounts", h.FindMonthlyTransferAmounts))
		g.GET("/amount/yearly", h.apiHandler.Handle("find-yearly-transfer-amounts", h.FindYearlyTransferAmounts))
		g.GET("/amount/monthly/sender/:card_number", h.apiHandler.Handle("find-monthly-transfer-amounts-by-sender-card-number", h.FindMonthlyTransferAmountsBySenderCardNumber))
		g.GET("/amount/monthly/receiver/:card_number", h.apiHandler.Handle("find-monthly-transfer-amounts-by-receiver-card-number", h.FindMonthlyTransferAmountsByReceiverCardNumber))
		g.GET("/amount/yearly/sender/:card_number", h.apiHandler.Handle("find-yearly-transfer-amounts-by-sender-card-number", h.FindYearlyTransferAmountsBySenderCardNumber))
		g.GET("/amount/yearly/receiver/:card_number", h.apiHandler.Handle("find-yearly-transfer-amounts-by-receiver-card-number", h.FindYearlyTransferAmountsByReceiverCardNumber))

		// Status Handlers
		g.GET("/status/monthly/success", h.apiHandler.Handle("find-monthly-transfer-status-success", h.FindMonthlyTransferStatusSuccess))
		g.GET("/status/yearly/success", h.apiHandler.Handle("find-yearly-transfer-status-success", h.FindYearlyTransferStatusSuccess))
		g.GET("/status/monthly/failed", h.apiHandler.Handle("find-monthly-transfer-status-failed", h.FindMonthlyTransferStatusFailed))
		g.GET("/status/yearly/failed", h.apiHandler.Handle("find-yearly-transfer-status-failed", h.FindYearlyTransferStatusFailed))
		g.GET("/status/monthly/success/:card_number", h.apiHandler.Handle("find-monthly-transfer-status-success-by-card-number", h.FindMonthlyTransferStatusSuccessByCardNumber))
		g.GET("/status/yearly/success/:card_number", h.apiHandler.Handle("find-yearly-transfer-status-success-by-card-number", h.FindYearlyTransferStatusSuccessByCardNumber))
		g.GET("/status/monthly/failed/:card_number", h.apiHandler.Handle("find-monthly-transfer-status-failed-by-card-number", h.FindMonthlyTransferStatusFailedByCardNumber))
		g.GET("/status/yearly/failed/:card_number", h.apiHandler.Handle("find-yearly-transfer-status-failed-by-card-number", h.FindYearlyTransferStatusFailedByCardNumber))
	}
}

// FindMonthlyTransferAmounts godoc
// @Summary Get monthly transfer amounts
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferMonthAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/monthly [get]
func (h *transferStatsHandleApi) FindMonthlyTransferAmounts(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindMonthlyTransferAmounts(ctx, &transfer.FindYearTransferStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferAmounts godoc
// @Summary Get yearly transfer amounts
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/yearly [get]
func (h *transferStatsHandleApi) FindYearlyTransferAmounts(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindYearlyTransferAmounts(ctx, &transfer.FindYearTransferStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferAmountsBySenderCardNumber godoc
// @Summary Get monthly transfer amounts by sender card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferMonthAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/monthly/sender/{card_number} [get]
func (h *transferStatsHandleApi) FindMonthlyTransferAmountsBySenderCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:monthly:sender:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindMonthlyTransferAmountsBySenderCardNumber(ctx, &transfer.FindByCardNumberTransferRequest{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferAmountsByReceiverCardNumber godoc
// @Summary Get monthly transfer amounts by receiver card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferMonthAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/monthly/receiver/{card_number} [get]
func (h *transferStatsHandleApi) FindMonthlyTransferAmountsByReceiverCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:monthly:receiver:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindMonthlyTransferAmountsByReceiverCardNumber(ctx, &transfer.FindByCardNumberTransferRequest{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferAmountsBySenderCardNumber godoc
// @Summary Get yearly transfer amounts by sender card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/yearly/sender/{card_number} [get]
func (h *transferStatsHandleApi) FindYearlyTransferAmountsBySenderCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:yearly:sender:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindYearlyTransferAmountsBySenderCardNumber(ctx, &transfer.FindByCardNumberTransferRequest{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferAmountsByReceiverCardNumber godoc
// @Summary Get yearly transfer amounts by receiver card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/amount/yearly/receiver/{card_number} [get]
func (h *transferStatsHandleApi) FindYearlyTransferAmountsByReceiverCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:amount:yearly:receiver:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferAmount.FindYearlyTransferAmountsByReceiverCardNumber(ctx, &transfer.FindByCardNumberTransferRequest{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearAmount(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferStatusSuccess godoc
// @Summary Get monthly transfer status success stats
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {object} response.ApiResponseTransferMonthStatusSuccess
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/monthly/success [get]
func (h *transferStatsHandleApi) FindMonthlyTransferStatusSuccess(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:monthly:success:%d:%d", year, month)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindMonthlyTransferStatusSuccess(ctx, &transfer.FindMonthlyTransferStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthStatusSuccess(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferStatusSuccess godoc
// @Summary Get yearly transfer status success stats
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearStatusSuccess
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/yearly/success [get]
func (h *transferStatsHandleApi) FindYearlyTransferStatusSuccess(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:yearly:success:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindYearlyTransferStatusSuccess(ctx, &transfer.FindYearTransferStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearStatusSuccess(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferStatusFailed godoc
// @Summary Get monthly transfer status failed stats
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {object} response.ApiResponseTransferMonthStatusFailed
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/monthly/failed [get]
func (h *transferStatsHandleApi) FindMonthlyTransferStatusFailed(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:monthly:failed:%d:%d", year, month)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindMonthlyTransferStatusFailed(ctx, &transfer.FindMonthlyTransferStatus{Year: int32(year), Month: int32(month)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthStatusFailed(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferStatusFailed godoc
// @Summary Get yearly transfer status failed stats
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearStatusFailed
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/yearly/failed [get]
func (h *transferStatsHandleApi) FindYearlyTransferStatusFailed(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:yearly:failed:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindYearlyTransferStatusFailed(ctx, &transfer.FindYearTransferStatus{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearStatusFailed(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferStatusSuccessByCardNumber godoc
// @Summary Get monthly transfer status success stats by card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {object} response.ApiResponseTransferMonthStatusSuccess
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/monthly/success/{card_number} [get]
func (h *transferStatsHandleApi) FindMonthlyTransferStatusSuccessByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:monthly:success:%s:%d:%d", cardNumber, year, month)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindMonthlyTransferStatusSuccessByCardNumber(ctx, &transfer.FindMonthlyTransferStatusCardNumber{Year: int32(year), Month: int32(month), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthStatusSuccess(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferStatusSuccessByCardNumber godoc
// @Summary Get yearly transfer status success stats by card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearStatusSuccess
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/yearly/success/{card_number} [get]
func (h *transferStatsHandleApi) FindYearlyTransferStatusSuccessByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:yearly:success:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindYearlyTransferStatusSuccessByCardNumber(ctx, &transfer.FindYearTransferStatusCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearStatusSuccess(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTransferStatusFailedByCardNumber godoc
// @Summary Get monthly transfer status failed stats by card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {object} response.ApiResponseTransferMonthStatusFailed
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/monthly/failed/{card_number} [get]
func (h *transferStatsHandleApi) FindMonthlyTransferStatusFailedByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	month, _ := strconv.Atoi(c.QueryParam("month"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:monthly:failed:%s:%d:%d", cardNumber, year, month)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindMonthlyTransferStatusFailedByCardNumber(ctx, &transfer.FindMonthlyTransferStatusCardNumber{Year: int32(year), Month: int32(month), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferMonthStatusFailed(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyTransferStatusFailedByCardNumber godoc
// @Summary Get yearly transfer status failed stats by card number
// @Tags Stats Transfer
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseTransferYearStatusFailed
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/transfer/stats/status/yearly/failed/{card_number} [get]
func (h *transferStatsHandleApi) FindYearlyTransferStatusFailedByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:transfer:status:yearly:failed:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.transferStatus.FindYearlyTransferStatusFailedByCardNumber(ctx, &transfer.FindYearTransferStatusCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := transferapimapper.NewTransferStatsStatusResponseMapper()
	apiRes := mapper.ToApiResponseTransferYearStatusFailed(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}
