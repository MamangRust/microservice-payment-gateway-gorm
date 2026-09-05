package transaction_test

import (
	"context"

	pbAISecurity "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type mockAISecurityServer struct {
	pbAISecurity.UnimplementedAISecurityServiceServer
}

func (m *mockAISecurityServer) DetectFraud(ctx context.Context, in *pbAISecurity.FraudRequest) (*pbAISecurity.FraudResponse, error) {
	return &pbAISecurity.FraudResponse{
		TransactionId: in.TransactionId,
		RiskScore:     0.1,
		IsFraudulent:  false,
		Reason:        "Mocked safe transaction",
	}, nil
}

// transactionCardRepo wraps the card repository to satisfy transaction's CardRepository interface.
type transactionCardRepo struct {
	query   card_repo.CardQueryRepository
	command card_repo.CardCommandRepository
}

func (r *transactionCardRepo) FindCardByUserId(ctx context.Context, user_id int) (*models.Card, error) {
	return r.query.FindCardByUserId(ctx, user_id)
}
func (r *transactionCardRepo) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error) {
	return r.query.FindUserCardByCardNumber(ctx, card_number)
}
func (r *transactionCardRepo) FindCardByCardNumber(ctx context.Context, card_number string) (*models.Card, error) {
	return r.query.FindCardByCardNumber(ctx, card_number)
}
func (r *transactionCardRepo) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.Card, error) {
	return r.command.UpdateCard(ctx, request)
}
