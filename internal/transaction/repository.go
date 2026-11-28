package transaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Columns returned by every query. Joined with categories to get name and type.
const selectColumns = `t.id, t.category_id, c.name, c.type, t.description,
    t.amount_cents, to_char(t.occurred_on, 'YYYY-MM-DD'), t.created_at`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// scanner is satisfied by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanTransaction(s scanner) (Transaction, error) {
	var t Transaction
	err := s.Scan(
		&t.ID, &t.CategoryID, &t.CategoryName, &t.Type, &t.Description,
		&t.AmountCents, &t.OccurredOn, &t.CreatedAt,
	)
	return t, err
}

// Create inserts a transaction only if the category belongs to the same user.
// When it does not, the INSERT ... SELECT produces no row and we return ErrCategoryNotFound.
func (r *Repository) Create(ctx context.Context, userID int64, in Input) (Transaction, error) {
	date, err := time.Parse(dateLayout, in.OccurredOn)
	if err != nil {
		return Transaction{}, err
	}

	query := `
        WITH t AS (
            INSERT INTO transactions (user_id, category_id, description, amount_cents, occurred_on)
            SELECT $1, c.id, $3, $4, $5
            FROM categories c
            WHERE c.id = $2 AND c.user_id = $1
            RETURNING *
        )
        SELECT ` + selectColumns + `
        FROM t
        JOIN categories c ON c.id = t.category_id`

	row := r.db.QueryRow(ctx, query, userID, in.CategoryID, in.Description, in.AmountCents, date)
	tx, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrCategoryNotFound
		}
		return Transaction{}, err
	}
	return tx, nil
}

func (r *Repository) List(ctx context.Context, userID int64, f Filter, p Pagination) ([]Transaction, int, error) {
	args := []any{userID}
	conditions := []string{"t.user_id = $1"}

	// Placeholders ($1, $2...) are generated, but values always go in args.
	if f.From != nil {
		args = append(args, *f.From)
		conditions = append(conditions, fmt.Sprintf("t.occurred_on >= $%d", len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		conditions = append(conditions, fmt.Sprintf("t.occurred_on <= $%d", len(args)))
	}
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		conditions = append(conditions, fmt.Sprintf("t.category_id = $%d", len(args)))
	}

	where := " WHERE " + strings.Join(conditions, " AND ")

	// First, how many rows match the filters (ignoring the page).
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions t`+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Then, only the rows of the requested page.
	args = append(args, p.PageSize, p.Offset())
	query := `SELECT ` + selectColumns + `
        FROM transactions t
        JOIN categories c ON c.id = t.category_id` + where +
		fmt.Sprintf(" ORDER BY t.occurred_on DESC, t.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, total, rows.Err()
}

func (r *Repository) Get(ctx context.Context, userID, id int64) (Transaction, error) {
	query := `SELECT ` + selectColumns + `
        FROM transactions t
        JOIN categories c ON c.id = t.category_id
        WHERE t.id = $1 AND t.user_id = $2`

	tx, err := scanTransaction(r.db.QueryRow(ctx, query, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrNotFound
		}
		return Transaction{}, err
	}
	return tx, nil
}

// Update changes a transaction of the user. The new category must also belong to the user.
func (r *Repository) Update(ctx context.Context, userID, id int64, in Input) (Transaction, error) {
	date, err := time.Parse(dateLayout, in.OccurredOn)
	if err != nil {
		return Transaction{}, err
	}

	// Check the target category first, so we can tell the two failures apart.
	var categoryExists bool
	err = r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`,
		in.CategoryID, userID,
	).Scan(&categoryExists)
	if err != nil {
		return Transaction{}, err
	}
	if !categoryExists {
		return Transaction{}, ErrCategoryNotFound
	}

	query := `
        WITH t AS (
            UPDATE transactions
            SET category_id = $1, description = $2, amount_cents = $3, occurred_on = $4
            WHERE id = $5 AND user_id = $6
            RETURNING *
        )
        SELECT ` + selectColumns + `
        FROM t
        JOIN categories c ON c.id = t.category_id`

	row := r.db.QueryRow(ctx, query, in.CategoryID, in.Description, in.AmountCents, date, id, userID)
	tx, err := scanTransaction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Transaction{}, ErrNotFound
		}
		return Transaction{}, err
	}
	return tx, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
