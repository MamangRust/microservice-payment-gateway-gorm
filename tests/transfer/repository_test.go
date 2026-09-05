package transfer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type TransferRepositoryTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	db          *gorm.DB
	commandRepo repository.TransferCommandRepository
	queryRepo   repository.TransferQueryRepository
	userRepo    user_repo.UserCommandRepository
	cardRepo    card_repo.Repositories
	saldoRepo   saldo_repo.Repositories
	senderCardNumber   string
	receiverCardNumber string
}

func (s *TransferRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "transfer"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.userRepo = user_repo.NewUserCommandRepository(gormDB)
	s.cardRepo = *card_repo.NewRepositories(gormDB, nil)
	s.saldoRepo = saldo_repo.NewRepositories(gormDB, nil)
	s.commandRepo = repository.NewTransferCommandRepository(gormDB)
	s.queryRepo = repository.NewTransferQueryRepository(gormDB)

	ctx := context.Background()
	sender, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Sender", LastName: "Repo", Email: fmt.Sprintf("sender.repo-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	sCard, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(sender.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "111", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.senderCardNumber = sCard.CardNumber

	receiver, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Receiver", LastName: "Repo", Email: fmt.Sprintf("receiver.repo-%d@test.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	rCard, err := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(receiver.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(1, 0, 0), CVV: "222", CardProvider: "mastercard",
	})
	s.Require().NoError(err)
	s.receiverCardNumber = rCard.CardNumber
}

func (s *TransferRepositoryTestSuite) TearDownSuite() { s.ts.Teardown() }

func (s *TransferRepositoryTestSuite) createSeedTransfer() (*models.Transfer, error) {
	return s.commandRepo.CreateTransfer(context.Background(), &requests.CreateTransferRequest{
		TransferFrom: s.senderCardNumber, TransferTo: s.receiverCardNumber, TransferAmount: 25000,
	})
}

func (s *TransferRepositoryTestSuite) TestCreateTransfer() {
	txn, err := s.createSeedTransfer()
	s.NoError(err)
	s.NotNil(txn)
	s.Equal(int64(25000), txn.TransferAmount)
}

func (s *TransferRepositoryTestSuite) TestFindAll() {
	_, err := s.createSeedTransfer()
	s.Require().NoError(err)
	res, err := s.queryRepo.FindAll(context.Background(), &requests.FindAllTransfers{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TransferRepositoryTestSuite) TestFindById() {
	txn, err := s.createSeedTransfer()
	s.Require().NoError(err)
	found, err := s.queryRepo.FindById(context.Background(), int(txn.TransferID))
	s.NoError(err)
	s.NotNil(found)
}

func TestTransferRepositorySuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TransferRepositoryTestSuite))
}
