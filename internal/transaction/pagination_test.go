package transaction

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		name string
		p    Pagination
		want int
	}{
		{"first page", Pagination{Page: 1, PageSize: 10}, 0},
		{"second page", Pagination{Page: 2, PageSize: 10}, 10},
		{"third page of 20", Pagination{Page: 3, PageSize: 20}, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewPagedResultTotalPages(t *testing.T) {
	tests := []struct {
		name  string
		total int
		size  int
		want  int
	}{
		{"no rows", 0, 20, 0},
		{"exact fit", 40, 20, 2},
		{"rounds up", 5, 2, 3},
		{"single row", 1, 20, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := newPagedResult([]Transaction{}, tt.total, Pagination{Page: 1, PageSize: tt.size})
			if res.TotalPages != tt.want {
				t.Errorf("TotalPages = %d, want %d", res.TotalPages, tt.want)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantOK   bool
		wantPage int
		wantSize int
	}{
		{"defaults", "", true, 1, 20},
		{"custom values", "?page=3&page_size=50", true, 3, 50},
		{"max page size", "?page_size=100", true, 1, 100},
		{"page zero", "?page=0", false, 0, 0},
		{"negative page", "?page=-1", false, 0, 0},
		{"page not a number", "?page=abc", false, 0, 0},
		{"page size zero", "?page_size=0", false, 0, 0},
		{"page size too big", "?page_size=101", false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/transactions"+tt.query, nil)
			rec := httptest.NewRecorder()

			p, ok := parsePagination(rec, req)

			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
				}
				return
			}
			if p.Page != tt.wantPage || p.PageSize != tt.wantSize {
				t.Errorf("got page=%d size=%d, want page=%d size=%d", p.Page, p.PageSize, tt.wantPage, tt.wantSize)
			}
		})
	}
}
