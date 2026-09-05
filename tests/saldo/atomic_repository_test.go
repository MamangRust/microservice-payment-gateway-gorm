package saldo_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type SaldoAtomicRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	db       *gorm.DB
	repo     saldo_repo.Repositories
	userRepo user_repo.UserCommandRepository
	cardRepo card_repo.CardCommandRepository
}

func (s *SaldoAtomicRepositoryTestSuite) SetupSuite() {
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
}

func (s *SaldoAtomicRepositoryTestSuite) TearDownSuite() { s.ts.Teardown() }

func (s *SaldoAtomicRepositoryTestSuite) createSaldo(balance int) string {
	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Atomic", LastName: "Saldo", Email: fmt.Sprintf("saldo.atomic-%d@example.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)
	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0), CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	_, err = s.repo.CreateSaldo(ctx, &requests.CreateSaldoRequest{CardNumber: card.CardNumber, TotalBalance: balance})
	s.Require().NoError(err)
	return card.CardNumber
}

func (s *SaldoAtomicRepositoryTestSuite) TestDebitAndCreditAreAtomic() {
	ctx := context.Background()
	cardNumber := s.createSaldo(100000)

	debited, err := s.repo.DebitSaldo(ctx, &requests.DebitSaldoRequest{CardNumber: cardNumber, Amount: 25000})
	s.Require().NoError(err)
	s.Equal(int64(75000), debited.TotalBalance)

	credited, err := s.repo.CreditSaldo(ctx, &requests.CreditSaldoRequest{CardNumber: cardNumber, Amount: 10000})
	s.Require().NoError(err)
	s.Equal(int64(85000), credited.TotalBalance)

	_, err = s.repo.DebitSaldo(ctx, &requests.DebitSaldoRequest{CardNumber: cardNumber, Amount: 90000})
	s.Require().Error(err)
	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr))
	s.Equal(http.StatusConflict, appErr.Code)

	current, err := s.repo.FindByCardNumber(ctx, cardNumber)
	s.Require().NoError(err)
	s.Equal(int64(85000), current.TotalBalance)
}

func (s *SaldoAtomicRepositoryTestSuite) TestConcurrentDebitsNeverOverspend() {
	ctx := context.Background()
	cardNumber := s.createSaldo(100000)

	const attempts = 20
	const amount = 10000
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			defer wg.Done()
			_, err := s.repo.DebitSaldo(ctx, &requests.DebitSaldoRequest{CardNumber: cardNumber, Amount: amount})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil { successes++ }
	}
	s.Equal(10, successes)

	current, err := s.repo.FindByCardNumber(ctx, cardNumber)
	s.Require().NoError(err)
	s.Equal(int64(0), current.TotalBalance)
}

func TestSaldoAtomicRepositorySuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(SaldoAtomicRepositoryTestSuite))
}
