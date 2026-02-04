package handler

import (
	"context"
	"time"

	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/service"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CategoryHandler implements the CategoryService gRPC server
type CategoryHandler struct {
	pb.UnimplementedCategoryServiceServer
	service *service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// ListCategories lists all categories for a colocation
func (h *CategoryHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	categories, err := h.service.List(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbCategories []*pb.Category
	for _, c := range categories {
		pbCategories = append(pbCategories, categoryToProto(&c))
	}

	return &pb.ListCategoriesResponse{Categories: pbCategories}, nil
}

// CreateCategory creates a new custom category
func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	if req.ColocationId == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et name obligatoires")
	}

	category, err := h.service.Create(ctx, req.ColocationId, req.Name, req.Icon, req.Color)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return categoryToProto(category), nil
}

// UpdateCategory updates a custom category
func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	category, err := h.service.Update(ctx, req.ColocationId, req.Id, req.Name, req.Icon, req.Color)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return categoryToProto(category), nil
}

// DeleteCategory deletes a custom category
func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.Delete(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteCategoryResponse{Success: true}, nil
}

// GetCategoryStats returns category statistics
func (h *CategoryHandler) GetCategoryStats(ctx context.Context, req *pb.GetCategoryStatsRequest) (*pb.GetCategoryStatsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format de date invalide pour start_date")
		}
		startDate = &t
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format de date invalide pour end_date")
		}
		endDate = &t
	}

	stats, totalAmount, err := h.service.GetStats(ctx, req.ColocationId, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbStats []*pb.CategoryStat
	for _, s := range stats {
		pbStats = append(pbStats, categoryStatToProto(&s))
	}

	return &pb.GetCategoryStatsResponse{
		Stats:       pbStats,
		TotalAmount: totalAmount,
	}, nil
}

// Helper functions

func categoryToProto(c *domain.ExpenseCategory) *pb.Category {
	cat := &pb.Category{
		Id:       c.ID,
		Name:     c.Name,
		Icon:     c.Icon,
		Color:    c.Color,
		IsGlobal: c.IsGlobal(),
	}
	if c.ColocationID != nil {
		cat.ColocationId = c.ColocationID
	}
	return cat
}

func categoryStatToProto(s *domain.CategoryStat) *pb.CategoryStat {
	return &pb.CategoryStat{
		CategoryId:   s.CategoryID,
		CategoryName: s.CategoryName,
		Icon:         s.Icon,
		Color:        s.Color,
		TotalAmount:  s.TotalAmount,
		ExpenseCount: int32(s.ExpenseCount),
		Percentage:   s.Percentage,
	}
}
