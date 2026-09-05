package adapter

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
)

// guardSetter is implemented by every gRPC adapter so WithDependencyGuard can
// attach a guard without changing each constructor's signature for existing
// call sites.
type guardSetter interface {
	setGuard(*resilience.DependencyGuard)
}

// WithDependencyGuard attaches a dependency guard (per-call timeout + circuit
// breaker + bulkhead) to a gRPC adapter. Passing nil disables guarding.
//
// Usage:
//
//	guard := resilience.NewDependencyGuard("saldo", 5, 30, 100, 3*time.Second, srv.Logger)
//	saldoAdapter := adapter.NewSaldoAdapter(q, c, adapter.WithDependencyGuard(guard))
func WithDependencyGuard(guard *resilience.DependencyGuard) func(guardSetter) {
	return func(s guardSetter) {
		s.setGuard(guard)
	}
}
