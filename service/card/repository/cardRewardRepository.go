package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"gorm.io/gorm"
)

type cardRewardRepository struct {
	db *gorm.DB
}

func NewCardRewardRepository(db *gorm.DB) CardRewardRepository {
	return &cardRewardRepository{db: db}
}

func earnPoints(amount int64, mcc string) int32 {
	points := int32(amount / 10000)
	switch mcc {
	case "5812", "4511", "5541":
		points *= 2
	}
	if points < 1 {
		points = 1
	}
	return points
}

func (r *cardRewardRepository) EarnRewards(ctx context.Context, req *requests.EarnRewardsRequest) (*models.CardReward, error) {
	reward := &models.CardReward{
		CardNumber:   req.CardNumber,
		TxnID:        &req.TxnID,
		Amount:       req.Amount,
		Mcc:          &req.Mcc,
		PointsEarned: earnPoints(req.Amount, req.Mcc),
		ExpiresAt:    time.Now().AddDate(1, 0, 0),
	}
	if err := r.db.WithContext(ctx).Create(reward).Error; err != nil {
		return nil, sharedErrors.ErrFailed("earn card rewards").WithInternal(err)
	}
	return reward, nil
}

func (r *cardRewardRepository) GetBalance(ctx context.Context, cardNumber string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.CardReward{}).
		Where("card_number = ? AND redeemed = false AND expires_at > NOW()", cardNumber).
		Select("COALESCE(SUM(points_earned), 0)").Scan(&total).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("get reward balance").WithInternal(err)
	}
	return total, nil
}

func (r *cardRewardRepository) GetHistory(ctx context.Context, cardNumber string) ([]*models.CardReward, error) {
	var rewards []*models.CardReward
	err := r.db.WithContext(ctx).Where("card_number = ?", cardNumber).Order("created_at DESC").Find(&rewards).Error
	if err != nil {
		return nil, sharedErrors.ErrFailed("get reward history").WithInternal(err)
	}
	return rewards, nil
}

func (r *cardRewardRepository) RedeemRewards(ctx context.Context, cardNumber string, points int64) (int64, error) {
	var rewards []models.CardReward
	err := r.db.WithContext(ctx).
		Where("card_number = ? AND redeemed = false AND expires_at > NOW() AND points_earned > 0", cardNumber).
		Order("created_at ASC").Find(&rewards).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("find redeemable rewards").WithInternal(err)
	}

	var ids []int32
	var total int64
	for _, row := range rewards {
		if total >= points {
			break
		}
		ids = append(ids, row.RewardID)
		total += int64(row.PointsEarned)
	}
	if total < points {
		return 0, sharedErrors.ErrBadRequest.WithMessage("Insufficient reward points")
	}

	err = r.db.WithContext(ctx).Model(&models.CardReward{}).
		Where("reward_id IN ?", ids).Update("redeemed", true).Error
	if err != nil {
		return 0, sharedErrors.ErrFailed("redeem rewards").WithInternal(err)
	}
	return points, nil
}
