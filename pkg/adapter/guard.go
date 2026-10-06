package adapter

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// GuardSetter is implemented by every gRPC adapter so WithDependencyGuard can
// attach a guard without changing each constructor's signature for existing
// call sites.
type GuardSetter interface {
	SetGuard(*resilience.DependencyGuard)
}

// WithDependencyGuard attaches a dependency guard (per-call timeout + circuit
// breaker + bulkhead) to a gRPC adapter. Passing nil disables guarding.
//
// Usage:
//
//	guard := resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)
//	saldoAdapter := adapter.NewSaldoAdapter(q, c, adapter.WithDependencyGuard(guard))
func WithDependencyGuard(guard *resilience.DependencyGuard) GuardOption {
	return func(s GuardSetter) {
		s.SetGuard(guard)
	}
}

// GuardOption configures a gRPC adapter. It is shared by every adapter
// constructor (flat or sub-package) so call sites stay uniform.
type GuardOption func(GuardSetter)

// callGuarded runs fn through the guard when one is configured, and calls it
// directly otherwise. Every adapter method funnels through this so an adapter
// built without WithDependencyGuard still works.
func callGuarded(ctx context.Context, guard *resilience.DependencyGuard, fn func(context.Context) error) error {
	if guard == nil {
		return fn(ctx)
	}
	return guard.Call(ctx, fn)
}

// CallGuarded is the exported entry point used by sub-package adapters
// (pkg/adapter/role, pkg/adapter/user_role). It mirrors callGuarded.
func CallGuarded(ctx context.Context, guard *resilience.DependencyGuard, fn func(context.Context) error) error {
	return callGuarded(ctx, guard, fn)
}
