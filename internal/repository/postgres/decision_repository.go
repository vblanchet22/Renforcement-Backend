package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/domain"
)

// DecisionRepository handles decision database operations
type DecisionRepository struct {
	pool *pgxpool.Pool
}

// NewDecisionRepository creates a new DecisionRepository
func NewDecisionRepository(pool *pgxpool.Pool) *DecisionRepository {
	return &DecisionRepository{pool: pool}
}

// Create creates a new decision
func (r *DecisionRepository) Create(ctx context.Context, decision *domain.Decision) error {
	optionsJSON, err := json.Marshal(decision.Options)
	if err != nil {
		return fmt.Errorf("erreur de serialisation des options: %w", err)
	}

	query := `
		INSERT INTO decisions (colocation_id, created_by, title, description, options, deadline, allow_multiple, is_anonymous)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, status, created_at
	`

	return r.pool.QueryRow(ctx, query,
		decision.ColocationID,
		decision.CreatedBy,
		decision.Title,
		decision.Description,
		optionsJSON,
		decision.Deadline,
		decision.AllowMultiple,
		decision.IsAnonymous,
	).Scan(&decision.ID, &decision.Status, &decision.CreatedAt)
}

// GetByID retrieves a decision by ID, including current user's vote info
func (r *DecisionRepository) GetByID(ctx context.Context, id, currentUserID string) (*domain.Decision, error) {
	query := `
		SELECT d.id, d.colocation_id, d.created_by, d.title, d.description, d.options,
		       d.status, d.deadline, d.allow_multiple, d.is_anonymous, d.created_at,
		       u.nom, u.prenom,
		       (SELECT COUNT(DISTINCT dv.user_id) FROM decision_votes dv WHERE dv.decision_id = d.id) as vote_count
		FROM decisions d
		INNER JOIN users u ON d.created_by = u.id
		WHERE d.id = $1
	`

	var d domain.Decision
	var optionsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.ColocationID, &d.CreatedBy, &d.Title, &d.Description, &optionsJSON,
		&d.Status, &d.Deadline, &d.AllowMultiple, &d.IsAnonymous, &d.CreatedAt,
		&d.CreatedByNom, &d.CreatedByPrenom,
		&d.VoteCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la recuperation: %w", err)
	}

	if err := json.Unmarshal(optionsJSON, &d.Options); err != nil {
		return nil, fmt.Errorf("erreur de deserialization des options: %w", err)
	}

	// Check if current user has voted and get their votes
	userVotes, err := r.GetUserVotes(ctx, id, currentUserID)
	if err != nil {
		return nil, err
	}
	d.HasVoted = len(userVotes) > 0
	d.UserVotes = userVotes

	return &d, nil
}

// GetUserVotes returns the option indices the user voted for
func (r *DecisionRepository) GetUserVotes(ctx context.Context, decisionID, userID string) ([]int, error) {
	query := `SELECT option_index FROM decision_votes WHERE decision_id = $1 AND user_id = $2`

	rows, err := r.pool.Query(ctx, query, decisionID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []int
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err != nil {
			return nil, err
		}
		votes = append(votes, idx)
	}

	return votes, rows.Err()
}

