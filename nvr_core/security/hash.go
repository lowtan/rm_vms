package security

import "golang.org/x/crypto/bcrypt"

// HashPassword generates a bcrypt hash from a plaintext password.
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost is currently 10, which provides a solid balance 
	// between security and server CPU load during login.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}