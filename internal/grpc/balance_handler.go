package handler

import (
	"context"
	"time"

	"github.com/vblanchet22/back_coloc/internal/service"
	"github.com/vblanchet22/back_coloc/internal/utils"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BalanceHandler implements the BalanceService gRPC server
type BalanceHandler struct {
	pb.UnimplementedBalanceServiceServer
	service *service.BalanceService
}

// NewBalanceHandler creates a new BalanceHandler
func NewBalanceHandler(service *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{service: service}
}

// GetBalances returns all balances and debts for a colocation
func (h *BalanceHandler) GetBalances(ctx context.Context, req *pb.GetBalancesRequest) (*pb.GetBalancesResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	balances, debts, err := h.service.GetBalances(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbBalances []*pb.UserBalance
	for _, b := range balances {
		pbBalances = append(pbBalances, &pb.UserBalance{
			UserId:     b.UserID,
			UserNom:    b.UserNom,
			UserPrenom: b.UserPrenom,
			AvatarUrl:  b.AvatarURL,
			TotalPaid:  b.TotalPaid,
			TotalOwed:  b.TotalOwed,
			NetBalance: b.NetBalance,
		})
	}

	var pbDebts []*pb.Debt
	for _, d := range debts {
		pbDebts = append(pbDebts, &pb.Debt{
			FromUserId:     d.FromUserID,
			FromUserNom:    d.FromUserNom,
			FromUserPrenom: d.FromUserPrenom,
			ToUserId:       d.ToUserID,
			ToUserNom:      d.ToUserNom,
			ToUserPrenom:   d.ToUserPrenom,
			Amount:         d.Amount,
		})
	}

	return &pb.GetBalancesResponse{
		Balances: pbBalances,
		Debts:    pbDebts,
	}, nil
}

// GetSimplifiedDebts returns simplified debts using min-cash-flow algorithm
func (h *BalanceHandler) GetSimplifiedDebts(ctx context.Context, req *pb.GetSimplifiedDebtsRequest) (*pb.GetSimplifiedDebtsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	debts, err := h.service.GetSimplifiedDebts(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbDebts []*pb.SimplifiedDebt
	for _, d := range debts {
		pbDebts = append(pbDebts, &pb.SimplifiedDebt{
			FromUserId:     d.FromUserID,
			FromUserNom:    d.FromUserNom,
			FromUserPrenom: d.FromUserPrenom,
			FromAvatarUrl:  d.FromAvatarURL,
			ToUserId:       d.ToUserID,
			ToUserNom:      d.ToUserNom,
			ToUserPrenom:   d.ToUserPrenom,
			ToAvatarUrl:    d.ToAvatarURL,
			Amount:         d.Amount,
		})
	}

	return &pb.GetSimplifiedDebtsResponse{Debts: pbDebts}, nil
}

// GetBalanceHistory returns balance history for the current user
func (h *BalanceHandler) GetBalanceHistory(ctx context.Context, req *pb.GetBalanceHistoryRequest) (*pb.GetBalanceHistoryResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format start_date invalide")
		}
		startDate = &t
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format end_date invalide")
		}
		endDate = &t
	}

	entries, err := h.service.GetBalanceHistory(ctx, req.ColocationId, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbEntries []*pb.BalanceHistoryEntry
	for _, e := range entries {
		pbEntries = append(pbEntries, &pb.BalanceHistoryEntry{
			Date:              utils.FormatFrenchDate(e.Date),
			CumulativeBalance: e.CumulativeBalance,
			EventType:         e.EventType,
			EventId:           e.EventID,
			Description:       e.Description,
			Amount:            e.Amount,
		})
	}

	return &pb.GetBalanceHistoryResponse{Entries: pbEntries}, nil
}
