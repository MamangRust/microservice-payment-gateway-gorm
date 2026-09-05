package handler

import (
	"context"
	"math"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/service"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbutils "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	card_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/card_errors/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type cardQueryHandleGrpc struct {
	pb.UnimplementedCardQueryServiceServer
	cardQuery service.CardQueryService
}

func NewCardQueryHandleGrpc(cardQuery service.CardQueryService) CardQueryService {
	return &cardQueryHandleGrpc{
		cardQuery: cardQuery,
	}
}

func formatCardTime(tp *time.Time) string {
	if tp == nil {
		return ""
	}
	return tp.Format(time.RFC3339)
}

func formatCardTimeTim(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (s *cardQueryHandleGrpc) FindAllCard(ctx context.Context, req *pb.FindAllCardRequest) (*pb.ApiResponsePaginationCard, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCards{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	cards, totalRecords, err := s.cardQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoCards := make([]*pb.CardResponse, len(cards))
	for i, card := range cards {
		protoCards[i] = &pb.CardResponse{
			Id:           int32(card.CardID),
			UserId:       int32(card.UserID),
			CardNumber:   card.CardNumber,
			CardType:     card.CardType,
			CardProvider: card.CardProvider,
			Cvv:          card.Cvv,
			ExpireDate:   formatCardTimeTim(card.ExpireDate),
			CreatedAt:    formatCardTime(card.CreatedAt),
			UpdatedAt:    formatCardTime(card.UpdatedAt),
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationCard{
		Status:         "success",
		Message:        "Successfully fetched card records",
		Data:           protoCards,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *cardQueryHandleGrpc) FindByIdCard(ctx context.Context, req *pb.FindByIdCardRequest) (*pb.ApiResponseCard, error) {
	id := int(req.GetCardId())
	if id == 0 {
		return nil, card_errors.ErrGrpcInvalidCardID
	}

	card, err := s.cardQuery.FindById(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseCard{
		Message: "successfully",
		Status:  "success",
		Data: &pb.CardResponse{
			Id:           int32(card.CardID),
			UserId:       int32(card.UserID),
			CardNumber:   card.CardNumber,
			CardType:     card.CardType,
			CardProvider: card.CardProvider,
			Cvv:          card.Cvv,
			ExpireDate:   formatCardTimeTim(card.ExpireDate),
			CreatedAt:    formatCardTime(card.CreatedAt),
			UpdatedAt:    formatCardTime(card.UpdatedAt),
		},
	}, nil
}

func (s *cardQueryHandleGrpc) FindByActiveCard(ctx context.Context, req *pb.FindAllCardRequest) (*pb.ApiResponsePaginationCardDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCards{Page: page, PageSize: pageSize, Search: search}

	cards, totalRecords, err := s.cardQuery.FindByActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoCards := make([]*pb.CardResponseDeleteAt, len(cards))
	for i, card := range cards {
		var deletedAt *wrapperspb.StringValue
		if card.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: card.DeletedAt.Format(time.RFC3339)}
		}
		protoCards[i] = &pb.CardResponseDeleteAt{
			Id:           int32(card.CardID),
			UserId:       int32(card.UserID),
			CardNumber:   card.CardNumber,
			CardType:     card.CardType,
			CardProvider: card.CardProvider,
			Cvv:          card.Cvv,
			ExpireDate:   formatCardTimeTim(card.ExpireDate),
			CreatedAt:    formatCardTime(card.CreatedAt),
			UpdatedAt:    formatCardTime(card.UpdatedAt),
			DeletedAt:    deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationCardDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched active card records",
		Data:           protoCards,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *cardQueryHandleGrpc) FindByTrashedCard(ctx context.Context, req *pb.FindAllCardRequest) (*pb.ApiResponsePaginationCardDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCards{Page: page, PageSize: pageSize, Search: search}

	cards, totalRecords, err := s.cardQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoCards := make([]*pb.CardResponseDeleteAt, len(cards))
	for i, card := range cards {
		var deletedAt *wrapperspb.StringValue
		if card.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: card.DeletedAt.Format(time.RFC3339)}
		}
		protoCards[i] = &pb.CardResponseDeleteAt{
			Id:           int32(card.CardID),
			UserId:       int32(card.UserID),
			CardNumber:   card.CardNumber,
			CardType:     card.CardType,
			CardProvider: card.CardProvider,
			Cvv:          card.Cvv,
			ExpireDate:   formatCardTimeTim(card.ExpireDate),
			CreatedAt:    formatCardTime(card.CreatedAt),
			UpdatedAt:    formatCardTime(card.UpdatedAt),
			DeletedAt:    deletedAt,
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationCardDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched trashed card records",
		Data:           protoCards,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *cardQueryHandleGrpc) FindByCardNumber(ctx context.Context, req *pb.FindByCardNumberRequest) (*pb.ApiResponseCard, error) {
	card_number := req.GetCardNumber()
	if card_number == "" {
		return nil, card_errors.ErrGrpcInvalidCardNumber
	}

	res, err := s.cardQuery.FindByCardNumber(ctx, card_number)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseCard{
		Status:  "success",
		Message: "Successfully fetched card record",
		Data: &pb.CardResponse{
			Id:           int32(res.CardID),
			UserId:       int32(res.UserID),
			CardNumber:   res.CardNumber,
			CardType:     res.CardType,
			CardProvider: res.CardProvider,
			Cvv:          res.Cvv,
			ExpireDate:   formatCardTimeTim(res.ExpireDate),
			CreatedAt:    formatCardTime(res.CreatedAt),
			UpdatedAt:    formatCardTime(res.UpdatedAt),
		},
	}, nil
}

func (s *cardQueryHandleGrpc) FindByUserIdCard(ctx context.Context, req *pb.FindByUserIdCardRequest) (*pb.ApiResponseCard, error) {
	id := int(req.GetUserId())
	if id == 0 {
		return nil, card_errors.ErrGrpcInvalidUserID
	}
	res, err := s.cardQuery.FindByUserID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseCard{
		Message: "successfully",
		Status:  "success",
		Data: &pb.CardResponse{
			Id:           int32(res.CardID),
			UserId:       int32(res.UserID),
			CardNumber:   res.CardNumber,
			CardType:     res.CardType,
			CardProvider: res.CardProvider,
			Cvv:          res.Cvv,
			ExpireDate:   formatCardTimeTim(res.ExpireDate),
			CreatedAt:    formatCardTime(res.CreatedAt),
			UpdatedAt:    formatCardTime(res.UpdatedAt),
		},
	}, nil
}

func (s *cardQueryHandleGrpc) FindUserCardByCardNumber(ctx context.Context, req *pb.FindByCardNumberRequest) (*pb.CardWithEmailResponse, error) {
	card_number := req.GetCardNumber()
	if card_number == "" {
		return nil, card_errors.ErrGrpcInvalidCardNumber
	}

	res, err := s.cardQuery.FindUserCardByCardNumber(ctx, card_number)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.CardWithEmailResponse{
		Id:           int32(res.CardID),
		UserId:       int32(res.UserID),
		CardNumber:   res.CardNumber,
		CardType:     res.CardType,
		Cvv:          res.Cvv,
		CardProvider: res.CardProvider,
		ExpireDate:   formatCardTimeTim(res.ExpireDate),
		CreatedAt:    formatCardTime(res.CreatedAt),
		UpdatedAt:    formatCardTime(res.UpdatedAt),
	}, nil
}
