package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unitly/internal/domain"
	"unitly/internal/log"
)

const (
	timeout = 15 * time.Second
	baseUrl = "https://rest.payamak-panel.com/api/SendSMS/"
)

const (
	otpOp = "BaseServiceNumber"
)

type patternRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	To        string `json:"to"`
	Text      string `json:"text"`
	PatternID string `json:"bodyId"`
}

type response struct {
	Value        string `json:"Value"`
	RetStatus    int    `json:"RetStatus"`
	StrRetStatus string `json:"StrRetStatus"`
}

type meliPayamakService struct {
	client    *http.Client
	logger    log.Logger
	username  string
	apiKey    string
	otpTempID string
}

func NewMeliPayamakService(username, apiKey, otpTempID string, logger log.Logger) domain.SmsService {
	return &meliPayamakService{
		client:    newHTTPClient(),
		logger:    logger,
		username:  username,
		apiKey:    apiKey,
		otpTempID: otpTempID,
	}
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: timeout}
}

func (s *meliPayamakService) SendOTP(ctx context.Context, phone, code string) error {
	data := patternRequest{
		Username:  s.username,
		Password:  s.apiKey,
		To:        phone,
		Text:      code,
		PatternID: s.otpTempID,
	}

	req, err := createRequest(ctx, data, otpOp)
	if err != nil {
		s.logger.DebugCtx(ctx, "failed to create sms request", "phone", phone, "error", err)
		return err
	}

	s.logger.DebugCtx(ctx, "sending OTP", "phone", phone)
	res, err := s.client.Do(req)
	if err != nil {
		s.logger.DebugCtx(ctx, "failed to send OTP", "phone", phone, "error", err)
		return err
	}
	defer func() {
		// Drain and close to safely reuse the TCP connection via Keep-Alive
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	// Ensure we handle non-200 HTTP statuses safely
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err = fmt.Errorf("sms service returned unexpected status: %d", res.StatusCode)
		s.logger.DebugCtx(ctx, err.Error())
		return err
	}

	var apiResp response
	if err := json.NewDecoder(res.Body).Decode(&apiResp); err != nil {
		s.logger.DebugCtx(ctx, "failed to decode response from sms service", "phone", phone, "error", err)
		return err
	}

	s.logger.DebugCtx(ctx, "response from sms service", "phone", phone, "response", apiResp)

	return nil
}

func createRequest(ctx context.Context, data any, op string) (*http.Request, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return nil, fmt.Errorf("encode payload failed: %w", err)
	}

	url := baseUrl + op
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
