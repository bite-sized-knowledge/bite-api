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
	subject := "[BITE] 이메일 인증을 완료해주세요"
	text := fmt.Sprintf(
		"BITE에 오신 것을 환영합니다!\n\n아래 버튼을 눌러 이메일 인증을 완료해주세요.\n\n인증 링크: %s\n\n버튼이 동작하지 않으면 인증 코드를 직접 입력하세요.\n인증 코드: %s\n\n본인이 요청하지 않았다면 이 메일을 무시해주세요.",
		verifyURL, code,
	)
	html := renderEmail(emailContent{
		Heading:    "이메일 인증을 완료해주세요",
		Intro:      "BITE에 오신 것을 환영합니다. 가입을 마무리하려면 아래 버튼을 눌러 이메일 인증을 완료해주세요.",
		ButtonText: "이메일 인증하기",
		ButtonURL:  verifyURL,
		CodeLabel:  "인증 코드",
		Code:       code,
		Footer:     "본인이 요청하지 않았다면 이 메일을 무시해주세요.",
	})
	return c.send(to, subject, text, html)
}

func (c *Client) SendPasswordResetEmail(to, code, verifyURL string) error {
	subject := "[BITE] 비밀번호 재설정 안내"
	text := fmt.Sprintf(
		"비밀번호 재설정을 진행하려면 아래 링크를 열어주세요.\n\n재설정 링크: %s\n\n버튼이 동작하지 않으면 인증 코드를 직접 입력하세요.\n인증 코드: %s\n\n본인이 요청하지 않았다면 이 메일을 무시해주세요.",
		verifyURL, code,
	)
	html := renderEmail(emailContent{
		Heading:    "비밀번호 재설정 안내",
		Intro:      "비밀번호 재설정을 요청하셨습니다. 아래 버튼을 눌러 절차를 진행해주세요. 인증 후 임시 비밀번호가 발급됩니다.",
		ButtonText: "비밀번호 재설정하기",
		ButtonURL:  verifyURL,
		CodeLabel:  "인증 코드",
		Code:       code,
		Footer:     "본인이 요청하지 않았다면 이 메일을 무시해주세요. 비밀번호는 변경되지 않습니다.",
	})
	return c.send(to, subject, text, html)
}

func (c *Client) SendTemporaryPassword(to, tempPassword string) error {
	subject := "[BITE] 임시 비밀번호 발급 안내"
	text := fmt.Sprintf(
		"임시 비밀번호가 발급되었습니다.\n\n임시 비밀번호: %s\n\n로그인 후 반드시 새로운 비밀번호로 변경해주세요.",
		tempPassword,
	)
	html := renderEmail(emailContent{
		Heading:   "임시 비밀번호가 발급되었습니다",
		Intro:     "아래 임시 비밀번호로 로그인한 뒤, 반드시 새로운 비밀번호로 변경해주세요.",
		CodeLabel: "임시 비밀번호",
		Code:      tempPassword,
		Footer:    "보안을 위해 임시 비밀번호는 첫 로그인 후 즉시 변경해주세요.",
	})
	return c.send(to, subject, text, html)
}

type emailContent struct {
	Heading    string
	Intro      string
	ButtonText string
	ButtonURL  string
	CodeLabel  string
	Code       string
	Footer     string
}

func renderEmail(c emailContent) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="ko"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>BITE</title></head><body style="margin:0;padding:0;background:#f6f7f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Apple SD Gothic Neo','Pretendard',sans-serif;color:#1a1a1a;">`)
	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f6f7f9;padding:32px 16px;"><tr><td align="center">`)
	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:520px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.06);">`)
	// Header band
	b.WriteString(`<tr><td style="padding:28px 32px 8px 32px;background:#fff;">`)
	b.WriteString(`<div style="display:inline-flex;align-items:center;gap:8px;"><span style="display:inline-block;width:32px;height:32px;background:#FF6B2C;color:#fff;font-weight:800;border-radius:8px;text-align:center;line-height:32px;font-size:18px;">B</span><span style="font-size:18px;font-weight:700;color:#1a1a1a;">BITE</span></div>`)
	b.WriteString(`</td></tr>`)
	// Heading
	b.WriteString(fmt.Sprintf(`<tr><td style="padding:24px 32px 0 32px;"><h1 style="margin:0;font-size:22px;font-weight:700;line-height:1.4;color:#1a1a1a;">%s</h1></td></tr>`, htmlEscape(c.Heading)))
	// Intro
	if c.Intro != "" {
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:12px 32px 0 32px;"><p style="margin:0;font-size:15px;line-height:1.6;color:#4a4a4a;">%s</p></td></tr>`, htmlEscape(c.Intro)))
	}
	// CTA button
	if c.ButtonURL != "" && c.ButtonText != "" {
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:24px 32px 0 32px;"><a href="%s" style="display:inline-block;background:#FF6B2C;color:#ffffff;text-decoration:none;padding:14px 28px;border-radius:10px;font-weight:600;font-size:15px;">%s</a></td></tr>`, htmlEscape(c.ButtonURL), htmlEscape(c.ButtonText)))
	}
	// Code box
	if c.Code != "" {
		label := c.CodeLabel
		if label == "" {
			label = "코드"
		}
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:24px 32px 0 32px;"><div style="background:#f6f7f9;border:1px solid #ececec;border-radius:10px;padding:16px;"><div style="font-size:12px;color:#7a7a7a;margin-bottom:6px;">%s</div><div style="font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:15px;color:#1a1a1a;word-break:break-all;">%s</div></div></td></tr>`, htmlEscape(label), htmlEscape(c.Code)))
	}
	// Footer text
	if c.Footer != "" {
		b.WriteString(fmt.Sprintf(`<tr><td style="padding:24px 32px 32px 32px;"><p style="margin:0;font-size:13px;line-height:1.6;color:#7a7a7a;">%s</p></td></tr>`, htmlEscape(c.Footer)))
	}
	b.WriteString(`</table>`)
	// Outside footer
	b.WriteString(`<div style="max-width:520px;margin:16px auto 0 auto;text-align:center;font-size:12px;color:#9a9a9a;line-height:1.6;">© BITE — 한입 크기 기술 지식<br>이 메일은 발신 전용입니다.</div>`)
	b.WriteString(`</td></tr></table></body></html>`)
	return b.String()
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
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
