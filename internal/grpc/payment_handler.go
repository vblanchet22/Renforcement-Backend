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

// PaymentHandler implements the PaymentService gRPC server
type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// CreatePayment creates a new payment
func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.Payment, error) {
	if req.ColocationId == "" || req.ToUserId == "" || req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, to_user_id et amount sont obligatoires")
	}

	payment, err := h.service.Create(ctx, req.ColocationId, req.ToUserId, req.Amount, req.Note)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return paymentToProto(payment), nil
}

// GetPayment retrieves a payment by ID
func (h *PaymentHandler) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	payment, err := h.service.GetByID(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return paymentToProto(payment), nil
}

// ListPayments lists payments for a colocation
func (h *PaymentHandler) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	var statusFilter *string
	if req.Status != nil && *req.Status != pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		s := protoPaymentStatusToDomain(*req.Status)
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

	payments, totalCount, err := h.service.List(ctx, service.ListPaymentsInput{
		ColocationID: req.ColocationId,
		Status:       statusFilter,
		FromUserID:   req.FromUserId,
		ToUserID:     req.ToUserId,
		Page:         int(page),
		PageSize:     int(pageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbPayments []*pb.Payment
	for _, p := range payments {
		pbPayments = append(pbPayments, paymentToProto(&p))
	}

	return &pb.ListPaymentsResponse{
		Payments:   pbPayments,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// ConfirmPayment confirms a payment
func (h *PaymentHandler) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.Payment, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	payment, err := h.service.Confirm(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return paymentToProto(payment), nil
}

// RejectPayment rejects a payment
func (h *PaymentHandler) RejectPayment(ctx context.Context, req *pb.RejectPaymentRequest) (*pb.Payment, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	payment, err := h.service.Reject(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return paymentToProto(payment), nil
}

// CancelPayment cancels a pending payment
func (h *PaymentHandler) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.Cancel(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.CancelPaymentResponse{Success: true}, nil
}

// Helper functions

func paymentToProto(p *domain.Payment) *pb.Payment {
	payment := &pb.Payment{
		Id:             p.ID,
		ColocationId:   p.ColocationID,
		FromUserId:     p.FromUserID,
		FromUserNom:    p.FromUserNom,
		FromUserPrenom: p.FromUserPrenom,
		FromAvatarUrl:  p.FromAvatarURL,
		ToUserId:       p.ToUserID,
		ToUserNom:      p.ToUserNom,
		ToUserPrenom:   p.ToUserPrenom,
		ToAvatarUrl:    p.ToAvatarURL,
		Amount:         p.Amount,
		Status:         domainPaymentStatusToProto(p.Status),
		Note:           p.Note,
		CreatedAt:      utils.FormatFrenchDateTime(p.CreatedAt),
	}

	if p.ConfirmedAt != nil {
		confirmedAt := utils.FormatFrenchDateTime(*p.ConfirmedAt)
		payment.ConfirmedAt = &confirmedAt
	}

	return payment
}

func domainPaymentStatusToProto(s domain.PaymentStatus) pb.PaymentStatus {
	switch s {
	case domain.PaymentStatusPending:
		return pb.PaymentStatus_PAYMENT_STATUS_PENDING
	case domain.PaymentStatusConfirmed:
		return pb.PaymentStatus_PAYMENT_STATUS_CONFIRMED
	case domain.PaymentStatusRejected:
		return pb.PaymentStatus_PAYMENT_STATUS_REJECTED
	default:
		return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func protoPaymentStatusToDomain(s pb.PaymentStatus) domain.PaymentStatus {
	switch s {
	case pb.PaymentStatus_PAYMENT_STATUS_PENDING:
		return domain.PaymentStatusPending
	case pb.PaymentStatus_PAYMENT_STATUS_CONFIRMED:
		return domain.PaymentStatusConfirmed
	case pb.PaymentStatus_PAYMENT_STATUS_REJECTED:
		return domain.PaymentStatusRejected
	default:
		return domain.PaymentStatusPending
	}
}
