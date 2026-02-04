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

// ExpenseHandler implements the ExpenseService gRPC server
type ExpenseHandler struct {
	pb.UnimplementedExpenseServiceServer
	service *service.ExpenseService
}

// NewExpenseHandler creates a new ExpenseHandler
func NewExpenseHandler(service *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

// CreateExpense creates a new expense
func (h *ExpenseHandler) CreateExpense(ctx context.Context, req *pb.CreateExpenseRequest) (*pb.Expense, error) {
	if req.ColocationId == "" || req.Title == "" || req.Amount <= 0 || req.CategoryId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, title, amount et category_id sont obligatoires")
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "format de date invalide (attendu: YYYY-MM-DD)")
	}

	var splits []domain.ExpenseSplitInput
	for _, s := range req.Splits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     s.UserId,
			Amount:     s.Amount,
			Percentage: s.Percentage,
		})
	}

	expense, err := h.service.Create(ctx, service.CreateExpenseInput{
		ColocationID: req.ColocationId,
		Title:        req.Title,
		Description:  req.Description,
		Amount:       req.Amount,
		CategoryID:   req.CategoryId,
		SplitType:    protoSplitTypeToDomain(req.SplitType),
		Splits:       splits,
		ExpenseDate:  expenseDate,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return expenseToProto(expense), nil
}

// GetExpense retrieves an expense by ID
func (h *ExpenseHandler) GetExpense(ctx context.Context, req *pb.GetExpenseRequest) (*pb.Expense, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	expense, err := h.service.GetByID(ctx, req.ColocationId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	return expenseToProto(expense), nil
}

// ListExpenses lists expenses for a colocation
func (h *ExpenseHandler) ListExpenses(ctx context.Context, req *pb.ListExpensesRequest) (*pb.ListExpensesResponse, error) {
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

	page := int32(1)
	pageSize := int32(20)
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = *req.PageSize
	}

	expenses, totalCount, err := h.service.List(ctx, service.ListExpensesInput{
		ColocationID: req.ColocationId,
		CategoryID:   req.CategoryId,
		PaidBy:       req.PaidBy,
		StartDate:    startDate,
		EndDate:      endDate,
		Page:         int(page),
		PageSize:     int(pageSize),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbExpenses []*pb.Expense
	for _, e := range expenses {
		pbExpenses = append(pbExpenses, expenseToProto(&e))
	}

	return &pb.ListExpensesResponse{
		Expenses:   pbExpenses,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// UpdateExpense updates an expense
func (h *ExpenseHandler) UpdateExpense(ctx context.Context, req *pb.UpdateExpenseRequest) (*pb.Expense, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	var expenseDate *time.Time
	if req.ExpenseDate != nil && *req.ExpenseDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpenseDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format de date invalide")
		}
		expenseDate = &t
	}

	var splitType *domain.SplitType
	if req.SplitType != nil {
		st := protoSplitTypeToDomain(*req.SplitType)
		splitType = &st
	}

	var splits []domain.ExpenseSplitInput
	for _, s := range req.Splits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     s.UserId,
			Amount:     s.Amount,
			Percentage: s.Percentage,
		})
	}

	expense, err := h.service.Update(ctx, service.UpdateExpenseInput{
		ColocationID: req.ColocationId,
		ExpenseID:    req.Id,
		Title:        req.Title,
		Description:  req.Description,
		Amount:       req.Amount,
		CategoryID:   req.CategoryId,
		SplitType:    splitType,
		Splits:       splits,
		ExpenseDate:  expenseDate,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return expenseToProto(expense), nil
}

// DeleteExpense deletes an expense
func (h *ExpenseHandler) DeleteExpense(ctx context.Context, req *pb.DeleteExpenseRequest) (*pb.DeleteExpenseResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.Delete(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteExpenseResponse{Success: true}, nil
}

// CreateRecurringExpense creates a recurring expense template
func (h *ExpenseHandler) CreateRecurringExpense(ctx context.Context, req *pb.CreateRecurringExpenseRequest) (*pb.RecurringExpense, error) {
	if req.ColocationId == "" || req.Title == "" || req.Amount <= 0 || req.CategoryId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id, title, amount et category_id sont obligatoires")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "format start_date invalide")
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format end_date invalide")
		}
		endDate = &t
	}

	var splits []domain.ExpenseSplitInput
	for _, s := range req.Splits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     s.UserId,
			Amount:     s.Amount,
			Percentage: s.Percentage,
		})
	}

	recurring, err := h.service.CreateRecurring(ctx, service.CreateRecurringInput{
		ColocationID: req.ColocationId,
		Title:        req.Title,
		Description:  req.Description,
		Amount:       req.Amount,
		CategoryID:   req.CategoryId,
		SplitType:    protoSplitTypeToDomain(req.SplitType),
		Splits:       splits,
		Recurrence:   protoRecurrenceToDomain(req.Recurrence),
		StartDate:    startDate,
		EndDate:      endDate,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return recurringExpenseToProto(recurring), nil
}

// ListRecurringExpenses lists recurring expenses
func (h *ExpenseHandler) ListRecurringExpenses(ctx context.Context, req *pb.ListRecurringExpensesRequest) (*pb.ListRecurringExpensesResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	recurrings, err := h.service.ListRecurring(ctx, req.ColocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbRecurrings []*pb.RecurringExpense
	for _, re := range recurrings {
		pbRecurrings = append(pbRecurrings, recurringExpenseToProto(&re))
	}

	return &pb.ListRecurringExpensesResponse{RecurringExpenses: pbRecurrings}, nil
}

// UpdateRecurringExpense updates a recurring expense
func (h *ExpenseHandler) UpdateRecurringExpense(ctx context.Context, req *pb.UpdateRecurringExpenseRequest) (*pb.RecurringExpense, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "format end_date invalide")
		}
		endDate = &t
	}

	var splitType *domain.SplitType
	if req.SplitType != nil {
		st := protoSplitTypeToDomain(*req.SplitType)
		splitType = &st
	}

	var recurrence *domain.Recurrence
	if req.Recurrence != nil {
		r := protoRecurrenceToDomain(*req.Recurrence)
		recurrence = &r
	}

	var splits []domain.ExpenseSplitInput
	for _, s := range req.Splits {
		splits = append(splits, domain.ExpenseSplitInput{
			UserID:     s.UserId,
			Amount:     s.Amount,
			Percentage: s.Percentage,
		})
	}

	recurring, err := h.service.UpdateRecurring(ctx, service.UpdateRecurringInput{
		ColocationID: req.ColocationId,
		RecurringID:  req.Id,
		Title:        req.Title,
		Description:  req.Description,
		Amount:       req.Amount,
		CategoryID:   req.CategoryId,
		SplitType:    splitType,
		Splits:       splits,
		Recurrence:   recurrence,
		EndDate:      endDate,
		IsActive:     req.IsActive,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return recurringExpenseToProto(recurring), nil
}

// DeleteRecurringExpense deletes a recurring expense
func (h *ExpenseHandler) DeleteRecurringExpense(ctx context.Context, req *pb.DeleteRecurringExpenseRequest) (*pb.DeleteRecurringExpenseResponse, error) {
	if req.ColocationId == "" || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id et id obligatoires")
	}

	if err := h.service.DeleteRecurring(ctx, req.ColocationId, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.DeleteRecurringExpenseResponse{Success: true}, nil
}

// GetForecast returns expense forecast
func (h *ExpenseHandler) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	if req.ColocationId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "colocation_id obligatoire")
	}

	monthsAhead := int(req.MonthsAhead)
	if monthsAhead < 1 {
		monthsAhead = 3
	}

	forecasts, err := h.service.GetForecast(ctx, req.ColocationId, monthsAhead)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	var pbForecasts []*pb.MonthlyForecast
	for _, f := range forecasts {
		var categories []*pb.CategoryForecast
		for _, c := range f.Categories {
			categories = append(categories, &pb.CategoryForecast{
				CategoryId:   c.CategoryID,
				CategoryName: c.CategoryName,
				Amount:       c.Amount,
			})
		}
		pbForecasts = append(pbForecasts, &pb.MonthlyForecast{
			Month:       f.Month,
			TotalAmount: f.TotalAmount,
			Categories:  categories,
		})
	}

	return &pb.GetForecastResponse{Forecasts: pbForecasts}, nil
}

