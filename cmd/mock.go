package main

import (
	"context"
	"log"
	"time"
)

type smsServiceMock struct {
}

func (s *smsServiceMock) SendOTP(ctx context.Context, phone string, code string) error {
	log.Println("sending otp sms")
	time.Sleep(5 * time.Second)
	log.Printf("send sms with code=%s to %s", code, phone)
	return nil
}
