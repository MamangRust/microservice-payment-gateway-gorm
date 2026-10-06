package adapter

import (
	"context"

	pbai "github.com/MamangRust/microservice-payment-gateway-grpc/pb/ai_security"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// FraudCheckRequest is the request envelope calling services use to ask the AI
// security service to score a transaction for fraud.
type FraudCheckRequest struct {
	TransactionID string
	MerchantID    int
	UserID        int
	Amount        float64
	PaymentMethod string
}

// FraudCheckResponse is the fraud score returned by the AI security service.
type FraudCheckResponse struct {
	TransactionID string
	RiskScore     float64
	IsFraudulent  bool
	Reason        string
}

// SecurityDomain enumerates the business domains the AI security service can
// screen. Values mirror the proto SecurityDomain enum.
type SecurityDomain int32

const (
	SecurityDomainUnknown     SecurityDomain = 0
	SecurityDomainCard        SecurityDomain = 1
	SecurityDomainMerchant    SecurityDomain = 2
	SecurityDomainSaldo       SecurityDomain = 3
	SecurityDomainTopup       SecurityDomain = 4
	SecurityDomainTransaction SecurityDomain = 5
	SecurityDomainTransfer    SecurityDomain = 6
	SecurityDomainWithdraw    SecurityDomain = 7
)

// SecurityCheckRequest asks the AI security service to verify a single entity
// (a card, merchant, or user) within a given domain.
type SecurityCheckRequest struct {
	Domain   SecurityDomain
	EntityID string
	Amount   float64
	Metadata map[string]string
}

// SecurityCheckResponse is the verdict returned by the AI security service.
type SecurityCheckResponse struct {
	IsSafe    bool
	RiskScore float64
	Reason    string
	Action    string
}

// AISecurityAdapter is the contract consumers depend on for AI-driven fraud and
// security screening. It wraps the generated AI security gRPC client.
type AISecurityAdapter interface {
	DetectFraud(ctx context.Context, request *FraudCheckRequest) (*FraudCheckResponse, error)
	VerifySecurity(ctx context.Context, request *SecurityCheckRequest) (*SecurityCheckResponse, error)
}

type aisecurityGRPCAdapter struct {
	Client pbai.AISecurityServiceClient
	guard  *resilience.DependencyGuard
}

func (a *aisecurityGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewAISecurityAdapter wraps the generated AI security gRPC client.
func NewAISecurityAdapter(client pbai.AISecurityServiceClient, opts ...GuardOption) AISecurityAdapter {
	a := &aisecurityGRPCAdapter{
		Client: client,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *aisecurityGRPCAdapter) DetectFraud(ctx context.Context, request *FraudCheckRequest) (*FraudCheckResponse, error) {
	var resp *pbai.FraudResponse
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.Client.DetectFraud(callCtx, &pbai.FraudRequest{
			TransactionId: request.TransactionID,
			MerchantId:    int32(request.MerchantID),
			UserId:        int32(request.UserID),
			Amount:        request.Amount,
			PaymentMethod: request.PaymentMethod,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	return &FraudCheckResponse{
		TransactionID: resp.TransactionId,
		RiskScore:     resp.RiskScore,
		IsFraudulent:  resp.IsFraudulent,
		Reason:        resp.Reason,
	}, nil
}

func (a *aisecurityGRPCAdapter) VerifySecurity(ctx context.Context, request *SecurityCheckRequest) (*SecurityCheckResponse, error) {
	var resp *pbai.SecurityResponse
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.Client.VerifySecurity(callCtx, &pbai.SecurityRequest{
			Domain:   pbai.SecurityDomain(request.Domain),
			EntityId: request.EntityID,
			Amount:   request.Amount,
			Metadata: request.Metadata,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	return &SecurityCheckResponse{
		IsSafe:    resp.IsSafe,
		RiskScore: resp.RiskScore,
		Reason:    resp.Reason,
		Action:    resp.Action,
	}, nil
}
