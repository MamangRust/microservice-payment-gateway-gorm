package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errTransport simulates a transport-level failure.
var errTransport = status.Error(codes.Unavailable, "dependency down")

// errBusiness simulates a normal dependency response error that must NOT open
// the circuit (e.g. not found, invalid argument).
var errBusiness = status.Error(codes.NotFound, "record not found")

func TestDependencyGuard_PassesThroughSuccess(t *testing.T) {
	g := NewDependencyGuard("test", 3, 30, 10, time.Second, nil)

	var called atomic.Int32
	err := g.Call(context.Background(), func(ctx context.Context) error {
		called.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("fn called %d times, want 1", called.Load())
	}
}

func TestDependencyGuard_AppliesTimeout(t *testing.T) {
	g := NewDependencyGuard("test", 3, 30, 10, 50*time.Millisecond, nil)

	start := time.Now()
	err := g.Call(context.Background(), func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) && status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("call took %v, timeout not applied", elapsed)
	}
}

func TestDependencyGuard_BulkheadRejectsWhenFull(t *testing.T) {
	g := NewDependencyGuard("test", 3, 30, 1, time.Second, nil)

	block := make(chan struct{})
	release := make(chan struct{})

	// Occupy the single permit.
	go func() {
		_ = g.Call(context.Background(), func(ctx context.Context) error {
			close(block)
			<-release
			return nil
		})
	}()
	<-block

	err := g.Call(context.Background(), func(ctx context.Context) error { return nil })
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted from bulkhead, got %v", err)
	}
	close(release)
}

func TestDependencyGuard_CircuitOpensOnTransportFailures(t *testing.T) {
	g := NewDependencyGuard("test", 3, 30, 10, time.Second, nil)

	// Trip the breaker with 3 transport failures.
	for i := 0; i < 3; i++ {
		err := g.Call(context.Background(), func(ctx context.Context) error {
			return errTransport
		})
		if status.Code(err) != codes.Unavailable {
			t.Fatalf("iteration %d: expected Unavailable, got %v", i, err)
		}
	}

	// The 4th call must fail fast without invoking fn.
	var called atomic.Int32
	err := g.Call(context.Background(), func(ctx context.Context) error {
		called.Add(1)
		return nil
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected fail-fast Unavailable, got %v", err)
	}
	if called.Load() != 0 {
		t.Fatalf("fn called while circuit open, want 0 calls, got %d", called.Load())
	}
}

func TestDependencyGuard_BusinessErrorsDoNotOpenCircuit(t *testing.T) {
	g := NewDependencyGuard("test", 3, 30, 10, time.Second, nil)

	// Many business errors must NOT trip the breaker.
	for i := 0; i < 10; i++ {
		err := g.Call(context.Background(), func(ctx context.Context) error {
			return errBusiness
		})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("expected NotFound pass-through, got %v", err)
		}
	}

	// The breaker must still be closed and fn must still run.
	var called atomic.Int32
	err := g.Call(context.Background(), func(ctx context.Context) error {
		called.Add(1)
		return nil
	})
	if err != nil || called.Load() != 1 {
		t.Fatalf("circuit should be closed after business errors (err=%v, called=%d)", err, called.Load())
	}
}

func TestDependencyGuard_RecoversAfterTimeout(t *testing.T) {
	g := NewDependencyGuard("test", 2, 1, 10, time.Second, nil)

	// Trip the breaker.
	for i := 0; i < 2; i++ {
		_ = g.Call(context.Background(), func(ctx context.Context) error { return errTransport })
	}
	if !g.breaker.IsOpen() {
		t.Fatal("breaker should be open")
	}

	// After the timeout (half-open window), a run of successful calls closes it
	// again. The breaker requires halfOpenMaxRequests (5) consecutive successes.
	time.Sleep(1100 * time.Millisecond)
	for i := 0; i < 5; i++ {
		if err := g.Call(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
			t.Fatalf("iteration %d: expected recovery, got %v", i, err)
		}
	}
	if g.breaker.IsOpen() {
		t.Fatal("breaker should be closed after recovery")
	}
}

func TestDependencyGuard_NilGuardFallsThrough(t *testing.T) {
	var g *DependencyGuard
	var called atomic.Int32
	err := g.Call(context.Background(), func(ctx context.Context) error {
		called.Add(1)
		return nil
	})
	if err != nil || called.Load() != 1 {
		t.Fatalf("nil guard must pass through (err=%v, called=%d)", err, called.Load())
	}
}
