package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vblanchet22/back_coloc/internal/config"
	"github.com/vblanchet22/back_coloc/internal/repository/postgres"
)

type member struct {
	ID     string
	Prenom string
	Nom    string
}

func main() {
	var colocationID string
	flag.StringVar(&colocationID, "colocation-id", "", "ID de la colocation à alimenter (optionnel)")
	flag.Parse()

	if err := run(colocationID); err != nil {
		log.Fatalf("échec du seed: %v", err)
	}
}

func run(requestedColocID string) error {
	ctx := context.Background()
	cfg := config.Load()
	pool, err := postgres.Connect(&cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	colocID, err := resolveColocationID(ctx, pool, requestedColocID)
	if err != nil {
		return err
	}
	log.Printf("Colocation ciblée: %s", colocID)

	members, err := ensureDemoMembers(ctx, pool, colocID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return fmt.Errorf("aucun membre trouvé pour la colocation %s", colocID)
	}
	log.Printf("%d membres disponibles pour la colocation", len(members))

	categories, err := ensureCategories(ctx, pool, colocID)
	if err != nil {
		return err
	}
	log.Printf("%d catégories prêtes", len(categories))

	if err := seedDemoExpenses(ctx, pool, colocID, members, categories); err != nil {
		return err
	}

	log.Println("Données de démonstration créées avec succès.")
	return nil
}

func resolveColocationID(ctx context.Context, pool *pgxpool.Pool, requested string) (string, error) {
	if requested != "" {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM colocations WHERE id = $1)`, requested).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return "", fmt.Errorf("colocation %s introuvable", requested)
		}
		return requested, nil
	}

	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM colocations ORDER BY created_at DESC LIMIT 1`).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("aucune colocation trouvée en base")
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureDemoMembers(ctx context.Context, pool *pgxpool.Pool, colocID string) ([]member, error) {
	demoProfiles := []struct {
		Email     string
		Prenom    string
		Nom       string
		Telephone string
	}{
		{"alice@coloc.app", "Alice", "Martin", "0600000001"},
		{"bruno@coloc.app", "Bruno", "Dupont", "0600000002"},
		{"clara@coloc.app", "Clara", "Bernard", "0600000003"},
	}

	for _, profile := range demoProfiles {
		var userID string
		err := pool.QueryRow(ctx, `
			INSERT INTO users (email, prenom, nom, telephone, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (email) DO UPDATE
			SET prenom = EXCLUDED.prenom,
			    nom = EXCLUDED.nom,
			    telephone = COALESCE(users.telephone, EXCLUDED.telephone),
			    is_active = true
			RETURNING id
		`, profile.Email, profile.Prenom, profile.Nom, profile.Telephone).Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("création utilisateur demo (%s) impossible: %w", profile.Email, err)
		}

		if _, err := pool.Exec(ctx, `
			INSERT INTO colocation_members (colocation_id, user_id, role)
			VALUES ($1, $2, CASE WHEN NOT EXISTS (
				SELECT 1 FROM colocation_members WHERE colocation_id = $1
			) THEN 'admin' ELSE 'member' END)
			ON CONFLICT (colocation_id, user_id) DO NOTHING
		`, colocID, userID); err != nil {
			return nil, fmt.Errorf("association membre demo impossible: %w", err)
		}
	}

	return fetchMembers(ctx, pool, colocID)
}

func fetchMembers(ctx context.Context, pool *pgxpool.Pool, colocID string) ([]member, error) {
	rows, err := pool.Query(ctx, `
		SELECT cm.user_id, COALESCE(u.prenom, ''), COALESCE(u.nom, '')
		FROM colocation_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.colocation_id = $1
		ORDER BY cm.joined_at
	`, colocID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []member
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.ID, &m.Prenom, &m.Nom); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func ensureCategories(ctx context.Context, pool *pgxpool.Pool, colocID string) (map[string]string, error) {
	type category struct {
		Name  string
		Icon  string
		Color string
	}
	categories := []category{
		{"Loyer", "home", "#4C6EF5"},
		{"Courses", "shopping-cart", "#10B981"},
		{"Énergie", "zap", "#F59E0B"},
		{"Internet", "wifi", "#0EA5E9"},
	}

	for _, cat := range categories {
		if _, err := pool.Exec(ctx, `
			INSERT INTO expense_categories (name, icon, color, colocation_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`, cat.Name, cat.Icon, cat.Color, colocID); err != nil {
			return nil, fmt.Errorf("impossible de créer la catégorie %s: %w", cat.Name, err)
		}
	}

	rows, err := pool.Query(ctx, `
		SELECT name, id
		FROM expense_categories
		WHERE colocation_id = $1
		  AND name IN ($2, $3, $4, $5)
	`, colocID, categories[0].Name, categories[1].Name, categories[2].Name, categories[3].Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, id string
		if err := rows.Scan(&name, &id); err != nil {
			return nil, err
		}
		result[name] = id
	}

	return result, nil
}

func seedDemoExpenses(ctx context.Context, pool *pgxpool.Pool, colocID string, members []member, categories map[string]string) error {
	if len(members) == 0 {
		return fmt.Errorf("aucun membre pour générer des dépenses")
	}

	if _, err := pool.Exec(ctx, `
		DELETE FROM expenses
		WHERE colocation_id = $1
		  AND description LIKE '[DEMO]%'
	`, colocID); err != nil {
		return err
	}

	type demoExpense struct {
		Title       string
		Description string
		Amount      float64
		Category    string
		DaysAgo     int
		PayerIndex  int
	}

	demos := []demoExpense{
		{"Loyer février", "[DEMO] Loyer mensuel", 1200, "Loyer", 7, 0},
		{"Courses semaine", "[DEMO] Plein de courses", 180, "Courses", 4, 1},
		{"Facture d'énergie", "[DEMO] Électricité + gaz", 95.4, "Énergie", 12, 2},
		{"Box Internet", "[DEMO] Abonnement fibre", 39.99, "Internet", 2, 0},
		{"Courses week-end", "[DEMO] Apéro & brunch", 72.2, "Courses", 1, 2},
	}

	for _, exp := range demos {
		categoryID, ok := categories[exp.Category]
		if !ok {
			continue
		}
		payer := members[exp.PayerIndex%len(members)]
		expenseDate := time.Now().AddDate(0, 0, -exp.DaysAgo).Format("2006-01-02")

		var expenseID string
		if err := pool.QueryRow(ctx, `
			INSERT INTO expenses (colocation_id, paid_by, category_id, title, description, amount, split_type, expense_date)
			VALUES ($1, $2, $3, $4, $5, $6, 'equal', $7)
			RETURNING id
		`, colocID, payer.ID, categoryID, exp.Title, exp.Description, exp.Amount, expenseDate).Scan(&expenseID); err != nil {
			return fmt.Errorf("impossible d'insérer la dépense %s: %w", exp.Title, err)
		}

		share := exp.Amount / float64(len(members))
		share = math.Round(share*100) / 100
		totalAssigned := 0.0

		for i, m := range members {
			amount := share
			if i == len(members)-1 {
				amount = math.Round((exp.Amount-totalAssigned)*100) / 100
			}
			totalAssigned += amount

			if _, err := pool.Exec(ctx, `
				INSERT INTO expense_splits (expense_id, user_id, amount, percentage)
				VALUES ($1, $2, $3, $4)
			`, expenseID, m.ID, amount, 100.0/float64(len(members))); err != nil {
				return fmt.Errorf("impossible d'insérer le split de %s: %w", exp.Title, err)
			}
		}
	}

	return nil
}
