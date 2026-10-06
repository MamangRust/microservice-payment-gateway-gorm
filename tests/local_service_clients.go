package tests

import (
	"context"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbmerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	card_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/handler"
	merchant_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/handler"
	saldo_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/handler"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// LocalCardClient implements both pbcard.CardQueryServiceClient and
// pbcard.CardCommandServiceClient by delegating to the in-process card handler.
type LocalCardClient struct {
	Handler card_handler.Handler
}

func (c *LocalCardClient) FindAllCard(ctx context.Context, in *pbcard.FindAllCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponsePaginationCard, error) {
	return c.Handler.FindAllCard(ctx, in)
}
func (c *LocalCardClient) FindByIdCard(ctx context.Context, in *pbcard.FindByIdCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.FindByIdCard(ctx, in)
}
func (c *LocalCardClient) FindByUserIdCard(ctx context.Context, in *pbcard.FindByUserIdCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.FindByUserIdCard(ctx, in)
}
func (c *LocalCardClient) FindByActiveCard(ctx context.Context, in *pbcard.FindAllCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponsePaginationCardDeleteAt, error) {
	return c.Handler.FindByActiveCard(ctx, in)
}
func (c *LocalCardClient) FindByTrashedCard(ctx context.Context, in *pbcard.FindAllCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponsePaginationCardDeleteAt, error) {
	return c.Handler.FindByTrashedCard(ctx, in)
}
func (c *LocalCardClient) FindByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.FindByCardNumber(ctx, in)
}
func (c *LocalCardClient) FindUserCardByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, opts ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	return c.Handler.FindUserCardByCardNumber(ctx, in)
}

func (c *LocalCardClient) CreateCard(ctx context.Context, in *pbcard.CreateCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.CreateCard(ctx, in)
}
func (c *LocalCardClient) UpdateCard(ctx context.Context, in *pbcard.UpdateCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.UpdateCard(ctx, in)
}
func (c *LocalCardClient) TrashedCard(ctx context.Context, in *pbcard.FindByIdCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCardDeleteAt, error) {
	return c.Handler.TrashedCard(ctx, in)
}
func (c *LocalCardClient) RestoreCard(ctx context.Context, in *pbcard.FindByIdCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCardDeleteAt, error) {
	return c.Handler.RestoreCard(ctx, in)
}
func (c *LocalCardClient) DeleteCardPermanent(ctx context.Context, in *pbcard.FindByIdCardRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCardDelete, error) {
	return c.Handler.DeleteCardPermanent(ctx, in)
}
func (c *LocalCardClient) RestoreAllCard(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbcard.ApiResponseCardAll, error) {
	return c.Handler.RestoreAllCard(ctx, in)
}
func (c *LocalCardClient) DeleteAllCardPermanent(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbcard.ApiResponseCardAll, error) {
	return c.Handler.DeleteAllCardPermanent(ctx, in)
}
func (c *LocalCardClient) ToggleCardStatus(ctx context.Context, in *pbcard.ToggleCardStatusRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.ToggleCardStatus(ctx, in)
}
func (c *LocalCardClient) UpdateCreditLimit(ctx context.Context, in *pbcard.UpdateCreditLimitRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.UpdateCreditLimit(ctx, in)
}
func (c *LocalCardClient) RedeemPoints(ctx context.Context, in *pbcard.RedeemPointsRequest, opts ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return c.Handler.RedeemPoints(ctx, in)
}
func (c *LocalCardClient) ProcessBillingCycles(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.Handler.ProcessBillingCycles(ctx, in)
}

// LocalSaldoClient implements both pbsaldo.SaldoQueryServiceClient and
// pbsaldo.SaldoCommandServiceClient by delegating to the in-process saldo handler.
type LocalSaldoClient struct {
	Handler saldo_handler.Handler
}

func (c *LocalSaldoClient) FindAllSaldo(ctx context.Context, in *pbsaldo.FindAllSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldo, error) {
	return c.Handler.FindAllSaldo(ctx, in)
}
func (c *LocalSaldoClient) FindByIdSaldo(ctx context.Context, in *pbsaldo.FindByIdSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.FindByIdSaldo(ctx, in)
}
func (c *LocalSaldoClient) FindByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.FindByCardNumber(ctx, in)
}
func (c *LocalSaldoClient) FindByActive(ctx context.Context, in *pbsaldo.FindAllSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldoDeleteAt, error) {
	return c.Handler.FindByActive(ctx, in)
}
func (c *LocalSaldoClient) FindByTrashed(ctx context.Context, in *pbsaldo.FindAllSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldoDeleteAt, error) {
	return c.Handler.FindByTrashed(ctx, in)
}
func (c *LocalSaldoClient) ListReconciliationQueue(ctx context.Context, in *pbsaldo.ListReconciliationRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseReconciliation, error) {
	return c.Handler.ListReconciliationQueue(ctx, in)
}
func (c *LocalSaldoClient) ListLedgerEntries(ctx context.Context, in *pbsaldo.ListLedgerEntriesRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseLedger, error) {
	return c.Handler.ListLedgerEntries(ctx, in)
}

