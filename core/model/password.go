package model

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
)

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword verifies a plaintext password against a stored value.
// Supports bcrypt hashes and legacy plaintext (for smooth migration).
// Returns (ok, needUpgrade) — needUpgrade is true when the stored value is
// still plaintext and should be upgraded to bcrypt.
func CheckPassword(stored, plain string) (bool, bool) {
	if stored == "" {
		return false, false
	}

	// bcrypt hash: "$2a$", "$2b$", "$2y$"
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		if bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil {
			return true, false
		}
		return false, false
	}

	// legacy plaintext (demo/seed data)
	if stored == plain {
		return true, true
	}
	return false, false
}
