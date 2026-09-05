package transaction_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	merchant_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type TransactionRepositoryTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	commandRepo repository.TransactionCommandRepository
	queryRepo   repository.TransactionQueryRepository
	userRepo    user_repo.UserCommandRepository
	cardRepo    card_repo.Repositories
	merchantRepo merchant_repo.Repositories
	customerCardNumber string
	merchantID  int
}

func (s *TransactionRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "merchant", "saldo", "transaction"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, nil)
	s.merchantRepo = merchant_repo.NewRepositories(gormDB, nil)

	transactionRepos := repository.NewRepositories(gormDB, nil, nil, nil)
	s.commandRepo = transactionRepos
	s.queryRepo = transactionRepos

	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Repo", LastName: "Owner", Email: fmt.Sprintf("repo-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.customerCardNumber = card.CardNumber
	merchant, err := s.merchantRepo.CreateMerchant(ctx, &requests.CreateMerchantRequest{
		Name: "Repo Merchant", UserID: int(user.UserID),
	})
	s.Require().NoError(err)
	s.merchantID = int(merchant.MerchantID)
}

func (s *TransactionRepositoryTestSuite) TearDownSuite() { s.ts.Teardown() }

func (s *TransactionRepositoryTestSuite) createSeedTransaction() (*models.Transaction, error) {
	merchantID := s.merchantID
	return s.commandRepo.CreateTransaction(context.Background(), &requests.CreateTransactionRequest{
		CardNumber: s.customerCardNumber, Amount: 100000, MerchantID: &merchantID,
		PaymentMethod: "debit", TransactionTime: time.Now(),
	})
}

func (s *TransactionRepositoryTestSuite) TestCreateTransaction() {
	txn, err := s.createSeedTransaction()
	s.NoError(err)
	s.NotNil(txn)
	s.Equal(int64(100000), txn.Amount)
}

func (s *TransactionRepositoryTestSuite) TestFindAll() {
	_, err := s.createSeedTransaction()
	s.Require().NoError(err)
	res, err := s.queryRepo.FindAllTransactions(context.Background(), &requests.FindAllTransactions{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransactionRepositoryTestSuite) TestFindById() {
	txn, err := s.createSeedTransaction()
	s.Require().NoError(err)
	found, err := s.queryRepo.FindById(context.Background(), int(txn.TransactionID))
	s.NoError(err)
	s.NotNil(found)
}

func TestTransactionRepositorySuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransactionRepositoryTestSuite))
}
