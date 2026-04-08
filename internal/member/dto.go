package member

type CreateGuestRequest struct {
	InterestIDs []int64 `json:"interestIds"`
}

type JoinRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Birth    int    `json:"birth"`
}

type UpdateInterestsRequest struct {
	InterestIDs []int64 `json:"interestIds"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UpdateProfileRequest struct {
	Name  *string `json:"name,omitempty"`
	Birth *int    `json:"birth,omitempty"`
}

type RegisterResponse struct {
	MemberID int64         `json:"memberId"`
	Token    TokenResponse `json:"token"`
}

type UpdateProfileResponse struct {
	Token TokenResponse `json:"token"`
}
