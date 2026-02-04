package handler

import (
	"context"
	"time"

	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/service"
	"github.com/vblanchet22/back_coloc/internal/utils"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DecisionHandler implements the DecisionService gRPC server
type DecisionHandler struct {
	pb.UnimplementedDecisionServiceServer
	service *service.DecisionService
}

// NewDecisionHandler creates a new DecisionHandler
func NewDecisionHandler(service *service.DecisionService) *DecisionHandler {
	return &DecisionHandler{service: service}
}

// CreateDecision creates a new decision
func (h *DecisionHandler) CreateDecision(ctx context.Context, req *pb.CreateDecisionRequest) (*pb.Decision, error) {
	if req.ColocationId == "" || req.Title == "" || len(req.Options) < 2 {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, title et au moins 2 options obligatoires")
	}

	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		t, err := time.Parse("2006-01-02 15:04", *req.Deadline)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format deadline invalide (attendu: YYYY-MM-DD HH:MM)")
		}
		deadline = &t
	}

	decision, err := h.service.Create(ctx, req.ColocationId, req.Title, req.Description, req.Options, deadline, req.AllowMultiple, req.IsAnonymous)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return decisionToProto(decision), nil
}

// GetDecision retrieves a decision by ID
func (h *DecisionHandler) GetDecision(ctx context.Context, req *pb.GetDecisionRequest) (*pb.Decision, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	decision, err := h.service.GetByID(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return decisionToProto(decision), nil
}

// ListDecisions lists decisions for a colocation
func (h *DecisionHandler) ListDecisions(ctx context.Context, req *pb.ListDecisionsRequest) (*pb.ListDecisionsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	var statusFilter *string
	if req.Status != nil && *req.Status != pb.DecisionStatus_DECISION_STATUS_UNSPECIFIED {
		s := protoDecisionStatusToDomain(*req.Status)
		str := string(s)
		statusFilter = &str
	}

	page := int32(1)
	pageSize := int32(20)
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = *req.PageSize
	}

	decisions, totalCount, err := h.service.List(ctx, req.ColocationId, statusFilter, int(page), int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbDecisions []*pb.Decision
	for _, d := range decisions {
		pbDecisions = append(pbDecisions, decisionToProto(&d))
	}

	return &pb.ListDecisionsResponse{
		Decisions:  pbDecisions,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// UpdateDecision updates a decision
func (h *DecisionHandler) UpdateDecision(ctx context.Context, req *pb.UpdateDecisionRequest) (*pb.Decision, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		t, err := time.Parse("2006-01-02 15:04", *req.Deadline)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format deadline invalide")
		}
		deadline = &t
	}

	decision, err := h.service.Update(ctx, req.ColocationId, req.Id, req.Title, req.Description, req.Options, deadline, req.AllowMultiple, req.IsAnonymous)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return decisionToProto(decision), nil
}

// DeleteDecision deletes a decision
func (h *DecisionHandler) DeleteDecision(ctx context.Context, req *pb.DeleteDecisionRequest) (*pb.DeleteDecisionResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.Delete(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteDecisionResponse{Success: true}, nil
}

// Vote votes on a decision
func (h *DecisionHandler) Vote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	if req.ColocationId == "" || req.DecisionId == "" || len(req.OptionIndices) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, decision_id et option_indices obligatoires")
	}

	var indices []int
	for _, idx := range req.OptionIndices {
		indices = append(indices, int(idx))
	}

	if err := h.service.Vote(ctx, req.ColocationId, req.DecisionId, indices); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.VoteResponse{Success: true}, nil
}

// CloseDecision closes a decision
func (h *DecisionHandler) CloseDecision(ctx context.Context, req *pb.CloseDecisionRequest) (*pb.Decision, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	decision, err := h.service.Close(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return decisionToProto(decision), nil
}

// GetResults returns the results of a decision
func (h *DecisionHandler) GetResults(ctx context.Context, req *pb.GetResultsRequest) (*pb.GetResultsResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	results, totalVotes, totalVoters, winningIdx, err := h.service.GetResults(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbResults []*pb.OptionResult
	for _, r := range results {
		pbResult := &pb.OptionResult{
			OptionIndex: int32(r.OptionIndex),
			OptionText:  r.OptionText,
			VoteCount:   int32(r.VoteCount),
			Percentage:  r.Percentage,
		}

		for _, v := range r.Voters {
			pbResult.Voters = append(pbResult.Voters, &pb.Voter{
				UserId:     v.UserID,
				UserNom:    v.UserNom,
				UserPrenom: v.UserPrenom,
			})
		}

		pbResults = append(pbResults, pbResult)
	}

	resp := &pb.GetResultsResponse{
		DecisionId:  req.Id,
		Status:      pb.DecisionStatus_DECISION_STATUS_OPEN,
		Results:     pbResults,
		TotalVotes:  int32(totalVotes),
		TotalVoters: int32(totalVoters),
	}

	if winningIdx != nil {
		idx := int32(*winningIdx)
		resp.WinningOptionIndex = &idx
	}

	return resp, nil
}

// Helper functions

func decisionToProto(d *domain.Decision) *pb.Decision {
	decision := &pb.Decision{
		Id:              d.ID,
		ColocationId:    d.ColocationID,
		CreatedBy:       d.CreatedBy,
		CreatedByNom:    d.CreatedByNom,
		CreatedByPrenom: d.CreatedByPrenom,
		Title:           d.Title,
		Description:     d.Description,
		Status:          domainDecisionStatusToProto(d.Status),
		AllowMultiple:   d.AllowMultiple,
		IsAnonymous:     d.IsAnonymous,
		VoteCount:       int32(d.VoteCount),
		HasVoted:        d.HasVoted,
		CreatedAt:       utils.FormatFrenchDateTime(d.CreatedAt),
	}

	for i, opt := range d.Options {
		decision.Options = append(decision.Options, &pb.DecisionOption{
			Index: int32(i),
			Text:  opt,
		})
	}

	if d.Deadline != nil {
		dl := d.Deadline.Format("2006-01-02 15:04")
		decision.Deadline = &dl
	}

	for _, v := range d.UserVotes {
		decision.UserVotes = append(decision.UserVotes, int32(v))
	}

	return decision
}

func domainDecisionStatusToProto(s domain.DecisionStatus) pb.DecisionStatus {
	switch s {
	case domain.DecisionStatusOpen:
		return pb.DecisionStatus_DECISION_STATUS_OPEN
	case domain.DecisionStatusClosed:
		return pb.DecisionStatus_DECISION_STATUS_CLOSED
	default:
		return pb.DecisionStatus_DECISION_STATUS_UNSPECIFIED
	}
}

func protoDecisionStatusToDomain(s pb.DecisionStatus) domain.DecisionStatus {
	switch s {
	case pb.DecisionStatus_DECISION_STATUS_OPEN:
		return domain.DecisionStatusOpen
	case pb.DecisionStatus_DECISION_STATUS_CLOSED:
		return domain.DecisionStatusClosed
	default:
		return domain.DecisionStatusOpen
	}
}
