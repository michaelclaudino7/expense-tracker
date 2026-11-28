package report

import "time"

const monthLayout = "2006-01"

// CategoryTotal is the sum of all transactions of one category in the month.
type CategoryTotal struct {
	CategoryID   int64  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Type         string `json:"type"`
	TotalCents   int64  `json:"total_cents"`
}

// Monthly is the report returned by GET /reports/monthly.
type Monthly struct {
	Month             string          `json:"month"`
	TotalIncomeCents  int64           `json:"total_income_cents"`
	TotalExpenseCents int64           `json:"total_expense_cents"`
	BalanceCents      int64           `json:"balance_cents"`
	ByCategory        []CategoryTotal `json:"by_category"`
}

// parseMonth converts "2026-11" into the first day of that month.
func parseMonth(s string) (time.Time, error) {
	return time.Parse(monthLayout, s)
}
