package handler

import (
	"context"

	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/service"
	"github.com/vblanchet22/back_coloc/internal/utils"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FundHandler implements the FundService gRPC server
type FundHandler struct {
	pb.UnimplementedFundServiceServer
	service *service.FundService
}

// NewFundHandler creates a new FundHandler
func NewFundHandler(service *service.FundService) *FundHandler {
	return &FundHandler{service: service}
}

// CreateFund creates a new fund
func (h *FundHandler) CreateFund(ctx context.Context, req *pb.CreateFundRequest) (*pb.Fund, error) {
	if req.ColocationId == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et name obligatoires")
	}

	fund, err := h.service.Create(ctx, req.ColocationId, req.Name, req.Description, req.TargetAmount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return fundToProto(fund), nil
}

// GetFund retrieves a fund by ID
func (h *FundHandler) GetFund(ctx context.Context, req *pb.GetFundRequest) (*pb.Fund, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	fund, err := h.service.GetByID(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return fundToProto(fund), nil
}

// ListFunds lists funds for a colocation
func (h *FundHandler) ListFunds(ctx context.Context, req *pb.ListFundsRequest) (*pb.ListFundsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	funds, err := h.service.List(ctx, req.ColocationId, req.IsActive)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbFunds []*pb.Fund
	for _, f := range funds {
		pbFunds = append(pbFunds, fundToProto(&f))
	}

	return &pb.ListFundsResponse{Funds: pbFunds}, nil
}

// UpdateFund updates a fund
func (h *FundHandler) UpdateFund(ctx context.Context, req *pb.UpdateFundRequest) (*pb.Fund, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	fund, err := h.service.Update(ctx, req.ColocationId, req.Id, req.Name, req.Description, req.TargetAmount, req.IsActive)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return fundToProto(fund), nil
}

// DeleteFund deletes a fund
func (h *FundHandler) DeleteFund(ctx context.Context, req *pb.DeleteFundRequest) (*pb.DeleteFundResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.Delete(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteFundResponse{Success: true}, nil
}

// AddContribution adds a contribution to a fund
func (h *FundHandler) AddContribution(ctx context.Context, req *pb.AddContributionRequest) (*pb.Contribution, error) {
	if req.ColocationId == "" || req.FundId == "" || req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, fund_id et amount obligatoires")
	}

	contribution, err := h.service.AddContribution(ctx, req.ColocationId, req.FundId, req.Amount, req.Note)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return contributionToProto(contribution), nil
}

// ListContributions lists contributions for a fund
func (h *FundHandler) ListContributions(ctx context.Context, req *pb.ListContributionsRequest) (*pb.ListContributionsResponse, error) {
	if req.ColocationId == "" || req.FundId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et fund_id obligatoires")
	}

	contributions, err := h.service.ListContributions(ctx, req.ColocationId, req.FundId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbContributions []*pb.Contribution
	for _, c := range contributions {
		pbContributions = append(pbContributions, contributionToProto(&c))
	}

	return &pb.ListContributionsResponse{Contributions: pbContributions}, nil
}

// DeleteContribution deletes a contribution
func (h *FundHandler) DeleteContribution(ctx context.Context, req *pb.DeleteContributionRequest) (*pb.DeleteContributionResponse, error) {
	if req.ColocationId == "" || req.FundId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, fund_id et id obligatoires")
	}

	if err := h.service.DeleteContribution(ctx, req.ColocationId, req.FundId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteContributionResponse{Success: true}, nil
}

// Helper functions

func fundToProto(f *domain.CommonFund) *pb.Fund {
	fund := &pb.Fund{
		Id:                 f.ID,
		ColocationId:       f.ColocationID,
		Name:               f.Name,
		Description:        f.Description,
		TargetAmount:       f.TargetAmount,
		CurrentAmount:      f.CurrentAmount,
		IsActive:           f.IsActive,
		CreatedBy:          f.CreatedBy,
		CreatedByNom:       f.CreatedByNom,
		CreatedByPrenom:    f.CreatedByPrenom,
		ProgressPercentage: f.ProgressPercentage,
		CreatedAt:          utils.FormatFrenchDateTime(f.CreatedAt),
	}

	for _, c := range f.Contributors {
		fund.Contributors = append(fund.Contributors, &pb.ContributorSummary{
			UserId:           c.UserID,
			UserNom:          c.UserNom,
			UserPrenom:       c.UserPrenom,
			TotalContributed: c.TotalContributed,
		})
	}

	return fund
}

func contributionToProto(c *domain.FundContribution) *pb.Contribution {
	return &pb.Contribution{
		Id:        c.ID,
		FundId:    c.FundID,
		UserId:    c.UserID,
		UserNom:   c.UserNom,
		UserPrenom: c.UserPrenom,
		Amount:    c.Amount,
		Note:      c.Note,
		CreatedAt: utils.FormatFrenchDateTime(c.CreatedAt),
	}
}
