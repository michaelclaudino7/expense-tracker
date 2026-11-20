package report

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{db: db}
}

// Monthly builds the report for the month that starts at the given date.
func (r *Repository) Monthly(ctx context.Context, start time.Time) (Monthly, error) {
    end := start.AddDate(0, 1, 0) // first day of the next month (exclusive)

    rows, err := r.db.Query(ctx,
        `SELECT c.id, c.name, c.type, SUM(t.amount_cents)::bigint
         FROM transactions t
         JOIN categories c ON c.id = t.category_id
         WHERE t.occurred_on >= $1 AND t.occurred_on < $2
         GROUP BY c.id, c.name, c.type
         ORDER BY c.type DESC, SUM(t.amount_cents) DESC`,
        start, end,
    )
    if err != nil {
        return Monthly{}, err
    }
    defer rows.Close()

    report := Monthly{
        Month:      start.Format(monthLayout),
        ByCategory: []CategoryTotal{},
    }

    for rows.Next() {
        var ct CategoryTotal
        if err := rows.Scan(&ct.CategoryID, &ct.CategoryName, &ct.Type, &ct.TotalCents); err != nil {
            return Monthly{}, err
        }

        switch ct.Type {
        case "income":
            report.TotalIncomeCents += ct.TotalCents
        case "expense":
            report.TotalExpenseCents += ct.TotalCents
        }
        report.ByCategory = append(report.ByCategory, ct)
    }
    if err := rows.Err(); err != nil {
        return Monthly{}, err
    }

    report.BalanceCents = report.TotalIncomeCents - report.TotalExpenseCents
    return report, nil
}