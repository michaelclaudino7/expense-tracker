package category

import "testing"

func TestInputValidate(t *testing.T) {
tests := []struct {
name    string
input   Input
wantMsg string
}{
{"valid expense", Input{Name: "Groceries", Type: TypeExpense}, ""},
{"valid income", Input{Name: "Salary", Type: TypeIncome}, ""},
{"trims spaces", Input{Name: "  Rent  ", Type: TypeExpense}, ""},
{"empty name", Input{Name: "", Type: TypeExpense}, "name is required"},
{"blank name", Input{Name: "   ", Type: TypeExpense}, "name is required"},
{"invalid type", Input{Name: "Test", Type: "banana"}, "type must be 'income' or 'expense'"},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := tt.input.Validate()
if got != tt.wantMsg {
t.Errorf("Validate() = %q, want %q", got, tt.wantMsg)
}
})
}
}

func TestInputValidateTrimsName(t *testing.T) {
in := Input{Name: "  Rent  ", Type: TypeExpense}
in.Validate()
if in.Name != "Rent" {
t.Errorf("Name = %q, want %q", in.Name, "Rent")
}
}