func (c *LocalSaldoClient) CreateSaldo(ctx context.Context, in *pbsaldo.CreateSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.CreateSaldo(ctx, in)
}
func (c *LocalSaldoClient) UpdateSaldo(ctx context.Context, in *pbsaldo.UpdateSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.UpdateSaldo(ctx, in)
}
func (c *LocalSaldoClient) UpdateSaldoWithdraw(ctx context.Context, in *pbsaldo.UpdateSaldoWithdrawRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.UpdateSaldoWithdraw(ctx, in)
}
func (c *LocalSaldoClient) DebitSaldo(ctx context.Context, in *pbsaldo.DebitSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.DebitSaldo(ctx, in)
}
func (c *LocalSaldoClient) CreditSaldo(ctx context.Context, in *pbsaldo.CreditSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return c.Handler.CreditSaldo(ctx, in)
}
func (c *LocalSaldoClient) ApplySaldoAdjustment(ctx context.Context, in *pbsaldo.ApplySaldoAdjustmentRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseAdjustment, error) {
	return c.Handler.ApplySaldoAdjustment(ctx, in)
}
func (c *LocalSaldoClient) ResolveReconciliation(ctx context.Context, in *pbsaldo.ResolveReconciliationRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoAll, error) {
	return c.Handler.ResolveReconciliation(ctx, in)
}
func (c *LocalSaldoClient) TrashedSaldo(ctx context.Context, in *pbsaldo.FindByIdSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDeleteAt, error) {
	return c.Handler.TrashedSaldo(ctx, in)
}
func (c *LocalSaldoClient) RestoreSaldo(ctx context.Context, in *pbsaldo.FindByIdSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDeleteAt, error) {
	return c.Handler.RestoreSaldo(ctx, in)
}
func (c *LocalSaldoClient) DeleteSaldoPermanent(ctx context.Context, in *pbsaldo.FindByIdSaldoRequest, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDelete, error) {
	return c.Handler.DeleteSaldoPermanent(ctx, in)
}
func (c *LocalSaldoClient) RestoreAllSaldo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoAll, error) {
	return c.Handler.RestoreAllSaldo(ctx, in)
}
func (c *LocalSaldoClient) DeleteAllSaldoPermanent(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoAll, error) {
	return c.Handler.DeleteAllSaldoPermanent(ctx, in)
}

// LocalMerchantClient implements pbmerchant.MerchantQueryServiceClient and
// pbmerchant.MerchantCommandServiceClient by delegating to the in-process
// merchant handler.
type LocalMerchantClient struct {
	Handler merchant_handler.Handler
}

func (c *LocalMerchantClient) FindAllMerchant(ctx context.Context, in *pbmerchant.FindAllMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchant, error) {
	return c.Handler.FindAllMerchant(ctx, in)
}
func (c *LocalMerchantClient) FindByIdMerchant(ctx context.Context, in *pbmerchant.FindByIdMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return c.Handler.FindByIdMerchant(ctx, in)
}
func (c *LocalMerchantClient) FindByApiKey(ctx context.Context, in *pbmerchant.FindByApiKeyRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return c.Handler.FindByApiKey(ctx, in)
}
func (c *LocalMerchantClient) FindByMerchantUserId(ctx context.Context, in *pbmerchant.FindByMerchantUserIdRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponsesMerchant, error) {
	return c.Handler.FindByMerchantUserId(ctx, in)
}
func (c *LocalMerchantClient) FindByActive(ctx context.Context, in *pbmerchant.FindAllMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchantDeleteAt, error) {
	return c.Handler.FindByActive(ctx, in)
}
func (c *LocalMerchantClient) FindByTrashed(ctx context.Context, in *pbmerchant.FindAllMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchantDeleteAt, error) {
	return c.Handler.FindByTrashed(ctx, in)
}

func (c *LocalMerchantClient) CreateMerchant(ctx context.Context, in *pbmerchant.CreateMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return c.Handler.CreateMerchant(ctx, in)
}
func (c *LocalMerchantClient) UpdateMerchant(ctx context.Context, in *pbmerchant.UpdateMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return c.Handler.UpdateMerchant(ctx, in)
}
func (c *LocalMerchantClient) UpdateMerchantStatus(ctx context.Context, in *pbmerchant.UpdateMerchantStatusRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return c.Handler.UpdateMerchantStatus(ctx, in)
}
func (c *LocalMerchantClient) TrashedMerchant(ctx context.Context, in *pbmerchant.FindByIdMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchantDeleteAt, error) {
	return c.Handler.TrashedMerchant(ctx, in)
}
func (c *LocalMerchantClient) RestoreMerchant(ctx context.Context, in *pbmerchant.FindByIdMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchantDeleteAt, error) {
	return c.Handler.RestoreMerchant(ctx, in)
}
func (c *LocalMerchantClient) DeleteMerchantPermanent(ctx context.Context, in *pbmerchant.FindByIdMerchantRequest, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchantDelete, error) {
	return c.Handler.DeleteMerchantPermanent(ctx, in)
}
func (c *LocalMerchantClient) RestoreAllMerchant(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchantAll, error) {
	return c.Handler.RestoreAllMerchant(ctx, in)
}
func (c *LocalMerchantClient) DeleteAllMerchantPermanent(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pbmerchant.ApiResponseMerchantAll, error) {
	return c.Handler.DeleteAllMerchantPermanent(ctx, in)
}
