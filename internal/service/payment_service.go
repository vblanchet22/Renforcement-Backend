package service

import (
	"context"
	"fmt"

	"github.com/vblanchet22/back_coloc/internal/auth"
	"github.com/vblanchet22/back_coloc/internal/domain"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

// PaymentService handles payment business logic
type PaymentService struct {
	repo           *postgres.PaymentRepository
	colocationRepo *postgres.ColocationRepository
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(repo *postgres.PaymentRepository, colocationRepo *postgres.ColocationRepository) *PaymentService {
	return &PaymentService{
		repo:           repo,
		colocationRepo: colocationRepo,
	}
}

// ensureMembership verifies user is a member and returns the userID
func (s *PaymentService) ensureMembership(ctx context.Context, colocationID string) (string, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return "", fmt.Errorf("vous n'etes pas membre de cette colocation")
	}

	return userID, nil
}

// Create creates a new payment (declare reimbursement)
func (s *PaymentService) Create(ctx context.Context, colocationID, toUserID string, amount float64, note *string) (*domain.Payment, error) {
	userID, err := s.ensureMembership(ctx, colocationID)
	if err != nil {
		return nil, err
	}

	if userID == toUserID {
		return nil, fmt.Errorf("vous ne pouvez pas vous payer vous-meme")
	}

	if err := s.verifyRecipientMembership(ctx, colocationID, toUserID); err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, fmt.Errorf("le montant doit etre positif")
	}

	payment := &domain.Payment{
		ColocationID: colocationID,
		FromUserID:   userID,
		ToUserID:     toUserID,
		Amount:       amount,
		Note:         note,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("erreur lors de la creation du paiement: %w", err)
	}

	return s.repo.GetByID(ctx, payment.ID)
}

// verifyRecipientMembership checks if the recipient is a member of the colocation
func (s *PaymentService) verifyRecipientMembership(ctx context.Context, colocationID, userID string) error {
	isMember, err := s.colocationRepo.IsMember(ctx, colocationID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la verification: %w", err)
	}
	if !isMember {
		return fmt.Errorf("le destinataire n'est pas membre de cette colocation")
	}
	return nil
}

// GetByID retrieves a payment by ID
func (s *PaymentService) GetByID(ctx context.Context, colocationID, paymentID string) (*domain.Payment, error) {
	if _, err := s.ensureMembership(ctx, colocationID); err != nil {
		return nil, err
	}

	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if payment == nil || payment.ColocationID != colocationID {
		return nil, fmt.Errorf("paiement introuvable")
	}

	return payment, nil
}

// ListPaymentsInput contains filters for listing payments
type ListPaymentsInput struct {
	ColocationID string
	Status       *string
	FromUserID   *string
	ToUserID     *string
	Page         int
	PageSize     int
}

// List lists payments for a colocation
func (s *PaymentService) List(ctx context.Context, input ListPaymentsInput) ([]domain.Payment, int, error) {
	if _, err := s.ensureMembership(ctx, input.ColocationID); err != nil {
		return nil, 0, err
	}

	input.Page, input.PageSize = normalizePagination(input.Page, input.PageSize)

	return s.repo.ListByColocation(ctx, input.ColocationID, input.Status, input.FromUserID, input.ToUserID, input.Page, input.PageSize)
}

// Confirm confirms a payment (only by recipient)
func (s *PaymentService) Confirm(ctx context.Context, colocationID, paymentID string) (*domain.Payment, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	payment, err := s.getPaymentForAction(ctx, colocationID, paymentID)
	if err != nil {
		return nil, err
	}

	if payment.ToUserID != userID {
		return nil, fmt.Errorf("seul le destinataire peut confirmer ce paiement")
	}

	if payment.Status != domain.PaymentStatusPending {
		return nil, fmt.Errorf("ce paiement n'est pas en attente")
	}

	if err := s.repo.UpdateStatus(ctx, paymentID, string(domain.PaymentStatusConfirmed)); err != nil {
		return nil, fmt.Errorf("erreur lors de la confirmation: %w", err)
	}

	return s.repo.GetByID(ctx, paymentID)
}

// Reject rejects a payment (only by recipient)
func (s *PaymentService) Reject(ctx context.Context, colocationID, paymentID string) (*domain.Payment, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	payment, err := s.getPaymentForAction(ctx, colocationID, paymentID)
	if err != nil {
		return nil, err
	}

	if payment.ToUserID != userID {
		return nil, fmt.Errorf("seul le destinataire peut rejeter ce paiement")
	}

	if payment.Status != domain.PaymentStatusPending {
		return nil, fmt.Errorf("ce paiement n'est pas en attente")
	}

	if err := s.repo.UpdateStatus(ctx, paymentID, string(domain.PaymentStatusRejected)); err != nil {
		return nil, fmt.Errorf("erreur lors du rejet: %w", err)
	}

	return s.repo.GetByID(ctx, paymentID)
}

// Cancel cancels a pending payment (only by sender)
func (s *PaymentService) Cancel(ctx context.Context, colocationID, paymentID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	payment, err := s.getPaymentForAction(ctx, colocationID, paymentID)
	if err != nil {
		return err
	}

	if payment.FromUserID != userID {
		return fmt.Errorf("seul l'emetteur peut annuler ce paiement")
	}

	if payment.Status != domain.PaymentStatusPending {
		return fmt.Errorf("ce paiement n'est pas en attente")
	}

	return s.repo.Delete(ctx, paymentID)
}

// getPaymentForAction retrieves a payment and validates it belongs to the colocation
func (s *PaymentService) getPaymentForAction(ctx context.Context, colocationID, paymentID string) (*domain.Payment, error) {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}
	if payment == nil || payment.ColocationID != colocationID {
		return nil, fmt.Errorf("paiement introuvable")
	}
	return payment, nil
}
