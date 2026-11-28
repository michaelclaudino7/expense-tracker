package transaction

import (
	"net/http"
	"strconv"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Pagination holds the page requested by the client (page starts at 1).
type Pagination struct {
	Page     int
	PageSize int
}

// Offset is how many rows to skip in the SQL query.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PagedResult is the response of the list endpoint.
type PagedResult struct {
	Data       []Transaction `json:"data"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

func newPagedResult(data []Transaction, total int, p Pagination) PagedResult {
	return PagedResult{
		Data:       data,
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: (total + p.PageSize - 1) / p.PageSize,
	}
}

// parsePagination reads the optional page and page_size query params.
// On failure it writes a 400 and returns false.
func parsePagination(w http.ResponseWriter, r *http.Request) (Pagination, bool) {
	p := Pagination{Page: defaultPage, PageSize: defaultPageSize}
	q := r.URL.Query()

	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "page must be a positive integer")
			return Pagination{}, false
		}
		p.Page = n
	}
	if v := q.Get("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPageSize {
			writeError(w, http.StatusBadRequest, "page_size must be an integer between 1 and 100")
			return Pagination{}, false
		}
		p.PageSize = n
	}
	return p, true
}
