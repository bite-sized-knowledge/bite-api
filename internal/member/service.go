package member

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bite-sized/bite-api/internal/model"
	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Service struct {
	repo       *Repository
	jwtService *jwtpkg.Service
}

func NewService(repo *Repository, jwtService *jwtpkg.Service) *Service {
	return &Service{repo: repo, jwtService: jwtService}
}

// IssueGuestForDevice resolves (or lazily creates) the guest member bound to
// device_id and returns a fresh JWT pair. Idempotent via UNIQUE(device_id);
// concurrent first-time inserts are reconciled by re-reading after a UNIQUE
// conflict.
//
// `created` 는 이 호출에서 새로 INSERT 된 경우만 true — middleware 가 이때만
// recsys 의 device→member bandit migration 을 fire-and-forget 으로 호출한다
// (재호출은 멱등이지만 매 FK action 마다 RPC 는 낭비).
func (s *Service) IssueGuestForDevice(deviceID string) (memberID int64, accessToken, refreshToken string, created bool, err error) {
	deviceID = strings.ToLower(strings.TrimSpace(deviceID))
	if !uuidRe.MatchString(deviceID) {
		return 0, "", "", false, fmt.Errorf("%w: invalid device_id", model.ErrBadRequest)
	}

	memberRecord, err := s.repo.FindByDeviceID(deviceID)
	if err != nil {
		return 0, "", "", false, err
	}
	if memberRecord == nil {
		name, nerr := s.getAvailableName()
		if nerr != nil {
			return 0, "", "", false, nerr
		}
		newID, insertErr := s.repo.CreateGuestWithDeviceID(name, deviceID)
		if insertErr != nil {
			existing, refetchErr := s.repo.FindByDeviceID(deviceID)
			if refetchErr != nil || existing == nil {
				return 0, "", "", false, insertErr
			}
			memberRecord = existing
		} else {
			memberRecord = &model.Member{
				MemberID: newID,
				Name:     name,
				Status:   "ACTIVE",
				Role:     "ROLE_GUEST",
			}
			created = true
		}
	}

	accessToken, err = s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return 0, "", "", false, err
	}
	refreshToken, err = s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return 0, "", "", false, err
	}
	return memberRecord.MemberID, accessToken, refreshToken, created, nil
}

func (s *Service) CreateGuestMember(req CreateGuestRequest) (*RegisterResponse, error) {
	if len(req.InterestIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one interest is required", model.ErrBadRequest)
	}
	allExist, err := s.repo.AllInterestsExist(req.InterestIDs)
	if err != nil {
		return nil, err
	}
	if !allExist {
		return nil, fmt.Errorf("%w: interest not found", model.ErrBadRequest)
	}
	name, err := s.getAvailableName()
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.CreateGuest(name)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceInterests(memberID, req.InterestIDs); err != nil {
		return nil, err
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

func (s *Service) RegisterMember(req JoinRequest) (*RegisterResponse, error) {
	email := strings.TrimSpace(req.Email)

	verified, err := s.repo.IsEmailVerified(email)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, fmt.Errorf("%w: email not verified", model.ErrBadRequest)
	}

	duplicate, err := s.repo.ExistsByEmail(email)
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

	name, err := s.getAvailableName()
	if err != nil {
		return nil, err
	}

	memberID, err := s.repo.CreateMember(email, string(hash), req.Birth, name)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteEmailVerification(email); err != nil {
		return nil, err
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

func (s *Service) UpdateInterests(memberID int64, interestIDs []int64) error {
	if len(interestIDs) == 0 {
		return fmt.Errorf("%w: at least one interest is required", model.ErrBadRequest)
	}
	allExist, err := s.repo.AllInterestsExist(interestIDs)
	if err != nil {
		return err
	}
	if !allExist {
		return fmt.Errorf("%w: interest not found", model.ErrBadRequest)
	}
	return s.repo.ReplaceInterests(memberID, interestIDs)
}

func (s *Service) GetInterests(memberID int64) ([]int64, error) {
	return s.repo.GetMemberInterestIDs(memberID)
}

func (s *Service) HasDuplicateName(name string) (bool, error) {
	return s.repo.ExistsByName(strings.TrimSpace(name))
}

func (s *Service) UpdateProfile(memberID int64, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
	if req.Name == nil && req.Birth == nil {
		return nil, fmt.Errorf("%w: at least one field is required", model.ErrBadRequest)
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name must not be empty", model.ErrBadRequest)
		}
		req.Name = &name

		exists, err := s.repo.ExistsByName(name)
		if err != nil {
			return nil, err
		}
		// Allow keeping the same name
		current, err := s.repo.FindMemberByID(memberID)
		if err != nil {
			return nil, err
		}
		if exists && (current == nil || current.Name != name) {
			return nil, fmt.Errorf("%w: name already exists", model.ErrConflict)
		}
	}

	if req.Birth != nil {
		year := *req.Birth
		if year < 1920 || year > time.Now().Year()-10 {
			return nil, fmt.Errorf("%w: invalid birth year", model.ErrBadRequest)
		}
	}

	if err := s.repo.UpdateProfile(memberID, req.Name, req.Birth); err != nil {
		return nil, err
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

	return &UpdateProfileResponse{
		Token: TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken},
	}, nil
}

func (s *Service) GetProfile(memberID int64) (*model.Member, error) {
	member, err := s.repo.FindMemberByID(memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, fmt.Errorf("%w: member not found", model.ErrBadRequest)
	}
	return member, nil
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
