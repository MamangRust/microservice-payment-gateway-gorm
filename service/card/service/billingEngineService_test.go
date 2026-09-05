package service

import (
	"testing"
	"time"
)

func TestBillingPeriodClampsMonthEndWithoutOverflow(t *testing.T) {
	loc := time.FixedZone("UTC+7", 7*60*60)

	tests := []struct {
		name       string
		now        time.Time
		billingDay int
		wantStart  time.Time
		wantEnd    time.Time
	}{
		{
			name:       "after march 31 closes march period",
			now:        time.Date(2026, time.March, 31, 12, 0, 0, 0, loc),
			billingDay: 31,
			wantStart:  time.Date(2026, time.February, 28, 0, 0, 0, 0, loc),
			wantEnd:    time.Date(2026, time.March, 31, 0, 0, 0, 0, loc),
		},
		{
			name:       "before march 31 uses previous closed period",
			now:        time.Date(2026, time.March, 15, 12, 0, 0, 0, loc),
			billingDay: 31,
			wantStart:  time.Date(2026, time.January, 31, 0, 0, 0, 0, loc),
			wantEnd:    time.Date(2026, time.February, 28, 0, 0, 0, 0, loc),
		},
		{
			name:       "leap year keeps february 29",
			now:        time.Date(2028, time.April, 1, 12, 0, 0, 0, loc),
			billingDay: 31,
			wantStart:  time.Date(2028, time.February, 29, 0, 0, 0, 0, loc),
			wantEnd:    time.Date(2028, time.March, 31, 0, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := billingPeriod(tt.now, tt.billingDay)
			if !gotStart.Equal(tt.wantStart) || !gotEnd.Equal(tt.wantEnd) {
				t.Fatalf("billingPeriod() = (%s, %s), want (%s, %s)", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
			if !gotEnd.After(gotStart) {
				t.Fatalf("billing period must be positive: start=%s end=%s", gotStart, gotEnd)
			}
		})
	}
}

func TestNormalizePagination(t *testing.T) {
	if page, size := normalizePagination(0, 0); page != 1 || size != 10 {
		t.Fatalf("normalizePagination(0, 0) = (%d, %d), want (1, 10)", page, size)
	}
	if page, size := normalizePagination(2, 25); page != 2 || size != 25 {
		t.Fatalf("normalizePagination(2, 25) = (%d, %d), want (2, 25)", page, size)
	}
}