// Helper functions

func expenseToProto(e *domain.Expense) *pb.Expense {
	expense := &pb.Expense{
		Id:           e.ID,
		ColocationId: e.ColocationID,
		PaidBy:       e.PaidBy,
		PaidByNom:    e.PaidByNom,
		PaidByPrenom: e.PaidByPrenom,
		CategoryId:   e.CategoryID,
		CategoryName: e.CategoryName,
		Title:        e.Title,
		Description:  e.Description,
		Amount:       e.Amount,
		SplitType:    domainSplitTypeToProto(e.SplitType),
		ExpenseDate:  e.ExpenseDate.Format("2006-01-02"),
		RecurringId:  e.RecurringID,
		CreatedAt:    utils.FormatFrenchDateTime(e.CreatedAt),
	}

	for _, s := range e.Splits {
		expense.Splits = append(expense.Splits, &pb.ExpenseSplit{
			UserId:     s.UserID,
			Amount:     s.Amount,
			Percentage: s.Percentage,
			IsSettled:  s.IsSettled,
			UserNom:    s.UserNom,
			UserPrenom: s.UserPrenom,
		})
	}

	return expense
}

func recurringExpenseToProto(re *domain.RecurringExpense) *pb.RecurringExpense {
	recurring := &pb.RecurringExpense{
		Id:           re.ID,
		ColocationId: re.ColocationID,
		PaidBy:       re.PaidBy,
		PaidByNom:    re.PaidByNom,
		PaidByPrenom: re.PaidByPrenom,
		CategoryId:   re.CategoryID,
		CategoryName: re.CategoryName,
		Title:        re.Title,
		Description:  re.Description,
		Amount:       re.Amount,
		SplitType:    domainSplitTypeToProto(re.SplitType),
		Recurrence:   domainRecurrenceToProto(re.Recurrence),
		NextDueDate:  re.NextDueDate.Format("2006-01-02"),
		IsActive:     re.IsActive,
		CreatedAt:    utils.FormatFrenchDateTime(re.CreatedAt),
	}

	if re.EndDate != nil {
		endDate := re.EndDate.Format("2006-01-02")
		recurring.EndDate = &endDate
	}

	for _, s := range re.Splits {
		recurring.Splits = append(recurring.Splits, &pb.RecurringExpenseSplit{
			UserId:     s.UserID,
			Percentage: s.Percentage,
			UserNom:    s.UserNom,
			UserPrenom: s.UserPrenom,
		})
	}

	return recurring
}

