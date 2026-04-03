package member

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/bite-sized/bite-api/internal/model"
	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       *Repository
	jwtService *jwtpkg.Service
}

func NewService(repo *Repository, jwtService *jwtpkg.Service) *Service {
	return &Service{repo: repo, jwtService: jwtService}
}

func (s *Service) CreateGuestMember(req CreateGuestRequest) (*RegisterResponse, error) {
	if len(req.InterestIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one interest is required", model.ErrBadRequest)
	}
	name, err := s.getAvailableName()
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.CreateGuest(name)
	if err != nil {
		return nil, err
	}
	for _, interestID := range req.InterestIDs {
		exists, err := s.repo.InterestExists(interestID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("%w: interest not found", model.ErrBadRequest)
		}
		if err := s.repo.AddMemberInterest(memberID, interestID); err != nil {
			return nil, err
		}
	}
	memberRecord, err := s.repo.FindMemberByID(memberID)
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return nil, err
	}
	return &RegisterResponse{MemberID: memberID, Token: TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}}, nil
}

func (s *Service) JoinMember(memberID int64, req JoinRequest) (*RegisterResponse, error) {
	memberRecord, err := s.repo.FindMemberByID(memberID)
	if err != nil {
		return nil, err
	}
	if memberRecord == nil {
		return nil, fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	if memberRecord.Role == "ROLE_USER" || memberRecord.Role == "ROLE_ADMIN" {
		return nil, fmt.Errorf("%w: already joined member", model.ErrBadRequest)
	}
	verified, err := s.repo.IsEmailVerified(strings.TrimSpace(req.Email), memberID)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, fmt.Errorf("%w: email not verified", model.ErrBadRequest)
	}
	duplicate, err := s.repo.ExistsByEmail(strings.TrimSpace(req.Email))
	if err != nil {
		return nil, err
	}
	if duplicate {
		return nil, fmt.Errorf("%w: email already exists", model.ErrBadRequest)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := s.repo.JoinMember(memberID, strings.TrimSpace(req.Email), string(hash), req.Birth); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteEmailVerification(strings.TrimSpace(req.Email)); err != nil {
		return nil, err
	}
	memberRecord, err = s.repo.FindMemberByID(memberID)
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return nil, err
	}
	return &RegisterResponse{MemberID: memberID, Token: TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}}, nil
}

func (s *Service) HasDuplicateName(name string) (bool, error) {
	return s.repo.ExistsByName(strings.TrimSpace(name))
}

func (s *Service) DeleteMember(currentMemberID, memberID int64) error {
	if currentMemberID != memberID {
		return fmt.Errorf("%w: member id mismatch", model.ErrBadRequest)
	}
	memberRecord, err := s.repo.FindMemberByID(memberID)
	if err != nil {
		return err
	}
	if memberRecord == nil {
		return fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	return s.repo.SoftDeleteMember(memberID)
}

func (s *Service) getAvailableName() (string, error) {
	for i := 0; i < 20; i++ {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		candidate := fmt.Sprintf("User@%08x", binary.BigEndian.Uint32(b))
		exists, err := s.repo.ExistsByName(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: failed to generate available name", model.ErrConflict)
}