// ListByColocation lists decisions for a colocation
func (r *DecisionRepository) ListByColocation(ctx context.Context, colocationID, currentUserID string, status *string, page, pageSize int) ([]domain.Decision, int, error) {
	baseQuery := `
		FROM decisions d
		INNER JOIN users u ON d.created_by = u.id
		WHERE d.colocation_id = $1
	`

	args := []interface{}{colocationID}
	argIndex := 2

	if status != nil {
		baseQuery += fmt.Sprintf(" AND d.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// Count
	var totalCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Select
	selectQuery := fmt.Sprintf(`
		SELECT d.id, d.colocation_id, d.created_by, d.title, d.description, d.options,
		       d.status, d.deadline, d.allow_multiple, d.is_anonymous, d.created_at,
		       u.nom, u.prenom,
		       (SELECT COUNT(DISTINCT dv.user_id) FROM decision_votes dv WHERE dv.decision_id = d.id) as vote_count
	`+baseQuery+" ORDER BY d.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var decisions []domain.Decision
	for rows.Next() {
		var d domain.Decision
		var optionsJSON []byte

		if err := rows.Scan(
			&d.ID, &d.ColocationID, &d.CreatedBy, &d.Title, &d.Description, &optionsJSON,
			&d.Status, &d.Deadline, &d.AllowMultiple, &d.IsAnonymous, &d.CreatedAt,
			&d.CreatedByNom, &d.CreatedByPrenom,
			&d.VoteCount,
		); err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(optionsJSON, &d.Options); err != nil {
			return nil, 0, err
		}

		// Get user votes
		userVotes, err := r.GetUserVotes(ctx, d.ID, currentUserID)
		if err != nil {
			return nil, 0, err
		}
		d.HasVoted = len(userVotes) > 0
		d.UserVotes = userVotes

		decisions = append(decisions, d)
	}

	return decisions, totalCount, rows.Err()
}

// Update updates a decision
func (r *DecisionRepository) Update(ctx context.Context, decision *domain.Decision) error {
	optionsJSON, err := json.Marshal(decision.Options)
	if err != nil {
		return fmt.Errorf("erreur de serialisation des options: %w", err)
	}

	query := `
		UPDATE decisions
		SET title = $1, description = $2, options = $3, deadline = $4, allow_multiple = $5, is_anonymous = $6
		WHERE id = $7
	`

	_, err = r.pool.Exec(ctx, query,
		decision.Title,
		decision.Description,
		optionsJSON,
		decision.Deadline,
		decision.AllowMultiple,
		decision.IsAnonymous,
		decision.ID,
	)
	return err
}

// Delete deletes a decision and its votes
func (r *DecisionRepository) Delete(ctx context.Context, id string) error {
	// Votes are deleted by CASCADE
	query := `DELETE FROM decisions WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("decision introuvable")
	}

	return nil
}

// Vote adds votes for a user
func (r *DecisionRepository) Vote(ctx context.Context, decisionID, userID string, optionIndices []int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing votes for this user
	_, err = tx.Exec(ctx, "DELETE FROM decision_votes WHERE decision_id = $1 AND user_id = $2", decisionID, userID)
	if err != nil {
		return err
	}

	// Insert new votes
	for _, idx := range optionIndices {
		_, err = tx.Exec(ctx,
			"INSERT INTO decision_votes (decision_id, user_id, option_index) VALUES ($1, $2, $3)",
			decisionID, userID, idx,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// HasVotes checks if a decision has any votes
func (r *DecisionRepository) HasVotes(ctx context.Context, decisionID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM decision_votes WHERE decision_id = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, decisionID).Scan(&exists)
	return exists, err
}

// UpdateStatus updates the status of a decision
func (r *DecisionRepository) UpdateStatus(ctx context.Context, id string, status domain.DecisionStatus) error {
	query := `UPDATE decisions SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

// GetResults returns the results of a decision
func (r *DecisionRepository) GetResults(ctx context.Context, decisionID string) ([]domain.OptionResult, int, int, error) {
	// Get decision to know options and if anonymous
	var optionsJSON []byte
	var isAnonymous bool
	err := r.pool.QueryRow(ctx,
		"SELECT options, is_anonymous FROM decisions WHERE id = $1", decisionID,
	).Scan(&optionsJSON, &isAnonymous)
	if err != nil {
		return nil, 0, 0, err
	}

	var options []string
	if err := json.Unmarshal(optionsJSON, &options); err != nil {
		return nil, 0, 0, err
	}

	// Get vote counts per option
	query := `
		SELECT option_index, COUNT(*) as vote_count
		FROM decision_votes
		WHERE decision_id = $1
		GROUP BY option_index
		ORDER BY option_index
	`

	rows, err := r.pool.Query(ctx, query, decisionID)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	voteCounts := make(map[int]int)
	totalVotes := 0
	for rows.Next() {
		var idx, count int
		if err := rows.Scan(&idx, &count); err != nil {
			return nil, 0, 0, err
		}
		voteCounts[idx] = count
		totalVotes += count
	}

	// Get total unique voters
	var totalVoters int
	err = r.pool.QueryRow(ctx,
		"SELECT COUNT(DISTINCT user_id) FROM decision_votes WHERE decision_id = $1", decisionID,
	).Scan(&totalVoters)
	if err != nil {
		return nil, 0, 0, err
	}

	// Build results
	var results []domain.OptionResult
	for i, optionText := range options {
		result := domain.OptionResult{
			OptionIndex: i,
			OptionText:  optionText,
			VoteCount:   voteCounts[i],
		}

		if totalVotes > 0 {
			result.Percentage = float64(voteCounts[i]) / float64(totalVotes) * 100
		}

		// Get voters if not anonymous
		if !isAnonymous {
			voters, err := r.GetVotersForOption(ctx, decisionID, i)
			if err != nil {
				return nil, 0, 0, err
			}
			result.Voters = voters
		}

		results = append(results, result)
	}

	return results, totalVotes, totalVoters, nil
}

// GetVotersForOption returns voters for a specific option
func (r *DecisionRepository) GetVotersForOption(ctx context.Context, decisionID string, optionIndex int) ([]domain.Voter, error) {
	query := `
		SELECT u.id, u.nom, u.prenom
		FROM decision_votes dv
		INNER JOIN users u ON dv.user_id = u.id
		WHERE dv.decision_id = $1 AND dv.option_index = $2
	`

	rows, err := r.pool.Query(ctx, query, decisionID, optionIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voters []domain.Voter
	for rows.Next() {
		var v domain.Voter
		if err := rows.Scan(&v.UserID, &v.UserNom, &v.UserPrenom); err != nil {
			return nil, err
		}
		voters = append(voters, v)
	}

	return voters, rows.Err()
}
