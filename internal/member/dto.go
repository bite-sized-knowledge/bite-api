package member

type CreateGuestRequest struct {
	InterestIDs []int64 `json:"interestIds"`
}

type JoinRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Birth    int    `json:"birth"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RegisterResponse struct {
	MemberID int64         `json:"memberId"`
	Token    TokenResponse `json:"token"`
}
