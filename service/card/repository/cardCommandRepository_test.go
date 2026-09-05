package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB creates an in-memory SQLite database with the cards table.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Card{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// insertCard inserts a card row for test setup.
func insertCard(t *testing.T, db *gorm.DB, card *models.Card) {
	t.Helper()
	if err := db.Create(card).Error; err != nil {
		t.Fatalf("failed to insert test card: %v", err)
	}
}

// assertAppErrorCode verifies the returned error is an AppError with the
// expected HTTP code and message.
func assertAppErrorCode(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *sharedErrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *sharedErrors.AppError, got %T: %v", err, err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("AppError.Code = %d, want %d (message: %s)", appErr.Code, wantCode, appErr.Message)
	}
	if appErr.Message != wantMsg {
		t.Fatalf("AppError.Message = %q, want %q", appErr.Message, wantMsg)
	}
}

func TestToggleCardStatusCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 99})

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestToggleCardStatusInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	// Close the underlying *sql.DB to force connection errors.
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 1})

	assertAppErrorCode(t, err, 500, "Failed to toggle card status")
}

func TestToggleCardStatusSuccess(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active"})
	repo := NewCardCommandRepository(db)

	card, err := repo.ToggleCardStatus(context.Background(), &requests.ToggleCardStatusRequest{CardID: 1})
	if err != nil {
		t.Fatalf("ToggleCardStatus() unexpected error: %v", err)
	}
	if card.Status != "inactive" {
		t.Fatalf("expected status 'inactive', got %q", card.Status)
	}
}

func TestUpdateCardCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       99,
		UserID:       1,
		CardType:     "credit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "123",
		CardProvider: "visa",
	})

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestUpdateCardInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       1,
		UserID:       1,
		CardType:     "credit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "123",
		CardProvider: "visa",
	})

	assertAppErrorCode(t, err, 500, "Failed to update card")
}

func TestUpdateCardSuccess(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active"})
	repo := NewCardCommandRepository(db)

	_, err := repo.UpdateCard(context.Background(), &requests.UpdateCardRequest{
		CardID:       1,
		UserID:       1,
		CardType:     "debit",
		ExpireDate:   time.Date(2028, 12, 1, 0, 0, 0, 0, time.UTC),
		CVV:          "456",
		CardProvider: "mastercard",
	})
	if err != nil {
		t.Fatalf("UpdateCard() unexpected error: %v", err)
	}
}

func TestTrashedCardCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.TrashedCard(context.Background(), 99)

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestTrashedCardInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.TrashedCard(context.Background(), 1)

	assertAppErrorCode(t, err, 500, "Failed to trash card")
}

func TestTrashedCardSuccess(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active"})
	repo := NewCardCommandRepository(db)

	_, err := repo.TrashedCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("TrashedCard() unexpected error: %v", err)
	}
}

func TestRestoreCardCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.RestoreCard(context.Background(), 99)

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestRestoreCardInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.RestoreCard(context.Background(), 1)

	assertAppErrorCode(t, err, 500, "Failed to restore card")
}

func TestRestoreCardSuccess(t *testing.T) {
	db := newTestDB(t)
	// Insert a soft-deleted card.
	now := time.Now()
	insertCard(t, db, &models.Card{CardID: 1, Status: "active", DeletedAt: &now})
	repo := NewCardCommandRepository(db)

	_, err := repo.RestoreCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("RestoreCard() unexpected error: %v", err)
	}
}

func TestUpdateCreditLimitCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      99,
		CreditLimit: 5_000_000,
	})

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestUpdateCreditLimitInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      1,
		CreditLimit: 5_000_000,
	})

	assertAppErrorCode(t, err, 500, "Failed to update credit limit")
}

func TestUpdateCreditLimitSuccess(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active", CreditLimit: 1_000_000})
	repo := NewCardCommandRepository(db)

	_, err := repo.UpdateCreditLimit(context.Background(), &requests.UpdateCreditLimitRequest{
		CardID:      1,
		CreditLimit: 5_000_000,
	})
	if err != nil {
		t.Fatalf("UpdateCreditLimit() unexpected error: %v", err)
	}
}

func TestRedeemPointsCardNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 99,
		Points: 100,
	})

	assertAppErrorCode(t, err, 404, "card not found")
}

func TestRedeemPointsInsufficientRewardPoints(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active", RewardPoints: 50})
	repo := NewCardCommandRepository(db)

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})

	assertAppErrorCode(t, err, 400, "insufficient reward points")
}

func TestRedeemPointsInternalError(t *testing.T) {
	db := newTestDB(t)
	repo := NewCardCommandRepository(db)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	_, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})

	assertAppErrorCode(t, err, 500, "Failed to redeem reward points")
}

func TestRedeemPointsSuccess(t *testing.T) {
	db := newTestDB(t)
	insertCard(t, db, &models.Card{CardID: 1, Status: "active", RewardPoints: 200})
	repo := NewCardCommandRepository(db)

	card, err := repo.RedeemPoints(context.Background(), &requests.RedeemPointsRequest{
		CardID: 1,
		Points: 100,
	})
	if err != nil {
		t.Fatalf("RedeemPoints() unexpected error: %v", err)
	}
	if card.RewardPoints != 100 {
		t.Fatalf("expected 100 reward points, got %d", card.RewardPoints)
	}
}
