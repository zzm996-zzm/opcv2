package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type JWTManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *JWTManager) IssueAccess(user User) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, expiresAt, err
}

func (m *JWTManager) ParseAccess(rawToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(
		rawToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidAccessToken
			}
			return m.secret, nil
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return 0, ErrInvalidAccessToken
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, ErrInvalidAccessToken
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidAccessToken
	}
	return userID, nil
}
