package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unitly/internal/domain"
)

const (
	timeout  = 15 * time.Second
	apiKey   = ""
	baseUrl  = "https://rest.payamak-panel.com/api/SendSMS/"
	otpPatID = 468714
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
	client       *http.Client
	username     string
	apiKey       string
	otpPatternID int64
}

func NewMeliPayamakService(username string) domain.SmsService {
	return &meliPayamakService{
		client:       newHTTPClient(),
		username:     username,
		apiKey:       apiKey,
		otpPatternID: otpPatID,
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
		PatternID: strconv.FormatInt(s.otpPatternID, 10),
	}

	res, err := makeRequest(ctx, s.client, data, otpOp)
	if err != nil {
		return err
	}
	defer res.Body.Close()

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
