package handler

import (
	"context"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/service"
	"github.com/vblanchet22/back_coloc/internal/utils"
	pb "github.com/vblanchet22/back_coloc/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NotificationHandler implements the NotificationService gRPC server
type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	service *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// ListNotifications lists notifications for the current user
func (h *NotificationHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	unreadOnly := false
	if req.UnreadOnly != nil {
		unreadOnly = *req.UnreadOnly
	}

	page := int32(1)
	pageSize := int32(20)
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = *req.PageSize
	}

	notifications, totalCount, unreadCount, err := h.service.List(ctx, req.ColocationId, unreadOnly, int(page), int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbNotifications []*pb.Notification
	for _, n := range notifications {
		pbNotifications = append(pbNotifications, notificationToProto(&n))
	}

	return &pb.ListNotificationsResponse{
		Notifications: pbNotifications,
		TotalCount:    int32(totalCount),
		UnreadCount:   int32(unreadCount),
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	if err := h.service.MarkAsRead(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.MarkAsReadResponse{Success: true}, nil
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(ctx context.Context, req *pb.MarkAllAsReadRequest) (*pb.MarkAllAsReadResponse, error) {
	count, err := h.service.MarkAllAsRead(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.MarkAllAsReadResponse{MarkedCount: int32(count)}, nil
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*pb.DeleteNotificationResponse, error) {
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id obligatoire")
	}

	if err := h.service.Delete(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteNotificationResponse{Success: true}, nil
}

// GetUnreadCount returns the unread notification count
func (h *NotificationHandler) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := h.service.GetUnreadCount(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.GetUnreadCountResponse{Count: int32(count)}, nil
}

// StreamNotifications streams real-time notifications to the client
func (h *NotificationHandler) StreamNotifications(req *pb.StreamNotificationsRequest, stream pb.NotificationService_StreamNotificationsServer) error {
	ctx := stream.Context()

	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "authentification requise")
	}

	// Subscribe to notifications
	ch := h.service.Subscribe(userID)
	defer h.service.Unsubscribe(userID, ch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case notif, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(notificationToProto(notif)); err != nil {
				return err
			}
		}
	}
}

// Helper functions

func notificationToProto(n *domain.Notification) *pb.Notification {
	notif := &pb.Notification{
		Id:             n.ID,
		UserId:         n.UserID,
		ColocationId:   n.ColocationID,
		ColocationName: n.ColocationName,
		Type:           domainNotifTypeToProto(n.Type),
		Title:          n.Title,
		Body:           n.Body,
		IsRead:         n.IsRead,
		CreatedAt:      utils.FormatFrenchDateTime(n.CreatedAt),
	}

	if n.Data != nil {
		notif.Data = n.Data
	}

	return notif
}

func domainNotifTypeToProto(t domain.NotificationType) pb.NotificationType {
	switch t {
	case domain.NotifExpenseCreated:
		return pb.NotificationType_NOTIFICATION_TYPE_EXPENSE_CREATED
	case domain.NotifExpenseUpdated:
		return pb.NotificationType_NOTIFICATION_TYPE_EXPENSE_UPDATED
	case domain.NotifExpenseDeleted:
		return pb.NotificationType_NOTIFICATION_TYPE_EXPENSE_DELETED
	case domain.NotifPaymentReceived:
		return pb.NotificationType_NOTIFICATION_TYPE_PAYMENT_RECEIVED
	case domain.NotifPaymentConfirmed:
		return pb.NotificationType_NOTIFICATION_TYPE_PAYMENT_CONFIRMED
	case domain.NotifPaymentRejected:
		return pb.NotificationType_NOTIFICATION_TYPE_PAYMENT_REJECTED
	case domain.NotifMemberJoined:
		return pb.NotificationType_NOTIFICATION_TYPE_MEMBER_JOINED
	case domain.NotifMemberLeft:
		return pb.NotificationType_NOTIFICATION_TYPE_MEMBER_LEFT
	case domain.NotifMemberRemoved:
		return pb.NotificationType_NOTIFICATION_TYPE_MEMBER_REMOVED
	case domain.NotifInvitationReceived:
		return pb.NotificationType_NOTIFICATION_TYPE_INVITATION_RECEIVED
	case domain.NotifRoleChanged:
		return pb.NotificationType_NOTIFICATION_TYPE_ROLE_CHANGED
	case domain.NotifDecisionCreated:
		return pb.NotificationType_NOTIFICATION_TYPE_DECISION_CREATED
	case domain.NotifDecisionClosed:
		return pb.NotificationType_NOTIFICATION_TYPE_DECISION_CLOSED
	case domain.NotifDecisionDeadline:
		return pb.NotificationType_NOTIFICATION_TYPE_DECISION_DEADLINE
	case domain.NotifFundCreated:
		return pb.NotificationType_NOTIFICATION_TYPE_FUND_CREATED
	case domain.NotifFundContribution:
		return pb.NotificationType_NOTIFICATION_TYPE_FUND_CONTRIBUTION
	case domain.NotifFundGoalReached:
		return pb.NotificationType_NOTIFICATION_TYPE_FUND_GOAL_REACHED
	case domain.NotifEventCreated:
		return pb.NotificationType_NOTIFICATION_TYPE_EVENT_CREATED
	case domain.NotifEventUpdated:
		return pb.NotificationType_NOTIFICATION_TYPE_EVENT_UPDATED
	case domain.NotifEventReminder:
		return pb.NotificationType_NOTIFICATION_TYPE_EVENT_REMINDER
	case domain.NotifEventCancelled:
		return pb.NotificationType_NOTIFICATION_TYPE_EVENT_CANCELLED
	case domain.NotifRecurringDue:
		return pb.NotificationType_NOTIFICATION_TYPE_RECURRING_DUE
	default:
		return pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}
