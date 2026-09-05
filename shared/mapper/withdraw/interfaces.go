package withdrawapimapper

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	pbstats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/response"
)

type WithdrawBaseResponseMapper interface {
	// Converts a single withdraw response into an API response.
	ToApiResponseWithdraw(pbResponse *pb.ApiResponseWithdraw) *response.ApiResponseWithdraw
}

type WithdrawQueryResponseMapper interface {
	WithdrawBaseResponseMapper

	// Converts a list of withdraw responses into a grouped API response.
	ToApiResponsesWithdraw(pbResponse *pb.ApiResponsesWithdraw) *response.ApiResponsesWithdraw

	// Converts paginated withdraw records into an API response.
	ToApiResponsePaginationWithdraw(pbResponse *pb.ApiResponsePaginationWithdraw) *response.ApiResponsePaginationWithdraw

	// Converts paginated soft-deleted withdraw records into an API response.
	ToApiResponsePaginationWithdrawDeleteAt(pbResponse *pb.ApiResponsePaginationWithdrawDeleteAt) *response.ApiResponsePaginationWithdrawDeleteAt
}

type WithdrawCommandResponseMapper interface {
	WithdrawBaseResponseMapper

	ToApiResponseWithdrawDeleteAt(pbResponse *pb.ApiResponseWithdrawDeleteAt) *response.ApiResponseWithdrawDeleteAt

	// Converts a permanently deleted withdraw response into an API response.
	ToApiResponseWithdrawDelete(pbResponse *pb.ApiResponseWithdrawDelete) *response.ApiResponseWithdrawDelete

	// Converts all withdraw records into an API response.
	ToApiResponseWithdrawAll(pbResponse *pb.ApiResponseWithdrawAll) *response.ApiResponseWithdrawAll
}

type WithdrawStatsStatusResponseMapper interface {
	// Converts monthly successful withdraw statistics into an API response.
	ToApiResponseWithdrawMonthStatusSuccess(pbResponse *pbstats.ApiResponseWithdrawMonthStatusSuccess) *response.ApiResponseWithdrawMonthStatusSuccess

	// Converts yearly successful withdraw statistics into an API response.
	ToApiResponseWithdrawYearStatusSuccess(pbResponse *pbstats.ApiResponseWithdrawYearStatusSuccess) *response.ApiResponseWithdrawYearStatusSuccess

	// Converts monthly failed withdraw statistics into an API response.
	ToApiResponseWithdrawMonthStatusFailed(pbResponse *pbstats.ApiResponseWithdrawMonthStatusFailed) *response.ApiResponseWithdrawMonthStatusFailed

	// Converts yearly failed withdraw statistics into an API response.
	ToApiResponseWithdrawYearStatusFailed(pbResponse *pbstats.ApiResponseWithdrawYearStatusFailed) *response.ApiResponseWithdrawYearStatusFailed
}

type WithdrawStatsAmountResponseMapper interface {
	// Converts monthly total withdraw amount statistics into an API response.
	ToApiResponseWithdrawMonthAmount(pbResponse *pbstats.ApiResponseWithdrawMonthAmount) *response.ApiResponseWithdrawMonthAmount

	// Converts yearly total withdraw amount statistics into an API response.
	ToApiResponseWithdrawYearAmount(pbResponse *pbstats.ApiResponseWithdrawYearAmount) *response.ApiResponseWithdrawYearAmount
}
