package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "s3cret-pass" {
		t.Fatal("hash must not equal the plain password")
	}
	if !CheckPassword(hash, "s3cret-pass") {
		t.Error("CheckPassword() = false for the correct password")
	}
	if CheckPassword(hash, "wrong-pass") {
		t.Error("CheckPassword() = true for a wrong password")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	m := NewTokenManager("test-secret", time.Hour)

	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	got, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got != 42 {
		t.Errorf("Parse() = %d, want 42", got)
	}
}

func TestTokenRejected(t *testing.T) {
	m := NewTokenManager("test-secret", time.Hour)
	valid, _ := m.Generate(1)

	expired, _ := NewTokenManager("test-secret", -time.Minute).Generate(1)
	wrongSecret, _ := NewTokenManager("other-secret", time.Hour).Generate(1)

	noExp := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "1"})
	noExpString, _ := noExp.SignedString([]byte("test-secret"))

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"garbage", "not-a-token"},
		{"expired", expired},
		{"wrong secret", wrongSecret},
		{"no expiration", noExpString},
		{"tampered", valid + "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := m.Parse(tt.token); err != ErrInvalidToken {
				t.Errorf("Parse() error = %v, want ErrInvalidToken", err)
			}
		})
	}
}
