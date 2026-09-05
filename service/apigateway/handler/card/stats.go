package cardhandler

import (
	"fmt"
	"net/http"
	"strconv"

	pbCard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbCardStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	stats_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	cardapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/card"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type cardStatsHandleApi struct {
	logger logger.LoggerInterface
	cache  stats_cache.StatsCache

	apiHandler errors.ApiHandler

	// Card Clients
	cardDashboard pbCard.CardDashboardServiceClient
	cardBalance   pbCardStats.CardStatsBalanceServiceClient
	cardTopup     pbCardStats.CardStatsTopupServiceClient
	cardTrans     pbCardStats.CardStatsTransactionServiceClient
	cardTransfer  pbCardStats.CardStatsTransferServiceClient
	cardWithdraw  pbCardStats.CardStatsWithdrawServiceClient
}

type cardStatsHandleApiDeps struct {
	client      pbCard.CardDashboardServiceClient          // Fallback/Main card client if needed, but we use specific ones below
	statsClient *pbCardStats.CardStatsBalanceServiceClient // Example, but we'll use deps.StatsClient directly in constructor
}

func setupCardStatsHandler(deps *DepsCard) func() {
	return func() {
		statsCache := stats_cache.NewStatsCache(deps.Cache)
		h := &cardStatsHandleApi{
			logger:     deps.Logger,
			cache:      statsCache,
			apiHandler: deps.ApiHandler,

			cardDashboard: pbCard.NewCardDashboardServiceClient(deps.StatsClient),
			cardBalance:   pbCardStats.NewCardStatsBalanceServiceClient(deps.StatsClient),
			cardTopup:     pbCardStats.NewCardStatsTopupServiceClient(deps.StatsClient),
			cardTrans:     pbCardStats.NewCardStatsTransactionServiceClient(deps.StatsClient),
			cardTransfer:  pbCardStats.NewCardStatsTransferServiceClient(deps.StatsClient),
			cardWithdraw:  pbCardStats.NewCardStatsWithdrawServiceClient(deps.StatsClient),
		}

		g := deps.E.Group("/api/card/stats")

		g.GET("/dashboard", h.apiHandler.Handle("dashboard-card", h.DashboardCard))
		g.GET("/dashboard/:card_number", h.apiHandler.Handle("dashboard-card-number", h.DashboardCardNumber))

		g.GET("/balance/monthly", h.apiHandler.Handle("find-monthly-card-balance", h.FindMonthlyCardBalance))
		g.GET("/balance/yearly", h.apiHandler.Handle("find-yearly-card-balance", h.FindYearlyCardBalance))
		g.GET("/balance/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-balance-by-card-number", h.FindMonthlyCardBalanceByCardNumber))
		g.GET("/balance/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-balance-by-card-number", h.FindYearlyCardBalanceByCardNumber))

		g.GET("/topup/monthly", h.apiHandler.Handle("find-monthly-card-topup", h.FindMonthlyCardTopup))
		g.GET("/topup/yearly", h.apiHandler.Handle("find-yearly-card-topup", h.FindYearlyCardTopup))
		g.GET("/topup/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-topup-by-card-number", h.FindMonthlyCardTopupByCardNumber))
		g.GET("/topup/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-topup-by-card-number", h.FindYearlyCardTopupByCardNumber))

		g.GET("/transaction/monthly", h.apiHandler.Handle("find-monthly-card-transaction", h.FindMonthlyCardTransaction))
		g.GET("/transaction/yearly", h.apiHandler.Handle("find-yearly-card-transaction", h.FindYearlyCardTransaction))
		g.GET("/transaction/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-transaction-by-card-number", h.FindMonthlyCardTransactionByCardNumber))
		g.GET("/transaction/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-transaction-by-card-number", h.FindYearlyCardTransactionByCardNumber))

		g.GET("/transfer/sender/monthly", h.apiHandler.Handle("find-monthly-card-transfer-sender", h.FindMonthlyCardTransferSender))
		g.GET("/transfer/sender/yearly", h.apiHandler.Handle("find-yearly-card-transfer-sender", h.FindYearlyCardTransferSender))
		g.GET("/transfer/sender/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-transfer-sender-by-card-number", h.FindMonthlyCardTransferSenderByCardNumber))
		g.GET("/transfer/sender/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-transfer-sender-by-card-number", h.FindYearlyCardTransferSenderByCardNumber))

		g.GET("/transfer/receiver/monthly", h.apiHandler.Handle("find-monthly-card-transfer-receiver", h.FindMonthlyCardTransferReceiver))
		g.GET("/transfer/receiver/yearly", h.apiHandler.Handle("find-yearly-card-transfer-receiver", h.FindYearlyCardTransferReceiver))
		g.GET("/transfer/receiver/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-transfer-receiver-by-card-number", h.FindMonthlyCardTransferReceiverByCardNumber))
		g.GET("/transfer/receiver/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-transfer-receiver-by-card-number", h.FindYearlyCardTransferReceiverByCardNumber))

		g.GET("/withdraw/monthly", h.apiHandler.Handle("find-monthly-card-withdraw", h.FindMonthlyCardWithdraw))
		g.GET("/withdraw/yearly", h.apiHandler.Handle("find-yearly-card-withdraw", h.FindYearlyCardWithdraw))
		g.GET("/withdraw/monthly/:card_number", h.apiHandler.Handle("find-monthly-card-withdraw-by-card-number", h.FindMonthlyCardWithdrawByCardNumber))
		g.GET("/withdraw/yearly/:card_number", h.apiHandler.Handle("find-yearly-card-withdraw-by-card-number", h.FindYearlyCardWithdrawByCardNumber))
	}
}

