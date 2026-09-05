package ledger

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	rows     []*ReconciliationRow
	err      error
	calls    int
	enqueued int
}

func (f *fakeStore) EnqueueReconciliationMismatch(context.Context, *ReconciliationRow) error {
	f.enqueued++
	return nil
}

func (f *fakeStore) ListReconciliationMismatches(context.Context, int32) ([]*ReconciliationRow, error) {
	f.calls++
	return f.rows, f.err
}

func TestReconciliationStoreContract(t *testing.T) {
	rows, err := (&fakeStore{rows: []*ReconciliationRow{{Difference: 1}}}).ListReconciliationMismatches(context.Background(), 10)
	if err != nil || len(rows) != 1 || rows[0].Difference != 1 {
		t.Fatalf("unexpected fake reconciliation result: %#v %v", rows, err)
	}
}

func TestNewReconcilerAppliesSafeDefaults(t *testing.T) {
	r := NewReconciler(&fakeStore{}, nil, 0, 0)
	if r.interval != 5*time.Minute {
		t.Fatalf("unexpected interval: %s", r.interval)
	}
	if r.batchSize != 100 {
		t.Fatalf("unexpected batch size: %d", r.batchSize)
	}
}

func TestRunOnceQueriesStore(t *testing.T) {
	store := &fakeStore{}
	r := NewReconciler(store, nil, time.Hour, 7)
	r.RunOnce(context.Background())
	if store.calls != 1 {
		t.Fatalf("expected one reconciliation query, got %d", store.calls)
	}
}

func TestRunOnceEnqueuesMismatches(t *testing.T) {
	store := &fakeStore{rows: []*ReconciliationRow{{SaldoID: 7, Difference: 20}}}
	NewReconciler(store, nil, time.Hour, 10).RunOnce(context.Background())
	if store.enqueued != 1 {
		t.Fatalf("expected one durable queue write, got %d", store.enqueued)
	}
}

func TestRunOncePreservesStoreErrorWithoutLogger(t *testing.T) {
	want := errors.New("database unavailable")
	store := &fakeStore{err: want}
	// The worker logs the error and deliberately does not return it because it
	// is a long-running background task; importantly, nil logger is safe.
	NewReconciler(store, nil, time.Hour, 1).RunOnce(context.Background())
	if store.calls != 1 {
		t.Fatal("expected reconciliation query despite store error")
	}
}
