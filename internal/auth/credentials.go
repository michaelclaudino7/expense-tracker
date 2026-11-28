package auth

import (
	"net/mail"
	"strings"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // bcrypt only uses the first 72 bytes
)

// Credentials is the JSON body of the register and login endpoints.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Normalize trims and lowercases the email so "Me@Mail.com " and "me@mail.com" are the same user.
func (c *Credentials) Normalize() {
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
}

// Validate normalizes the input and returns a message describing the first problem found.
func (c *Credentials) Validate() string {
	c.Normalize()

	addr, err := mail.ParseAddress(c.Email)
	if err != nil || addr.Address != c.Email || len(c.Email) > 254 {
		return "email must be a valid address"
	}
	if len(c.Password) < minPasswordLength {
		return "password must have at least 8 characters"
	}
	if len(c.Password) > maxPasswordLength {
		return "password must have at most 72 bytes"
	}
	return ""
}
