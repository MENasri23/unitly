package otprepo

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"
	"unitly/internal/cache"
	"unitly/internal/domain"
	"unitly/internal/util/eql"
)

var (
	otpExpiry         = 3 * time.Minute
	blockExpiry       = 10 * time.Minute
	otpLength         = 6
	maxAttempts int64 = 3
)

type otpRepository struct {
	cache      *cache.Cache
	hmacSecret []byte
}

func NewOtpRepository(cache *cache.Cache, secret string) domain.OTPRepository {
	return &otpRepository{
		cache:      cache,
		hmacSecret: []byte(secret),
	}
}

func (o *otpRepository) IsBlocked(ctx context.Context, phone string, purpose domain.OTPPurpose) (bool, error) {
	return o.cache.ExistsKey(ctx, blockKey(phone, purpose))
}

func (o *otpRepository) Create(ctx context.Context, req *domain.CreateOTPReq) (string, error) {
	otp, err := generateOTP(otpLength)
	if err != nil {
		return "", err
	}

	otpHash := o.hashOTP(otp)
	session := &otpSession{
		Phone:    req.Phone,
		otpHash:  otpHash,
		Attempts: 0,
	}

	sessionKey := sessionKey(req.Token, req.Purpose)

	pipe := o.cache.Client.TxPipeline()
	pipe.HSet(
		ctx,
		sessionKey,
		session.serialize(),
	)
	pipe.Expire(ctx, sessionKey, otpExpiry)

	if _, err = pipe.Exec(ctx); err != nil {
		log.Printf("failed to create OTP session: %v", err)
		return "", err
	}

	return otp, nil
}

func (o *otpRepository) Verify(ctx context.Context, r *domain.VerifyOTPReq) (*domain.VerifyOTPResult, error) {
	sessionKey := sessionKey(r.Token, r.Purpose)
	sessionMap, err := o.cache.GetMap(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	if len(sessionMap) == 0 {
		return nil, domain.ErrInvalidOTP
	}

	session, err := deserializeOtpSession(sessionMap)
	if err != nil {
		return nil, err
	}

	blocked, err := o.IsBlocked(ctx, session.Phone, r.Purpose)
	if err != nil {
		return nil, err
	}

	if blocked {
		return nil, domain.ErrTooManyAttempts
	}

	inputHashedOTP := o.hashOTP(r.InputOTP)
	match := eql.IsEqual(session.otpHash, inputHashedOTP)

	if !match {
		attempts, err := o.incrementAttempts(ctx, sessionKey)
		if err != nil {
			return nil, err
		}

		if attempts >= maxAttempts {
			blockKey := blockKey(session.Phone, r.Purpose)
			if _, err = o.block(ctx, blockKey); err != nil {
				log.Printf("failed to block user after max OTP attempts: %v", err)
			}

			return nil, domain.ErrTooManyAttempts
		}

		return nil, domain.ErrInvalidOTP
	}

	deleted, err := o.cache.Delete(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, domain.ErrOTPAlreadyUsed
	}

	return &domain.VerifyOTPResult{
		Phone: session.Phone,
	}, nil
}

// hashOTP creates an HMAC-SHA256 hash of the OTP for secure storage
func (o *otpRepository) hashOTP(otp string) string {
	h := hmac.New(sha256.New, o.hmacSecret)
	h.Write([]byte(otp))
	return hex.EncodeToString(h.Sum(nil))
}

func (o *otpRepository) block(ctx context.Context, key string) (bool, error) {
	return o.cache.SetIfNotExist(ctx, key, 1, blockExpiry)
}

func (o *otpRepository) incrementAttempts(ctx context.Context, sessionKey string) (int64, error) {
	return o.cache.HIncrBy(ctx, sessionKey, otpAttemptsField, 1)
}

func blockKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:block:%s", purpose, phone)
}

func sessionKey(token string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:session:%s", purpose, token)
}
