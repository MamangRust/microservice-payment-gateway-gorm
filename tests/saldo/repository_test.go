package saldo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type SaldoRepositoryTestSuite struct {
	suite.Suite
	ts   *tests.TestSuite
	db   *gorm.DB
	repo saldo_repo.Repositories
	userRepo user_repo.UserCommandRepository
	cardRepo card_repo.CardCommandRepository
	cardNumber string
}

func (s *SaldoRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.repo = saldo_repo.NewRepositories(gormDB, nil)
	s.userRepo = user_repo.NewRepositories(gormDB).UserCommand()
	s.cardRepo = card_repo.NewRepositories(gormDB, nil).CardCommand

	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Saldo", LastName: "Owner", Email: fmt.Sprintf("saldo.repo-%d@example.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber
	_, err = s.repo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: s.cardNumber, TotalBalance: 100000})
	s.Require().NoError(err)
}

func (s *SaldoRepositoryTestSuite) TearDownSuite() { s.ts.Teardown() }

func (s *SaldoRepositoryTestSuite) TestFindById() {
	ctx := context.Background()
	saldos, _ := s.repo.FindAllSaldos(ctx, &requests.FindAllSaldos{Page: 1, PageSize: 10, Search: ""})
	s.GreaterOrEqual(len(saldos), 1)
}

func (s *SaldoRepositoryTestSuite) TestFindByCardNumber() {
	ctx := context.Background()
	found, err := s.repo.FindByCardNumber(ctx, s.cardNumber)
	s.NoError(err)
	s.NotNil(found)
	s.IsType(&models.Saldo{}, found)
}

func (s *SaldoRepositoryTestSuite) TestUpdateSaldoBalance() {
	ctx := context.Background()
	updated, err := s.repo.UpdateSaldoBalance(ctx, &requests.UpdateSaldoBalance{
		CardNumber: s.cardNumber, TotalBalance: 200000,
	})
	s.NoError(err)
	s.NotNil(updated)
}

func TestSaldoRepositorySuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(SaldoRepositoryTestSuite))
}
