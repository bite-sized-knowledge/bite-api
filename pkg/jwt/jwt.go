package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID     int64  `json:"id,omitempty"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Birth  *int   `json:"birth,omitempty"`
	Gender string `json:"gender"`
	Status string `json:"status"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

type RefreshClaims struct {
	MemberID int64 `json:"member_id"`
	jwtv5.RegisteredClaims
}

type Service struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewService(secretKey string, accessExpiry, refreshExpiry time.Duration) *Service {
	return &Service{
		secret:        []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (s *Service) GenerateAccessToken(member *model.Member) (string, error) {
	if member == nil {
		return "", errors.New("member is nil")
	}

	now := time.Now()
	claims := &Claims{
		ID:     member.MemberID,
		Email:  member.Email,
		Name:   member.Name,
		Birth:  member.Birth,
		Gender: member.Gender,
		Status: member.Status,
		Role:   member.Role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.accessExpiry)),
			Subject:   fmt.Sprintf("%d", member.MemberID),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) GenerateRefreshToken(member *model.Member) (string, error) {
	if member == nil {
		return "", errors.New("member is nil")
	}

	now := time.Now()
	claims := &RefreshClaims{
		MemberID: member.MemberID,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.refreshExpiry)),
			Subject:   fmt.Sprintf("%d", member.MemberID),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (interface{}, error) {
		if token.Method != jwtv5.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (interface{}, error) {
		if token.Method != jwtv5.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
