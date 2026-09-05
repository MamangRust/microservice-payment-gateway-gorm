package apimapper

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/response"
)

// mapPaginationMeta maps a gRPC PaginationMeta to an HTTP-compatible API response PaginationMeta.
//
// Args:
//   - s: A pointer to a pb.PaginationMeta containing the gRPC response data.
//
// Returns:
//   - A pointer to a response.PaginationMeta containing the mapped data, including current page, page size,
//     total records, and total pages.
func MapPaginationMeta(s *pb.PaginationMeta) *response.PaginationMeta {
	if s == nil {
		return nil
	}
	return &response.PaginationMeta{
		CurrentPage:  int(s.CurrentPage),
		PageSize:     int(s.PageSize),
		TotalRecords: int(s.TotalRecords),
		TotalPages:   int(s.TotalPages),
	}
}
