package category

import (
	"errors"
	"strings"
	"time"
)

// Category types allowed by the database CHECK constraint.
const (
	TypeIncome  = "income"
	TypeExpense = "expense"
)

var (
	ErrNotFound      = errors.New("category not found")
	ErrAlreadyExists = errors.New("category already exists")
	ErrInUse         = errors.New("category is used by transactions")
)

// Category is a group for transactions, e.g. "Groceries" (expense) or "Salary" (income).
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// Input is the JSON body accepted when creating or updating a category.
type Input struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Validate cleans the input and returns a message describing the first problem found.
func (in *Input) Validate() string {
	in.Name = strings.TrimSpace(in.Name)

	if in.Name == "" {
		return "name is required"
	}
	if len(in.Name) > 100 {
		return "name must have at most 100 characters"
	}
	if in.Type != TypeIncome && in.Type != TypeExpense {
		return "type must be 'income' or 'expense'"
	}
	return ""
}
