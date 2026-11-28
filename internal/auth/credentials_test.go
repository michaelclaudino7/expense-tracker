package auth

import "testing"

func TestCredentialsValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   Credentials
		wantMsg string
	}{
		{"valid", Credentials{Email: "me@example.com", Password: "super-secret-1"}, ""},
		{"uppercase and spaces are normalized", Credentials{Email: "  ME@Example.com ", Password: "super-secret-1"}, ""},
		{"empty email", Credentials{Email: "", Password: "super-secret-1"}, "email must be a valid address"},
		{"no at sign", Credentials{Email: "invalid", Password: "super-secret-1"}, "email must be a valid address"},
		{"display name form", Credentials{Email: "Bob <bob@example.com>", Password: "super-secret-1"}, "email must be a valid address"},
		{"short password", Credentials{Email: "me@example.com", Password: "1234567"}, "password must have at least 8 characters"},
		{"password too long", Credentials{Email: "me@example.com", Password: string(make([]byte, 73))}, "password must have at most 72 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Validate(); got != tt.wantMsg {
				t.Errorf("Validate() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestCredentialsNormalize(t *testing.T) {
	c := Credentials{Email: "  ME@Example.COM ", Password: "x"}
	c.Normalize()
	if c.Email != "me@example.com" {
		t.Errorf("Email = %q, want %q", c.Email, "me@example.com")
	}
}
