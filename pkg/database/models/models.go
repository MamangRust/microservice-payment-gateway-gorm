package models

import (
	"time"
)

// ---------------------------------------------------------------------------
// Core
// ---------------------------------------------------------------------------

type User struct {
	UserID           int32      `gorm:"column:user_id;primaryKey" json:"user_id"`
	Firstname        string     `gorm:"column:firstname" json:"firstname"`
	Lastname         string     `gorm:"column:lastname" json:"lastname"`
	Email            string     `gorm:"column:email" json:"email"`
	Password         string     `gorm:"column:password" json:"-"`
	VerificationCode string     `gorm:"column:verification_code" json:"verification_code,omitempty"`
	IsVerified       *bool      `gorm:"column:is_verified" json:"is_verified,omitempty"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (User) TableName() string { return "users" }

type Role struct {
	RoleID      int32      `gorm:"column:role_id;primaryKey" json:"role_id"`
	RoleName    string     `gorm:"column:role_name" json:"role_name"`
	Description *string    `gorm:"-" json:"description"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (r Role) Name() string { return r.RoleName }

func (Role) TableName() string { return "roles" }

type UserRole struct {
	UserID int32 `gorm:"column:user_id;primaryKey" json:"user_id"`
	RoleID int32 `gorm:"column:role_id;primaryKey" json:"role_id"`
}

func (UserRole) TableName() string { return "user_roles" }

// ---------------------------------------------------------------------------
// Merchant
// ---------------------------------------------------------------------------

type Merchant struct {
	MerchantID int32      `gorm:"column:merchant_id;primaryKey" json:"merchant_id"`
	Name       string     `gorm:"column:name" json:"name"`
	ApiKey     string     `gorm:"column:api_key" json:"api_key"`
	UserID     int32      `gorm:"column:user_id" json:"user_id"`
	Status     string     `gorm:"column:status" json:"status"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (Merchant) TableName() string { return "merchants" }

type MerchantDocument struct {
	DocumentID   int32      `gorm:"column:document_id;primaryKey" json:"document_id"`
	MerchantID   int32      `gorm:"column:merchant_id" json:"merchant_id"`
	DocumentType string     `gorm:"column:document_type" json:"document_type"`
	DocumentUrl  string     `gorm:"column:document_url" json:"document_url"`
	Status       string     `gorm:"column:status" json:"status"`
	Note         *string    `gorm:"column:note" json:"note"`
	UploadedAt   *time.Time `gorm:"column:uploaded_at" json:"uploaded_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (MerchantDocument) TableName() string { return "merchant_documents" }

// ---------------------------------------------------------------------------
// Card
// ---------------------------------------------------------------------------

type Card struct {
	CardID             int32      `gorm:"column:card_id;primaryKey" json:"card_id"`
	UserID             int32      `gorm:"column:user_id" json:"user_id"`
	CardNumber         string     `gorm:"column:card_number" json:"card_number"`
	CardType           string     `gorm:"column:card_type" json:"card_type"`
	ExpireDate         time.Time  `gorm:"column:expire_date" json:"expire_date"`
	Cvv                string     `gorm:"column:cvv" json:"-"`
	CardProvider       string     `gorm:"column:card_provider" json:"card_provider"`
	Email              string     `gorm:"-" json:"email,omitempty"`
	Status             string     `gorm:"column:status" json:"status"`
	CreditLimit        int32      `gorm:"column:credit_limit" json:"credit_limit"`
	OutstandingBalance int64      `gorm:"column:outstanding_balance" json:"outstanding_balance"`
	RewardPoints       int32      `gorm:"column:reward_points" json:"reward_points"`
	CreatedAt          *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (Card) TableName() string { return "cards" }

type CardAuthTransaction struct {
	TxnID          string     `gorm:"column:txn_id;primaryKey" json:"txn_id"`
	CardNumber     string     `gorm:"column:card_number" json:"card_number"`
	MerchantID     int32      `gorm:"column:merchant_id" json:"merchant_id"`
	Amount         int64      `gorm:"column:amount" json:"amount"`
	Currency       string     `gorm:"column:currency" json:"currency"`
	Mcc            string     `gorm:"column:mcc" json:"mcc"`
	PosEntryMode   string     `gorm:"column:pos_entry_mode" json:"pos_entry_mode"`
	IdempotencyKey string     `gorm:"column:idempotency_key" json:"idempotency_key"`
	Status         string     `gorm:"column:status" json:"status"`
	RiskScore      int32      `gorm:"column:risk_score" json:"risk_score"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CardAuthTransaction) TableName() string { return "card_auth_transactions" }

type CardPayment struct {
	PaymentID      int32      `gorm:"column:payment_id;primaryKey" json:"payment_id"`
	PaymentUuid    string     `gorm:"column:payment_uuid" json:"payment_uuid"`
	CardNumber     string     `gorm:"column:card_number" json:"card_number"`
	BillingID      *int32     `gorm:"column:billing_id" json:"billing_id"`
	Amount         int64      `gorm:"column:amount" json:"amount"`
	PaymentChannel string     `gorm:"column:payment_channel" json:"payment_channel"`
	ReferenceID    *string    `gorm:"column:reference_id" json:"reference_id"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CardPayment) TableName() string { return "card_payments" }

type CardReward struct {
	RewardID     int32      `gorm:"column:reward_id;primaryKey" json:"reward_id"`
	CardNumber   string     `gorm:"column:card_number" json:"card_number"`
	TxnID        *string    `gorm:"column:txn_id" json:"txn_id"`
	Amount       int64      `gorm:"column:amount" json:"amount"`
	Mcc          *string    `gorm:"column:mcc" json:"mcc"`
	PointsEarned int32      `gorm:"column:points_earned" json:"points_earned"`
	ExpiresAt    time.Time  `gorm:"column:expires_at" json:"expires_at"`
	Redeemed     bool       `gorm:"column:redeemed" json:"redeemed"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CardReward) TableName() string { return "card_rewards" }

type BillingCycle struct {
	BillingID   int32      `gorm:"column:billing_id;primaryKey" json:"billing_id"`
	CardNumber  string     `gorm:"column:card_number" json:"card_number"`
	CycleStart  time.Time  `gorm:"column:cycle_start" json:"cycle_start"`
	CycleEnd    time.Time  `gorm:"column:cycle_end" json:"cycle_end"`
	DueDate     time.Time  `gorm:"column:due_date" json:"due_date"`
	TotalAmount int64      `gorm:"column:total_amount" json:"total_amount"`
	Status      string     `gorm:"column:status" json:"status"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (BillingCycle) TableName() string { return "billing_cycles" }

// ---------------------------------------------------------------------------
// Saldo
// ---------------------------------------------------------------------------

type Saldo struct {
	SaldoID        int32      `gorm:"column:saldo_id;primaryKey" json:"saldo_id"`
	CardNumber     string     `gorm:"column:card_number" json:"card_number"`
	TotalBalance   int64      `gorm:"column:total_balance" json:"total_balance"`
	WithdrawAmount *int64     `gorm:"column:withdraw_amount" json:"withdraw_amount"`
	WithdrawTime   *time.Time `gorm:"column:withdraw_time" json:"withdraw_time"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (Saldo) TableName() string { return "saldos" }

type BalanceLedger struct {
	EntryID       int64      `gorm:"column:entry_id;primaryKey" json:"entry_id"`
	OperationID   string     `gorm:"column:operation_id" json:"operation_id"`
	CardNumber    string     `gorm:"column:card_number" json:"card_number"`
	Direction     string     `gorm:"column:direction" json:"direction"`
	Amount        int64      `gorm:"column:amount" json:"amount"`
	Delta         int64      `gorm:"column:delta" json:"delta"`
	BalanceBefore int64      `gorm:"column:balance_before" json:"balance_before"`
	BalanceAfter  int64      `gorm:"column:balance_after" json:"balance_after"`
	SourceType    string     `gorm:"column:source_type" json:"source_type"`
	SourceID      *string    `gorm:"column:source_id" json:"source_id"`
	Note          *string    `gorm:"column:note" json:"note"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (BalanceLedger) TableName() string { return "balance_ledger" }

type ReconciliationQueue struct {
	QueueID               int64      `gorm:"column:queue_id;primaryKey" json:"queue_id"`
	SaldoID               int32      `gorm:"column:saldo_id" json:"saldo_id"`
	CardNumber            string     `gorm:"column:card_number" json:"card_number"`
	CurrentBalance        int64      `gorm:"column:current_balance" json:"current_balance"`
	LedgerBalance         int64      `gorm:"column:ledger_balance" json:"ledger_balance"`
	Difference            int64      `gorm:"column:difference" json:"difference"`
	LedgerEntries         int64      `gorm:"column:ledger_entries" json:"ledger_entries"`
	Status                string     `gorm:"column:status" json:"status"`
	MismatchCount         int64      `gorm:"column:mismatch_count" json:"mismatch_count"`
	FirstSeenAt           *time.Time `gorm:"column:first_seen_at" json:"first_seen_at"`
	LastSeenAt            *time.Time `gorm:"column:last_seen_at" json:"last_seen_at"`
	ResolvedAt            *time.Time `gorm:"column:resolved_at" json:"resolved_at"`
	ResolutionOperationID *string    `gorm:"column:resolution_operation_id" json:"resolution_operation_id"`
	ResolutionNote        *string    `gorm:"column:resolution_note" json:"resolution_note"`
	CreatedAt             *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ReconciliationQueue) TableName() string { return "reconciliation_queue" }

// ---------------------------------------------------------------------------
// Financial
// ---------------------------------------------------------------------------

type Transaction struct {
	TransactionID   int32      `gorm:"column:transaction_id;primaryKey" json:"transaction_id"`
	TransactionNo   string     `gorm:"column:transaction_no;type:uuid" json:"transaction_no"`
	CardNumber      string     `gorm:"column:card_number" json:"card_number"`
	Amount          int64      `gorm:"column:amount" json:"amount"`
	PaymentMethod   string     `gorm:"column:payment_method" json:"payment_method"`
	MerchantID      int32      `gorm:"column:merchant_id" json:"merchant_id"`
	MerchantName    string     `gorm:"-" json:"merchant_name,omitempty"`
	TransactionTime time.Time  `gorm:"column:transaction_time" json:"transaction_time"`
	Status          string     `gorm:"column:status" json:"status"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	OperationID     string     `gorm:"column:operation_id;type:uuid" json:"operation_id"`
	FailureReason   *string    `gorm:"column:failure_reason" json:"failure_reason"`
}

func (Transaction) TableName() string { return "transactions" }

type Topup struct {
	TopupID       int32      `gorm:"column:topup_id;primaryKey" json:"topup_id"`
	TopupNo       string     `gorm:"column:topup_no;type:uuid" json:"topup_no"`
	CardNumber    string     `gorm:"column:card_number" json:"card_number"`
	TopupAmount   int64      `gorm:"column:topup_amount" json:"topup_amount"`
	TopupMethod   string     `gorm:"column:topup_method" json:"topup_method"`
	TopupTime     time.Time  `gorm:"column:topup_time" json:"topup_time"`
	Status        string     `gorm:"column:status" json:"status"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	OperationID   string     `gorm:"column:operation_id;type:uuid" json:"operation_id"`
	FailureReason *string    `gorm:"column:failure_reason" json:"failure_reason"`
}

func (Topup) TableName() string { return "topups" }

type Transfer struct {
	TransferID     int32      `gorm:"column:transfer_id;primaryKey" json:"transfer_id"`
	TransferNo     string     `gorm:"column:transfer_no;type:uuid" json:"transfer_no"`
	TransferFrom   string     `gorm:"column:transfer_from" json:"transfer_from"`
	TransferTo     string     `gorm:"column:transfer_to" json:"transfer_to"`
	TransferAmount int64      `gorm:"column:transfer_amount" json:"transfer_amount"`
	TransferTime   time.Time  `gorm:"column:transfer_time" json:"transfer_time"`
	Status         string     `gorm:"column:status" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	OperationID    string     `gorm:"column:operation_id;type:uuid" json:"operation_id"`
	FailureReason  *string    `gorm:"column:failure_reason" json:"failure_reason"`
}

func (Transfer) TableName() string { return "transfers" }

type Withdraw struct {
	WithdrawID     int32      `gorm:"column:withdraw_id;primaryKey" json:"withdraw_id"`
	WithdrawNo     string     `gorm:"column:withdraw_no;type:uuid" json:"withdraw_no"`
	CardNumber     string     `gorm:"column:card_number" json:"card_number"`
	WithdrawAmount int64      `gorm:"column:withdraw_amount" json:"withdraw_amount"`
	WithdrawTime   time.Time  `gorm:"column:withdraw_time" json:"withdraw_time"`
	Status         string     `gorm:"column:status" json:"status"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
	OperationID    string     `gorm:"column:operation_id;type:uuid" json:"operation_id"`
	FailureReason  *string    `gorm:"column:failure_reason" json:"failure_reason"`
}

func (Withdraw) TableName() string { return "withdraws" }

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

type RefreshToken struct {
	RefreshTokenID int32      `gorm:"column:refresh_token_id;primaryKey" json:"refresh_token_id"`
	UserID         int32      `gorm:"column:user_id" json:"user_id"`
	Token          string     `gorm:"column:token" json:"token"`
	Expiration     time.Time  `gorm:"column:expiration" json:"expiration"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type ResetToken struct {
	ResetTokenID int32     `gorm:"column:id;primaryKey" json:"reset_token_id"`
	UserID       int32     `gorm:"column:user_id" json:"user_id"`
	Token        string    `gorm:"column:token" json:"token"`
	ExpiryDate   time.Time `gorm:"column:expiry_date" json:"expiry_date"`
}

func (ResetToken) TableName() string { return "reset_tokens" }
