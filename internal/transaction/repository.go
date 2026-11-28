package transaction

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgForeignKeyViolation = "23503"

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

func (r *Repository) Create(ctx context.Context, in Input) (Transaction, error) {
	date, err := time.Parse(dateLayout, in.OccurredOn)
	if err != nil {
		return Transaction{}, err
	}

	query := `
        WITH t AS (
            INSERT INTO transactions (category_id, description, amount_cents, occurred_on)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT ` + selectColumns + `
        FROM t
        JOIN categories c ON c.id = t.category_id`

	row := r.db.QueryRow(ctx, query, in.CategoryID, in.Description, in.AmountCents, date)
	tx, err := scanTransaction(row)
	if err != nil {
		return Transaction{}, translateError(err)
	}
	return tx, nil
}

func (r *Repository) List(ctx context.Context, f Filter, p Pagination) ([]Transaction, int, error) {
	conditions := []string{}
	args := []any{}

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

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

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

func (r *Repository) Get(ctx context.Context, id int64) (Transaction, error) {
	query := `SELECT ` + selectColumns + `
        FROM transactions t
        JOIN categories c ON c.id = t.category_id
        WHERE t.id = $1`

	tx, err := scanTransaction(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return Transaction{}, translateError(err)
	}
	return tx, nil
}

func (r *Repository) Update(ctx context.Context, id int64, in Input) (Transaction, error) {
	date, err := time.Parse(dateLayout, in.OccurredOn)
	if err != nil {
		return Transaction{}, err
	}

	query := `
        WITH t AS (
            UPDATE transactions
            SET category_id = $1, description = $2, amount_cents = $3, occurred_on = $4
            WHERE id = $5
            RETURNING *
        )
        SELECT ` + selectColumns + `
        FROM t
        JOIN categories c ON c.id = t.category_id`

	row := r.db.QueryRow(ctx, query, in.CategoryID, in.Description, in.AmountCents, date, id)
	tx, err := scanTransaction(row)
	if err != nil {
		return Transaction{}, translateError(err)
	}
	return tx, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return translateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// translateError converts low-level database errors into our own named errors.
func translateError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
		return ErrCategoryNotFound
	}
	return err
}
