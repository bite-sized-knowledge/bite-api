package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.resend.com"

type Client struct {
	apiKey     string
	from       string
	baseURL    string
	httpClient *http.Client
}

type sendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

type sendEmailResponse struct {
	ID string `json:"id"`
}

func NewClient(apiKey, from string) *Client {
	return &Client{
		apiKey:  apiKey,
		from:    from,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewClientWithBaseURL(apiKey, from, baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{apiKey: apiKey, from: from, baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

func (c *Client) SendVerificationEmail(to, code, verifyURL string) error {
	return c.send(to, "[Bite Knowledge] 이메일 인증 메일입니다.",
		fmt.Sprintf("이메일 인증을 완료하려면 아래 링크를 열어주세요.\n\n인증 코드: %s\n인증 링크: %s", code, verifyURL),
		fmt.Sprintf(`<div><h1>Bite 이메일 인증</h1><p>아래 버튼을 눌러 이메일 인증을 완료해 주세요.</p><p><strong>인증 코드:</strong> %s</p><p><a href="%s">이메일 인증하기</a></p></div>`, code, verifyURL),
	)
}

func (c *Client) SendPasswordResetEmail(to, code, verifyURL string) error {
	return c.send(to, "[Bite Knowledge] 비밀번호 초기화 메일입니다.",
		fmt.Sprintf("비밀번호 초기화를 진행하려면 아래 링크를 열어주세요.\n\n인증 코드: %s\n초기화 링크: %s", code, verifyURL),
		fmt.Sprintf(`<div><h1>Bite 비밀번호 초기화</h1><p>아래 버튼을 눌러 비밀번호 초기화를 진행해 주세요.</p><p><strong>인증 코드:</strong> %s</p><p><a href="%s">비밀번호 초기화하기</a></p></div>`, code, verifyURL),
	)
}

func (c *Client) SendTemporaryPassword(to, tempPassword string) error {
	return c.send(to, "[Bite Knowledge] 임시 비밀번호가 발급되었습니다.",
		fmt.Sprintf("임시 비밀번호는 아래와 같습니다. 로그인 후 반드시 비밀번호를 변경해 주세요.\n\n임시 비밀번호: %s", tempPassword),
		fmt.Sprintf(`<div><h1>Bite 임시 비밀번호 발급</h1><p>로그인 후 반드시 비밀번호를 변경해 주세요.</p><p><strong>임시 비밀번호:</strong> %s</p></div>`, tempPassword),
	)
}

func (c *Client) send(to, subject, textBody, htmlBody string) error {
	if strings.TrimSpace(c.apiKey) == "" || strings.TrimSpace(c.from) == "" {
		return errors.New("email client is not configured")
	}
	payload := sendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("resend request failed with status %d", resp.StatusCode)
	}

	result := sendEmailResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if strings.TrimSpace(result.ID) == "" {
		return errors.New("resend response missing email id")
	}
	return nil
}
