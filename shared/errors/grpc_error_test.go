package errors

import (
	"errors"
	"net/http"
	"testing"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseGrpcError_NotFoundWithDetail(t *testing.T) {
	// Round-trip through ToGrpcError, exactly like a microservice emits a
	// not-found AppError to the API Gateway.
	err := ToGrpcError(ErrNotFoundResponse("merchant document"))

	got := ParseGrpcError(err)
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want %d", got.Code, http.StatusNotFound)
	}
	if got.Type != ErrorTypeNotFound {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeNotFound)
	}
	if got.Message != "merchant document not found" {
		t.Errorf("Message = %q, want %q", got.Message, "merchant document not found")
	}
}

func TestParseGrpcError_NotFoundFallbackWithoutDetail(t *testing.T) {
	// A plain gRPC error without any pb.ErrorResponse detail must still map
	// codes.NotFound to HTTP 404.
	err := status.Error(codes.NotFound, "merchant document not found")

	got := ParseGrpcError(err)
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want %d", got.Code, http.StatusNotFound)
	}
	if got.Type != ErrorTypeNotFound {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeNotFound)
	}
}

func TestParseGrpcError_DetailCodeZeroFallsBackToGrpcStatus(t *testing.T) {
	// A service that attaches an ErrorResponse detail but forgets to set the
	// HTTP code must NOT produce AppError.Code == 0 (which the API gateway
	// would turn into a bogus 500/200 response). It must fall back to the
	// gRPC status code (404 for codes.NotFound).
	st := status.New(codes.NotFound, "merchant document not found")
	stWithDetail, err := st.WithDetails(&pb.ErrorResponse{
		Status:  "NOT_FOUND",
		Message: "merchant document not found",
		Code:    0,
	})
	if err != nil {
		t.Fatalf("failed to attach detail: %v", err)
	}

	got := ParseGrpcError(stWithDetail.Err())
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want %d (fallback to codes.NotFound)", got.Code, http.StatusNotFound)
	}
	if got.Code == 0 {
		t.Fatal("ParseGrpcError returned Code 0; an HTTP handler would emit a non-500/invalid status")
	}
}

func TestParseGrpcError_DetailValidCodeWins(t *testing.T) {
	// When the detail carries a valid HTTP code, it is preserved even though
	// the gRPC status code is different (already exists / conflict).
	st := status.New(codes.AlreadyExists, "duplicate entry")
	stWithDetail, err := st.WithDetails(&pb.ErrorResponse{
		Status:  "CONFLICT",
		Message: "duplicate entry",
		Code:    http.StatusConflict,
	})
	if err != nil {
		t.Fatalf("failed to attach detail: %v", err)
	}

	got := ParseGrpcError(stWithDetail.Err())
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusConflict {
		t.Errorf("Code = %d, want %d", got.Code, http.StatusConflict)
	}
	if got.Type != ErrorTypeConflict {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeConflict)
	}
}

func TestParseGrpcError_NilError(t *testing.T) {
	if got := ParseGrpcError(nil); got != nil {
		t.Errorf("ParseGrpcError(nil) = %v, want nil", got)
	}
}

func TestParseGrpcError_NonGrpcError(t *testing.T) {
	got := ParseGrpcError(errors.New("boom"))
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a non-gRPC error")
	}
	if got.Code != http.StatusInternalServerError {
		t.Errorf("Code = %d, want %d", got.Code, http.StatusInternalServerError)
	}
	if got.Type != ErrorTypeInternal {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeInternal)
	}
}

func TestParseGrpcError_Unavailable(t *testing.T) {
	err := status.Error(codes.Unavailable, "service unavailable")

	got := ParseGrpcError(err)
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusServiceUnavailable {
		t.Errorf("Code = %d, want %d", got.Code, http.StatusServiceUnavailable)
	}
	if got.Type != ErrorTypeUnavailable {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeUnavailable)
	}
}

func TestParseGrpcError_UnavailableRoundTrip(t *testing.T) {
	// A 503 AppError sent through ToGrpcError must survive the round-trip
	// (httpToGrpcCode(503) -> codes.Unavailable -> 503) and not degrade to 500.
	err := ToGrpcError(ErrServiceUnavailable.WithMessage("stats service is temporarily unavailable"))

	got := ParseGrpcError(err)
	if got == nil {
		t.Fatal("ParseGrpcError returned nil for a gRPC error")
	}
	if got.Code != http.StatusServiceUnavailable {
		t.Errorf("Code = %d, want %d (round-trip must preserve 503)", got.Code, http.StatusServiceUnavailable)
	}
	if got.Type != ErrorTypeUnavailable {
		t.Errorf("Type = %q, want %q", got.Type, ErrorTypeUnavailable)
	}
}
