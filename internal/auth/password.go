package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns the bcrypt hash of a plain-text password.
// bcrypt adds a random salt automatically, so the same password
// produces a different hash every time.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether the plain-text password matches the hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