// DashboardCard godoc
// @Summary Get card dashboard stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseDashboardCard
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/dashboard [get]
func (h *cardStatsHandleApi) DashboardCard(c echo.Context) error {
	ctx := c.Request().Context()
	cacheKey := "stats:card:dashboard"

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardDashboard.DashboardCard(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardDashboardResponseMapper()
	apiRes := mapper.ToApiResponseDashboardCard(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// DashboardCardNumber godoc
// @Summary Get card dashboard stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Success 200 {object} response.ApiResponseDashboardCardNumber
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/dashboard/{card_number} [get]
func (h *cardStatsHandleApi) DashboardCardNumber(c echo.Context) error {
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:dashboard:%s", cardNumber)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardDashboard.DashboardCardNumber(ctx, &pbCard.FindByCardNumberRequest{
		CardNumber: cardNumber,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardDashboardResponseMapper()
	apiRes := mapper.ToApiResponseDashboardCardCardNumber(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardBalance godoc
// @Summary Get monthly card balance stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyBalance
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/balance/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardBalance(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:balance:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardBalance.FindMonthlyBalance(ctx, &pbCardStats.FindYearBalance{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyBalances(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardBalance godoc
// @Summary Get yearly card balance stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyBalance
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/balance/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardBalance(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:balance:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardBalance.FindYearlyBalance(ctx, &pbCardStats.FindYearBalance{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseYearlyBalances(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardBalanceByCardNumber godoc
// @Summary Get monthly card balance stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyBalance
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/balance/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardBalanceByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:balance:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardBalance.FindMonthlyBalanceByCardNumber(ctx, &pbCardStats.FindYearBalanceCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyBalances(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardBalanceByCardNumber godoc
// @Summary Get yearly card balance stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyBalance
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/balance/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardBalanceByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:balance:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardBalance.FindYearlyBalanceByCardNumber(ctx, &pbCardStats.FindYearBalanceCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseYearlyBalances(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTopup godoc
// @Summary Get monthly card topup stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/topup/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardTopup(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:topup:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTopup.FindMonthlyTopupAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTopup godoc
// @Summary Get yearly card topup stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/topup/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardTopup(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:topup:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTopup.FindYearlyTopupAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTopupByCardNumber godoc
// @Summary Get monthly card topup stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/topup/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardTopupByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:topup:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTopup.FindMonthlyTopupAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTopupByCardNumber godoc
// @Summary Get yearly card topup stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/topup/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardTopupByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:topup:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTopup.FindYearlyTopupAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransaction godoc
// @Summary Get monthly card transaction stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transaction/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransaction(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transaction:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTrans.FindMonthlyTransactionAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransaction godoc
// @Summary Get yearly card transaction stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transaction/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardTransaction(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transaction:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTrans.FindYearlyTransactionAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransactionByCardNumber godoc
// @Summary Get monthly card transaction stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transaction/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransactionByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transaction:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTrans.FindMonthlyTransactionAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransactionByCardNumber godoc
// @Summary Get yearly card transaction stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transaction/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardTransactionByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transaction:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTrans.FindYearlyTransactionAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransferSender godoc
// @Summary Get monthly card transfer sender stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/sender/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransferSender(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:sender:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindMonthlyTransferSenderAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransferSender godoc
// @Summary Get yearly card transfer sender stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/sender/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardTransferSender(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:sender:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindYearlyTransferSenderAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransferReceiver godoc
// @Summary Get monthly card transfer receiver stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/receiver/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransferReceiver(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:receiver:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindMonthlyTransferReceiverAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransferReceiver godoc
// @Summary Get yearly card transfer receiver stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/receiver/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardTransferReceiver(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:receiver:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindYearlyTransferReceiverAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransferSenderByCardNumber godoc
// @Summary Get monthly card transfer sender stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/sender/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransferSenderByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:sender:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindMonthlyTransferSenderAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransferSenderByCardNumber godoc
// @Summary Get yearly card transfer sender stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/sender/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardTransferSenderByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:sender:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindYearlyTransferSenderAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardTransferReceiverByCardNumber godoc
// @Summary Get monthly card transfer receiver stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/receiver/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardTransferReceiverByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:receiver:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindMonthlyTransferReceiverAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardTransferReceiverByCardNumber godoc
// @Summary Get yearly card transfer receiver stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/transfer/receiver/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardTransferReceiverByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:transfer:receiver:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardTransfer.FindYearlyTransferReceiverAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardWithdraw godoc
// @Summary Get monthly card withdraw stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/withdraw/monthly [get]
func (h *cardStatsHandleApi) FindMonthlyCardWithdraw(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:withdraw:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardWithdraw.FindMonthlyWithdrawAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardWithdraw godoc
// @Summary Get yearly card withdraw stats
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/withdraw/yearly [get]
func (h *cardStatsHandleApi) FindYearlyCardWithdraw(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:withdraw:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardWithdraw.FindYearlyWithdrawAmount(ctx, &pbCard.FindYearAmount{Year: int32(year)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyCardWithdrawByCardNumber godoc
// @Summary Get monthly card withdraw stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/withdraw/monthly/{card_number} [get]
func (h *cardStatsHandleApi) FindMonthlyCardWithdrawByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:withdraw:monthly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardWithdraw.FindMonthlyWithdrawAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseMonthlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlyCardWithdrawByCardNumber godoc
// @Summary Get yearly card withdraw stats by card number
// @Tags Stats Card
// @Accept json
// @Produce json
// @Param card_number path string true "Card Number"
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearlyAmount
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/card/stats/withdraw/yearly/{card_number} [get]
func (h *cardStatsHandleApi) FindYearlyCardWithdrawByCardNumber(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	cardNumber := c.Param("card_number")
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:card:withdraw:yearly:%s:%d", cardNumber, year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.cardWithdraw.FindYearlyWithdrawAmountByCardNumber(ctx, &pbCard.FindYearAmountCardNumber{Year: int32(year), CardNumber: cardNumber})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := cardapimapper.NewCardStatsAmountResponseMapper()
	apiRes := mapper.ToApiResponseYearlyAmounts(res)
	h.cache.SetCache(ctx, cacheKey, apiRes)

	return c.JSON(http.StatusOK, apiRes)
}
