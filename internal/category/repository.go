package category

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL error codes we care about.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID int64, in Input) (Category, error) {
	var c Category
	err := r.db.QueryRow(ctx,
		`INSERT INTO categories (user_id, name, type)
         VALUES ($1, $2, $3)
         RETURNING id, name, type, created_at`,
		userID, in.Name, in.Type,
	).Scan(&c.ID, &c.Name, &c.Type, &c.CreatedAt)
	if err != nil {
		return Category{}, translateError(err)
	}
	return c, nil
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, type, created_at
         FROM categories
         WHERE user_id = $1
         ORDER BY name`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *Repository) Get(ctx context.Context, userID, id int64) (Category, error) {
	var c Category
	err := r.db.QueryRow(ctx,
		`SELECT id, name, type, created_at
         FROM categories
         WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&c.ID, &c.Name, &c.Type, &c.CreatedAt)
	if err != nil {
		return Category{}, translateError(err)
	}
	return c, nil
}

func (r *Repository) Update(ctx context.Context, userID, id int64, in Input) (Category, error) {
	var c Category
	err := r.db.QueryRow(ctx,
		`UPDATE categories
         SET name = $1, type = $2
         WHERE id = $3 AND user_id = $4
         RETURNING id, name, type, created_at`,
		in.Name, in.Type, id, userID,
	).Scan(&c.ID, &c.Name, &c.Type, &c.CreatedAt)
	if err != nil {
		return Category{}, translateError(err)
	}
	return c, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM categories WHERE id = $1 AND user_id = $2`,
		id, userID)
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
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return ErrAlreadyExists
		case pgForeignKeyViolation:
			return ErrInUse
		}
	}
	return err
}
