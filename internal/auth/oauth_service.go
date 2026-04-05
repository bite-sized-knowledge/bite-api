package auth

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bite-sized/bite-api/internal/config"
	"github.com/bite-sized/bite-api/internal/member"
	"github.com/bite-sized/bite-api/internal/model"
	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"
)

type OAuthLoginResponse struct {
	Token member.TokenResponse `json:"token"`
	IsNew bool                 `json:"isNew"`
}

type OAuthService struct {
	cfg        *config.Config
	oauthRepo  *OAuthRepository
	memberRepo *member.Repository
	jwtService *jwtpkg.Service
}

func NewOAuthService(cfg *config.Config, oauthRepo *OAuthRepository, memberRepo *member.Repository, jwtService *jwtpkg.Service) *OAuthService {
	return &OAuthService{
		cfg:        cfg,
		oauthRepo:  oauthRepo,
		memberRepo: memberRepo,
		jwtService: jwtService,
	}
}

func (s *OAuthService) HandleGitHubLogin(code string) (*OAuthLoginResponse, error) {
	accessToken, err := s.exchangeGitHubCode(code)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to exchange GitHub code: %v", model.ErrBadRequest, err)
	}

	ghUser, err := s.fetchGitHubUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch GitHub user: %v", model.ErrBadRequest, err)
	}

	email, err := s.fetchGitHubEmail(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch GitHub email: %v", model.ErrBadRequest, err)
	}

	providerMemberID := fmt.Sprintf("%d", ghUser.ID)
	name := ghUser.Login
	if ghUser.Name != "" {
		name = ghUser.Name
	}

	return s.findOrCreateOAuthMember("GITHUB", providerMemberID, email, name)
}

func (s *OAuthService) HandleGoogleLogin(code string) (*OAuthLoginResponse, error) {
	accessToken, err := s.exchangeGoogleCode(code)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to exchange Google code: %v", model.ErrBadRequest, err)
	}

	gUser, err := s.fetchGoogleUser(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch Google user: %v", model.ErrBadRequest, err)
	}

	return s.findOrCreateOAuthMember("GOOGLE", gUser.ID, gUser.Email, gUser.Name)
}

func (s *OAuthService) findOrCreateOAuthMember(provider, providerMemberID, email, name string) (*OAuthLoginResponse, error) {
	oauthRecord, err := s.oauthRepo.FindByProviderAndProviderMemberID(provider, providerMemberID)
	if err != nil {
		return nil, err
	}

	isNew := false
	var memberRecord *model.Member

	if oauthRecord != nil {
		memberRecord, err = s.memberRepo.FindMemberByID(oauthRecord.MemberID)
		if err != nil {
			return nil, err
		}
		if memberRecord == nil {
			return nil, fmt.Errorf("%w: linked member not found", model.ErrBadRequest)
		}
	} else {
		isNew = true
		memberName, err := s.generateName()
		if err != nil {
			return nil, err
		}

		memberID, err := s.createOAuthMember(email, memberName)
		if err != nil {
			return nil, err
		}

		if err := s.oauthRepo.Create(memberID, provider, providerMemberID); err != nil {
			return nil, err
		}

		memberRecord, err = s.memberRepo.FindMemberByID(memberID)
		if err != nil {
			return nil, err
		}
	}

	accessToken, err := s.jwtService.GenerateAccessToken(memberRecord)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(memberRecord)
	if err != nil {
		return nil, err
	}

	return &OAuthLoginResponse{
		Token: member.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken},
		IsNew: isNew,
	}, nil
}

func (s *OAuthService) createOAuthMember(email, name string) (int64, error) {
	result, err := s.memberRepo.DB().Exec(
		`INSERT INTO member (email, name, status, role) VALUES (?, ?, 'ACTIVE', 'ROLE_USER')`,
		email, name,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *OAuthService) generateName() (string, error) {
	for i := 0; i < 20; i++ {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		candidate := fmt.Sprintf("User@%08x", binary.BigEndian.Uint32(b))
		exists, err := s.memberRepo.ExistsByName(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%w: failed to generate available name", model.ErrConflict)
}

// GitHub OAuth helpers

type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type gitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *OAuthService) exchangeGitHubCode(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.cfg.GitHubClientID)
	data.Set("client_secret", s.cfg.GitHubClientSecret)
	data.Set("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("github token error: %s", tokenResp.Error)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token from GitHub")
	}
	return tokenResp.AccessToken, nil
}

func (s *OAuthService) fetchGitHubUser(accessToken string) (*gitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user gitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *OAuthService) fetchGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []gitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found from GitHub")
}

// Google OAuth helpers

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

type googleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *OAuthService) exchangeGoogleCode(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.cfg.GoogleClientID)
	data.Set("client_secret", s.cfg.GoogleClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "postmessage")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("google token error: %s", tokenResp.Error)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token from Google")
	}
	return tokenResp.AccessToken, nil
}

func (s *OAuthService) fetchGoogleUser(accessToken string) (*googleUser, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user googleUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	if user.ID == "" {
		return nil, fmt.Errorf("empty user ID from Google")
	}
	return &user, nil
}
