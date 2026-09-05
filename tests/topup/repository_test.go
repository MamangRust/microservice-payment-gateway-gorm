package topup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	topup_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/topup/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type TopupRepositoryTestSuite struct {
	suite.Suite
	ts   *tests.TestSuite
	db   *gorm.DB
	repo topup_repo.Repositories

	userRepo  user_repo.UserCommandRepository
	cardRepo  card_repo.CardCommandRepository
	saldoRepo saldo_repo.Repositories

	cardNumber string
}

func (s *TopupRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts
	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card", "saldo", "topup"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	userRepos := user_repo.NewRepositories(gormDB)
	cardRepos := card_repo.NewRepositories(gormDB, nil)
	saldoRepos := saldo_repo.NewRepositories(gormDB, nil)

	cardAdapter := &topupCardRepoAdapter{
		CardQueryRepository:   cardRepos.CardQuery,
		CardCommandRepository: cardRepos.CardCommand,
	}
	s.repo = topup_repo.NewRepositories(gormDB, cardAdapter, saldoRepos)
	s.userRepo = userRepos.UserCommand()
	s.cardRepo = cardRepos.CardCommand
	s.saldoRepo = saldoRepos

	ctx := context.Background()
	user, err := s.userRepo.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Topup", LastName: "Owner",
		Email: fmt.Sprintf("topup.repo-%d@example.com", time.Now().UnixNano()), Password: "password123",
	})
	s.Require().NoError(err)

	card, err := s.cardRepo.CreateCard(ctx, &requests.CreateCardRequest{
		UserID: int(user.UserID), CardType: "debit", ExpireDate: time.Now().AddDate(2, 0, 0),
		CVV: "123", CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	_, err = s.saldoRepo.CreateSaldo(ctx, &requests.CreateSaldoRequest{
		CardNumber: s.cardNumber, TotalBalance: 0,
	})
	s.Require().NoError(err)
}

func (s *TopupRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *TopupRepositoryTestSuite) createSeedTopup() (*models.Topup, error) {
	return s.repo.CreateTopup(context.Background(), &requests.CreateTopupRequest{
		CardNumber: s.cardNumber, TopupAmount: 50000, TopupMethod: "visa",
	})
}

func (s *TopupRepositoryTestSuite) TestCreateTopup() {
	ctx := context.Background()
	topup, err := s.repo.CreateTopup(ctx, &requests.CreateTopupRequest{
		CardNumber: s.cardNumber, TopupAmount: 50000, TopupMethod: "visa",
	})
	s.NoError(err)
	s.NotNil(topup)
	s.Equal(int64(50000), topup.TopupAmount)
}

func (s *TopupRepositoryTestSuite) TestFindAllTopups() {
	_, err := s.createSeedTopup()
	s.Require().NoError(err)
	res, err := s.repo.FindAllTopups(context.Background(), &requests.FindAllTopups{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TopupRepositoryTestSuite) TestFindById() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	found, err := s.repo.FindById(context.Background(), int(topup.TopupID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(topup.TopupID, found.TopupID)
}

func (s *TopupRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedTopup()
	s.Require().NoError(err)
	res, err := s.repo.FindByActive(context.Background(), &requests.FindAllTopups{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TopupRepositoryTestSuite) TestFindByTrashed() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	_, err = s.repo.TrashedTopup(context.Background(), int(topup.TopupID))
	s.Require().NoError(err)
	res, err := s.repo.FindByTrashed(context.Background(), &requests.FindAllTopups{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
	_, err = s.repo.RestoreTopup(context.Background(), int(topup.TopupID))
	s.Require().NoError(err)
}

func (s *TopupRepositoryTestSuite) TestUpdateTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	id := int(topup.TopupID)
	updated, err := s.repo.UpdateTopup(context.Background(), &requests.UpdateTopupRequest{
		TopupID: &id, CardNumber: s.cardNumber, TopupAmount: 75000, TopupMethod: "visa",
	})
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int64(75000), updated.TopupAmount)
}

func (s *TopupRepositoryTestSuite) TestTrashTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	trashed, err := s.repo.TrashedTopup(context.Background(), int(topup.TopupID))
	s.NoError(err)
	s.NotNil(trashed)
	s.NotNil(trashed.DeletedAt)
}

func (s *TopupRepositoryTestSuite) TestRestoreTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	_, err = s.repo.TrashedTopup(context.Background(), int(topup.TopupID))
	s.Require().NoError(err)
	restored, err := s.repo.RestoreTopup(context.Background(), int(topup.TopupID))
	s.NoError(err)
	s.NotNil(restored)
	s.Nil(restored.DeletedAt)
}

func (s *TopupRepositoryTestSuite) TestDeleteTopupPermanent() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	_, err = s.repo.TrashedTopup(context.Background(), int(topup.TopupID))
	s.Require().NoError(err)
	success, err := s.repo.DeleteTopupPermanent(context.Background(), int(topup.TopupID))
	s.NoError(err)
	s.True(success)
	_, err = s.repo.FindById(context.Background(), int(topup.TopupID))
	s.Error(err)
}

func TestTopupRepositorySuite(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	suite.Run(t, new(TopupRepositoryTestSuite))
}
