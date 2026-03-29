package jwt

import (
	"testing"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	birth := 1994
	service := NewService("test-secret", 15*time.Minute, 24*time.Hour)
	member := &model.Member{
		MemberID: 7,
		Email:    "test@example.com",
		Name:     "User@test",
		Birth:    &birth,
		Gender:   "NON_BINARY",
		Status:   "ACTIVE",
		Role:     "ROLE_USER",
	}

	token, err := service.GenerateAccessToken(member)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.Subject != "7" {
		t.Fatalf("expected subject 7, got %s", claims.Subject)
	}
	if claims.Email != member.Email || claims.Role != member.Role {
		t.Fatalf("claims mismatch: %+v", claims)
	}
	if claims.Birth == nil || *claims.Birth != birth {
		t.Fatalf("birth claim mismatch: %+v", claims.Birth)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	service := NewService("test-secret", 15*time.Minute, 24*time.Hour)
	member := &model.Member{MemberID: 11}

	token, err := service.GenerateRefreshToken(member)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	claims, err := service.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}

	if claims.MemberID != 11 {
		t.Fatalf("expected refresh member id 11, got %d", claims.MemberID)
	}
}
