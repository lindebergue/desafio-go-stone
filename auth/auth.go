package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// jwtClaims contains the claims of a issued JSON Web Token.
type jwtClaims struct {
	jwt.StandardClaims
	AccountID int64 `json:"https://desafio-go-stone.local/account_id"`
}

// jwtMethod is the method for signing JSON Web Token.
var jwtMethod = jwt.SigningMethodHS256

// HashPassword returns the hash for password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compares a password hash with its value.
func ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// EncodeToken returns a JSON Web Token for a session authenticated as
// accountID. The token is valid for one hour.
func EncodeToken(secret []byte, accountID int64) (string, error) {
	claims := &jwtClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
		AccountID: accountID,
	}
	return jwt.NewWithClaims(jwtMethod, claims).SignedString(secret)
}

// DecodeToken returns the accountID encoded from a JSON Web token. Returns an
// error if the token is malformed or has expired.
func DecodeToken(secret []byte, token string) (accountID int64, err error) {
	t, err := jwt.ParseWithClaims(
		token,
		&jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwtMethod.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		},
	)
	if err != nil {
		return 0, err
	}
	return t.Claims.(*jwtClaims).AccountID, nil
}