func domainSplitTypeToProto(st domain.SplitType) pb.SplitType {
	switch st {
	case domain.SplitTypeEqual:
		return pb.SplitType_SPLIT_TYPE_EQUAL
	case domain.SplitTypePercentage:
		return pb.SplitType_SPLIT_TYPE_PERCENTAGE
	case domain.SplitTypeCustom:
		return pb.SplitType_SPLIT_TYPE_CUSTOM
	default:
		return pb.SplitType_SPLIT_TYPE_UNSPECIFIED
	}
}

func protoSplitTypeToDomain(st pb.SplitType) domain.SplitType {
	switch st {
	case pb.SplitType_SPLIT_TYPE_EQUAL:
		return domain.SplitTypeEqual
	case pb.SplitType_SPLIT_TYPE_PERCENTAGE:
		return domain.SplitTypePercentage
	case pb.SplitType_SPLIT_TYPE_CUSTOM:
		return domain.SplitTypeCustom
	default:
		return domain.SplitTypeEqual
	}
}

func domainRecurrenceToProto(r domain.Recurrence) pb.Recurrence {
	switch r {
	case domain.RecurrenceDaily:
		return pb.Recurrence_RECURRENCE_DAILY
	case domain.RecurrenceWeekly:
		return pb.Recurrence_RECURRENCE_WEEKLY
	case domain.RecurrenceMonthly:
		return pb.Recurrence_RECURRENCE_MONTHLY
	case domain.RecurrenceYearly:
		return pb.Recurrence_RECURRENCE_YEARLY
	default:
		return pb.Recurrence_RECURRENCE_UNSPECIFIED
	}
}

func protoRecurrenceToDomain(r pb.Recurrence) domain.Recurrence {
	switch r {
	case pb.Recurrence_RECURRENCE_DAILY:
		return domain.RecurrenceDaily
	case pb.Recurrence_RECURRENCE_WEEKLY:
		return domain.RecurrenceWeekly
	case pb.Recurrence_RECURRENCE_MONTHLY:
		return domain.RecurrenceMonthly
	case pb.Recurrence_RECURRENCE_YEARLY:
		return domain.RecurrenceYearly
	default:
		return domain.RecurrenceMonthly
	}
}
