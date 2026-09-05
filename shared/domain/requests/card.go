package requests

import (
	"fmt"
	"time"

	methodtopup "github.com/MamangRust/microservice-payment-gateway-grpc/pkg/method_topup"
	"github.com/go-playground/validator/v10"
)

type MonthYearCardNumberCard struct {
	CardNumber string `json:"card_number" validate:"required"`
	Year       int    `json:"year" validate:"required"`
}

type FindAllCards struct {
	Search   string `json:"search" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

type CreateCardRequest struct {
	UserID       int       `json:"user_id"`
	CardType     string    `json:"card_type" validate:"required"`
	ExpireDate   time.Time `json:"expire_date" validate:"required"`
	CVV          string    `json:"cvv" validate:"required"`
	CardProvider string    `json:"card_provider" validate:"required"`
}

func (r *CreateCardRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	if r.CardType != "credit" && r.CardType != "debit" {
		return fmt.Errorf("card type must be credit or debit")
	}
	if !methodtopup.PaymentMethodValidator(r.CardProvider) {
		return fmt.Errorf("card provider not found")
	}
	return nil
}

type UpdateCardRequest struct {
	CardID       int       `json:"card_id" validate:"required,min=1"`
	UserID       int       `json:"user_id" validate:"required,min=1"`
	CardType     string    `json:"card_type" validate:"required"`
	ExpireDate   time.Time `json:"expire_date" validate:"required"`
	CVV          string    `json:"cvv" validate:"required"`
	CardProvider string    `json:"card_provider" validate:"required"`
}

func (r *UpdateCardRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}
	if r.CardType != "credit" && r.CardType != "debit" {
		return fmt.Errorf("card type must be credit or debit")
	}
	if !methodtopup.PaymentMethodValidator(r.CardProvider) {
		return fmt.Errorf("card provider not found")
	}
	return nil
}

type ToggleCardStatusRequest struct {
	CardID int `json:"card_id" validate:"required,min=1"`
}

func (r *ToggleCardStatusRequest) Validate() error { return validator.New().Struct(r) }

type UpdateCreditLimitRequest struct {
	CardID      int `json:"card_id" validate:"required,min=1"`
	CreditLimit int `json:"credit_limit" validate:"required,min=0"`
}

func (r *UpdateCreditLimitRequest) Validate() error { return validator.New().Struct(r) }

type RedeemPointsRequest struct {
	CardID int `json:"card_id" validate:"required,min=1"`
	Points int `json:"points" validate:"required,min=1"`
}

func (r *RedeemPointsRequest) Validate() error { return validator.New().Struct(r) }

type AuthorizeCardRequest struct {
	CardNumber     string `json:"card_number" validate:"required"`
	MerchantID     int    `json:"merchant_id" validate:"required,min=1"`
	Amount         int64  `json:"amount" validate:"required,min=1"`
	Currency       string `json:"currency" validate:"required,len=3"`
	PosEntryMode   string `json:"pos_entry_mode" validate:"required"`
	Mcc            string `json:"mcc" validate:"required"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

func (r *AuthorizeCardRequest) Validate() error { return validator.New().Struct(r) }

type ReverseTransactionRequest struct {
	TxnID          string `json:"txn_id" validate:"required"`
	CardNumber     string `json:"card_number" validate:"required"`
	Amount         int64  `json:"amount" validate:"required,min=1"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

func (r *ReverseTransactionRequest) Validate() error { return validator.New().Struct(r) }

type PostPaymentRequest struct {
	ReferenceID    string `json:"reference_id" validate:"required"`
	CardNumber     string `json:"card_number" validate:"required"`
	Amount         int64  `json:"amount" validate:"required,min=1"`
	PaymentChannel string `json:"payment_channel" validate:"required"`
	BillingID      *int   `json:"billing_id"`
}

func (r *PostPaymentRequest) Validate() error { return validator.New().Struct(r) }

type EarnRewardsRequest struct {
	CardNumber string `json:"card_number" validate:"required"`
	TxnID      string `json:"txn_id" validate:"required"`
	Amount     int64  `json:"amount" validate:"required,min=1"`
	Mcc        string `json:"mcc"`
}

func (r *EarnRewardsRequest) Validate() error { return validator.New().Struct(r) }

type RedeemRewardsRequest struct {
	CardNumber string `json:"card_number" validate:"required"`
	Points     int64  `json:"points" validate:"required,min=1"`
}

func (r *RedeemRewardsRequest) Validate() error { return validator.New().Struct(r) }
