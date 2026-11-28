package transaction

import (
	"errors"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

var (
	ErrNotFound         = errors.New("transaction not found")
	ErrCategoryNotFound = errors.New("category not found")
)

// Transaction is a single income or expense entry.
// Type and CategoryName come from the related category.
type Transaction struct {
	ID           int64     `json:"id"`
	CategoryID   int64     `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	AmountCents  int64     `json:"amount_cents"`
	OccurredOn   string    `json:"occurred_on"`
	CreatedAt    time.Time `json:"created_at"`
}

// Input is the JSON body accepted when creating or updating a transaction.
type Input struct {
	CategoryID  int64  `json:"category_id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	OccurredOn  string `json:"occurred_on"` // format: YYYY-MM-DD
}

// Validate cleans the input and returns a message describing the first problem found.
func (in *Input) Validate() string {
	in.Description = strings.TrimSpace(in.Description)

	if in.CategoryID <= 0 {
		return "category_id is required"
	}
	if in.Description == "" {
		return "description is required"
	}
	if len(in.Description) > 200 {
		return "description must have at most 200 characters"
	}
	if in.AmountCents <= 0 {
		return "amount_cents must be greater than zero"
	}
	if _, err := time.Parse(dateLayout, in.OccurredOn); err != nil {
		return "occurred_on must be a valid date in the format YYYY-MM-DD"
	}
	return ""
}

// Filter holds the optional filters accepted when listing transactions.
type Filter struct {
	From       *time.Time
	To         *time.Time
	CategoryID *int64
}
