package auth

import (
	"encoding/json"
	"testing"

	"github.com/bite-sized/bite-api/internal/member"
)

// TestOAuthLoginResponseShape verifies the OAuth handler JSON shape matches
// what the web client reads: data.token.accessToken and data.token.refreshToken.
// The old flat shape {accessToken, refreshToken, isNew} broke web OAuth login
// because oauthGithub/oauthGoogle read data.token.accessToken (from api.ts).
func TestOAuthLoginResponseShape(t *testing.T) {
	resp := OAuthLoginResponse{
		Token: member.TokenResponse{
			AccessToken:  "access-xyz",
			RefreshToken: "refresh-xyz",
		},
		IsNew: true,
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	tokenObj, ok := decoded["token"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested 'token' object, got: %s", raw)
	}
	if tokenObj["accessToken"] != "access-xyz" {
		t.Fatalf("expected token.accessToken=access-xyz, got: %v", tokenObj["accessToken"])
	}
	if tokenObj["refreshToken"] != "refresh-xyz" {
		t.Fatalf("expected token.refreshToken=refresh-xyz, got: %v", tokenObj["refreshToken"])
	}
	if decoded["isNew"] != true {
		t.Fatalf("expected isNew=true, got: %v", decoded["isNew"])
	}
	if _, stillFlat := decoded["accessToken"]; stillFlat {
		t.Fatalf("regression: flat accessToken present at top level: %s", raw)
	}
}
