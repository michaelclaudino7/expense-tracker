package transaction

import "testing"

func TestInputValidate(t *testing.T) {
	valid := Input{
		CategoryID:  1,
		Description: "Weekly shopping",
		AmountCents: 8750,
		OccurredOn:  "2026-11-18",
	}

	tests := []struct {
		name    string
		mutate  func(in *Input)
		wantMsg string
	}{
		{"valid input", func(in *Input) {}, ""},
		{"missing category", func(in *Input) { in.CategoryID = 0 }, "category_id is required"},
		{"empty description", func(in *Input) { in.Description = "  " }, "description is required"},
		{"zero amount", func(in *Input) { in.AmountCents = 0 }, "amount_cents must be greater than zero"},
		{"negative amount", func(in *Input) { in.AmountCents = -100 }, "amount_cents must be greater than zero"},
		{"invalid date", func(in *Input) { in.OccurredOn = "2026-02-31" }, "occurred_on must be a valid date in the format YYYY-MM-DD"},
		{"wrong date format", func(in *Input) { in.OccurredOn = "18/11/2026" }, "occurred_on must be a valid date in the format YYYY-MM-DD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid
			tt.mutate(&in)
			got := in.Validate()
			if got != tt.wantMsg {
				t.Errorf("Validate() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}
