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

type meliPayamakService struct {
	client    *http.Client
	username  string
	apiKey    string
	otpTempID string
}

func NewMeliPayamakService(username, apiKey, otpTempID string) domain.SmsService {
	return &meliPayamakService{
		client:    newHTTPClient(),
		username:  username,
		apiKey:    apiKey,
		otpTempID: otpTempID,
	}
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: timeout}
}

func (s *meliPayamakService) SendOTP(ctx context.Context, phone string, code string) error {
	data := patternRequest{
		Username:  s.username,
		Password:  s.apiKey,
		To:        phone,
		Text:      code,
		PatternID: s.otpTempID,
	}

	res, err := makeRequest(ctx, s.client, data, otpOp)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println("value from melipayamak: ", string(response))

	return nil
}

func makeRequest(ctx context.Context, client *http.Client, data any, op string) (*http.Response, error) {
	jsonValue, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseUrl+op,
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}
