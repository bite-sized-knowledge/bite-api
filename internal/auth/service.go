package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/bite-sized/bite-api/internal/member"
	"github.com/bite-sized/bite-api/internal/model"
	"github.com/bite-sized/bite-api/pkg/email"
	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	verifyTypeRegister      = "REGISTER"
	verifyTypePasswordReset = "PASSWORD_RESET"
)

type Service struct {
	repo       *Repository
	jwtService *jwtpkg.Service
	email      *email.Client
	appBaseURL string
}

func NewService(repo *Repository, jwtService *jwtpkg.Service, emailClient *email.Client, appBaseURL string) *Service {
	return &Service{
		repo:       repo,
		jwtService: jwtService,
		email:      emailClient,
		appBaseURL: appBaseURL,
	}
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	memberRecord, err := s.repo.FindMemberByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		return nil, err
	}
	if memberRecord == nil {
		return nil, fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(memberRecord.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("%w: password mismatch", model.ErrBadRequest)
	}
	if memberRecord.Status != "ACTIVE" && memberRecord.Status != "PENDING" {
		return nil, fmt.Errorf("%w: member is not active", model.ErrBadRequest)
	}
	accessToken, err := s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: member.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}}, nil
}

func (s *Service) Refresh(req RefreshRequest) (*member.TokenResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", model.ErrBadRequest)
	}
	memberRecord, err := s.repo.FindMemberByID(claims.MemberID)
	if err != nil {
		return nil, err
	}
	if memberRecord == nil {
		return nil, fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	accessToken, err := s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return nil, err
	}
	return &member.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) SendEmailVerification(emailAddress string, memberID int64) error {
	return s.sendVerification(strings.TrimSpace(emailAddress), memberID, verifyTypeRegister)
}

func (s *Service) SendPasswordResetEmail(emailAddress string) error {
	memberRecord, err := s.repo.FindMemberByEmail(strings.TrimSpace(emailAddress))
	if err != nil {
		return err
	}
	if memberRecord == nil {
		return fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	if memberRecord.Status != "ACTIVE" {
		return fmt.Errorf("%w: member is not active", model.ErrBadRequest)
	}
	return s.sendVerification(memberRecord.Email, memberRecord.MemberID, verifyTypePasswordReset)
}

func (s *Service) sendVerification(emailAddress string, memberID int64, verifyType string) error {
	code := fmt.Sprintf("%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(30 * time.Minute)
	if err := s.repo.UpsertEmailVerify(emailAddress, code, memberID, verifyType, expiresAt); err != nil {
		return err
	}
	verifyURL := fmt.Sprintf("%s/v1/auth/email/verify?code=%s&email=%s&type=%s", strings.TrimRight(s.appBaseURL, "/"), code, emailAddress, verifyType)
	if verifyType == verifyTypeRegister {
		return s.email.SendVerificationEmail(emailAddress, code, verifyURL)
	}
	return s.email.SendPasswordResetEmail(emailAddress, code, verifyURL)
}

func (s *Service) VerifyEmail(emailAddress, code, verifyType string) error {
	verifyRecord, err := s.repo.FindEmailVerify(emailAddress, verifyType)
	if err != nil {
		return err
	}
	if verifyRecord == nil {
		return fmt.Errorf("%w: email verification not found", model.ErrBadRequest)
	}
	if verifyRecord.IsVerified {
		return fmt.Errorf("%w: already verified", model.ErrBadRequest)
	}
	if verifyRecord.VerifyCode != code {
		return fmt.Errorf("%w: invalid verification code", model.ErrBadRequest)
	}
	if time.Now().After(verifyRecord.ExpiredAt) {
		return fmt.Errorf("%w: verification code expired", model.ErrBadRequest)
	}
	if err := s.repo.MarkEmailVerified(emailAddress, verifyType); err != nil {
		return err
	}
	if verifyType == verifyTypePasswordReset && verifyRecord.MemberID != nil {
		tempPassword := fmt.Sprintf("tmp-%06d", time.Now().Unix()%1000000)
		hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := s.repo.UpdateMemberPassword(*verifyRecord.MemberID, string(hash)); err != nil {
			return err
		}
		if err := s.email.SendTemporaryPassword(emailAddress, tempPassword); err != nil {
			return err
		}
		return s.repo.DeleteEmailVerify(emailAddress, verifyType)
	}
	return nil
}

func (s *Service) IsVerified(emailAddress string, memberID int64) (bool, error) {
	verifyRecord, err := s.repo.FindEmailVerify(strings.TrimSpace(emailAddress), verifyTypeRegister)
	if err != nil {
		return false, err
	}
	if verifyRecord == nil || verifyRecord.MemberID == nil {
		return false, nil
	}
	return verifyRecord.IsVerified && *verifyRecord.MemberID == memberID, nil
}

func (s *Service) ChangePassword(memberID int64, req PasswordChangeRequest) error {
	memberRecord, err := s.repo.FindMemberByID(memberID)
	if err != nil {
		return err
	}
	if memberRecord == nil {
		return fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(memberRecord.Password), []byte(req.CurrentPassword)); err != nil {
		return fmt.Errorf("%w: current password mismatch", model.ErrBadRequest)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdateMemberPassword(memberID, string(hash))
}

