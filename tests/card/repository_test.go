package card_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type CardRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	db       *gorm.DB
	repo     *repository.Repositories
	userRepo user_repo.Repositories
	userID   int
}

func (s *CardRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "card"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.repo = repository.NewRepositories(gormDB, nil)
	s.userRepo = user_repo.NewRepositories(gormDB)

	// Create a user for card ownership
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Card",
		LastName:  "Owner",
		Email:     fmt.Sprintf("card.owner-%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *CardRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *CardRepositoryTestSuite) createSeedCard() (*models.Card, error) {
	return s.repo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	})
}

func (s *CardRepositoryTestSuite) TestCreateCard() {
	ctx := context.Background()
	req := &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	}

	res, err := s.repo.CardCommand.CreateCard(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.NotEmpty(res.CardNumber)
}

func (s *CardRepositoryTestSuite) TestFindAllCards() {
	_, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.CardQuery.FindAllCards(ctx, &requests.FindAllCards{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *CardRepositoryTestSuite) TestFindById() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.CardQuery.FindById(ctx, int(card.CardID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(card.CardID, found.CardID)
}

func (s *CardRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.CardQuery.FindByActive(ctx, &requests.FindAllCards{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *CardRepositoryTestSuite) TestFindByTrashed() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.Require().NoError(err)

	res, err := s.repo.CardQuery.FindByTrashed(ctx, &requests.FindAllCards{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *CardRepositoryTestSuite) TestUpdateCard() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	req := &requests.UpdateCardRequest{
		CardID:       int(card.CardID),
		UserID:       s.userID,
		CardType:     "credit",
		ExpireDate:   time.Now().AddDate(6, 0, 0),
		CVV:          "456",
		CardProvider: "MasterCard",
	}

	res, err := s.repo.CardCommand.UpdateCard(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("credit", res.CardType)
}

func (s *CardRepositoryTestSuite) TestTrashCard() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.NoError(err)
	s.NotNil(trashed)
	s.NotNil(trashed.DeletedAt)
}

func (s *CardRepositoryTestSuite) TestRestoreCard() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.Require().NoError(err)

	restored, err := s.repo.CardCommand.RestoreCard(ctx, int(card.CardID))
	s.NoError(err)
	s.NotNil(restored)
	s.Nil(restored.DeletedAt)
}

func (s *CardRepositoryTestSuite) TestDeleteCardPermanent() {
	card, err := s.createSeedCard()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.Require().NoError(err)

	success, err := s.repo.CardCommand.DeleteCardPermanent(ctx, int(card.CardID))
	s.NoError(err)
	s.True(success)

	_, err = s.repo.CardQuery.FindById(ctx, int(card.CardID))
	s.Error(err)
}

func (s *CardRepositoryTestSuite) TestRestoreAllCard() {
	ctx := context.Background()
	card, err := s.createSeedCard()
	s.Require().NoError(err)

	_, err = s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.Require().NoError(err)

	success, err := s.repo.CardCommand.RestoreAllCard(ctx)
	s.NoError(err)
	s.True(success)

	found, err := s.repo.CardQuery.FindById(ctx, int(card.CardID))
	s.NoError(err)
	s.NotNil(found)
}

func (s *CardRepositoryTestSuite) TestDeleteAllCardPermanent() {
	ctx := context.Background()
	card, err := s.createSeedCard()
	s.Require().NoError(err)

	_, err = s.repo.CardCommand.TrashedCard(ctx, int(card.CardID))
	s.Require().NoError(err)

	success, err := s.repo.CardCommand.DeleteAllCardPermanent(ctx)
	s.NoError(err)
	s.True(success)

	_, err = s.repo.CardQuery.FindById(ctx, int(card.CardID))
	s.Error(err)
}

func TestCardRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CardRepositoryTestSuite))
}
