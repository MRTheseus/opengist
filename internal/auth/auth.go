package auth

import (
	passwordpkg "github.com/thomiceli/opengist/internal/auth/password"
	"github.com/thomiceli/opengist/internal/db"
)

func VerifyPassword(user *db.User, password string) bool {
	return passwordpkg.VerifyPassword(user.Password, password)
}

func HashPassword(password string) (string, error) {
	return passwordpkg.HashPassword(password)
}
