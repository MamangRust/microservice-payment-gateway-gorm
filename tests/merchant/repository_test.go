package merchant_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/repository"
	user_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type MerchantRepositoryTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	db        *gorm.DB
	repo      repository.MerchantCommandRepository
	queryRepo repository.MerchantQueryRepository
	userRepo  user_repo.UserCommandRepository
	userID    int
}

func (s *MerchantRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.Require().NoError(s.ts.RunMigrations("user", "role", "auth", "merchant"))

	gormDB, err := s.ts.GormDB()
	s.Require().NoError(err)
	s.db = gormDB

	s.repo = repository.NewMerchantCommandRepository(gormDB)
	s.queryRepo = repository.NewMerchantQueryRepository(gormDB)
	s.userRepo = user_repo.NewUserCommandRepository(gormDB)

	user, err := s.userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Merchant",
		LastName:  "Owner",
		Email:     fmt.Sprintf("merchant.owner-%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *MerchantRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantRepositoryTestSuite) createSeedMerchant() (*models.Merchant, error) {
	return s.repo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		Name:   fmt.Sprintf("Test Merchant-%d", time.Now().UnixNano()),
		UserID: s.userID,
	})
}

func (s *MerchantRepositoryTestSuite) TestCreateMerchant() {
	merchant, err := s.repo.CreateMerchant(context.Background(), &requests.CreateMerchantRequest{
		Name:   "Test Merchant",
		UserID: s.userID,
	})
	s.NoError(err)
	s.NotNil(merchant)
	s.Equal("Test Merchant", merchant.Name)
	s.Equal(int32(s.userID), merchant.UserID)
}

func (s *MerchantRepositoryTestSuite) TestFindAllMerchants() {
	_, err := s.createSeedMerchant()
	s.Require().NoError(err)
	res, err := s.queryRepo.FindAllMerchants(context.Background(), &requests.FindAllMerchants{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantRepositoryTestSuite) TestFindById() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	found, err := s.queryRepo.FindByMerchantId(context.Background(), int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(merchant.MerchantID, found.MerchantID)
}

func (s *MerchantRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedMerchant()
	s.Require().NoError(err)
	res, err := s.queryRepo.FindByActive(context.Background(), &requests.FindAllMerchants{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantRepositoryTestSuite) TestFindByTrashed() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	_, err = s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.Require().NoError(err)
	res, err := s.queryRepo.FindByTrashed(context.Background(), &requests.FindAllMerchants{Page: 1, PageSize: 10, Search: ""})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *MerchantRepositoryTestSuite) TestUpdateMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	id := int(merchant.MerchantID)
	updated, err := s.repo.UpdateMerchant(context.Background(), &requests.UpdateMerchantRequest{
		MerchantID: &id, Name: "Updated Merchant", UserID: s.userID, Status: "active",
	})
	s.NoError(err)
	s.NotNil(updated)
	s.Equal("Updated Merchant", updated.Name)
}

func (s *MerchantRepositoryTestSuite) TestTrashMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	trashed, err := s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(trashed)
	s.NotNil(trashed.DeletedAt)
}

func (s *MerchantRepositoryTestSuite) TestRestoreMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	_, err = s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.Require().NoError(err)
	restored, err := s.repo.RestoreMerchant(context.Background(), int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(restored)
	s.Nil(restored.DeletedAt)
}

func (s *MerchantRepositoryTestSuite) TestDeleteMerchantPermanent() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	_, err = s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.Require().NoError(err)
	success, err := s.repo.DeleteMerchantPermanent(context.Background(), int(merchant.MerchantID))
	s.NoError(err)
	s.True(success)
	_, err = s.queryRepo.FindByMerchantId(context.Background(), int(merchant.MerchantID))
	s.Error(err)
}

func (s *MerchantRepositoryTestSuite) TestRestoreAllMerchant() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	_, err = s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.Require().NoError(err)
	success, err := s.repo.RestoreAllMerchant(context.Background())
	s.NoError(err)
	s.True(success)
	found, err := s.queryRepo.FindByMerchantId(context.Background(), int(merchant.MerchantID))
	s.NoError(err)
	s.NotNil(found)
}

func (s *MerchantRepositoryTestSuite) TestDeleteAllMerchantPermanent() {
	merchant, err := s.createSeedMerchant()
	s.Require().NoError(err)
	_, err = s.repo.TrashedMerchant(context.Background(), int(merchant.MerchantID))
	s.Require().NoError(err)
	success, err := s.repo.DeleteAllMerchantPermanent(context.Background())
	s.NoError(err)
	s.True(success)
	_, err = s.queryRepo.FindByMerchantId(context.Background(), int(merchant.MerchantID))
	s.Error(err)
}

func TestMerchantRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantRepositoryTestSuite))
}
