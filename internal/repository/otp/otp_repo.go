package otprepo

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
	"unitly/internal/cache"
	"unitly/internal/domain"
	"unitly/internal/util/eql"
	"unitly/internal/util/rand"
)

var (
	rateLimitExpiry       = 1 * time.Minute
	otpExpiry             = 3 * time.Minute
	blockExpiry           = 10 * time.Minute
	otpLength             = 6
	maxAttempts     int64 = 3
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

func (o *otpRepository) IsRateLimited(ctx context.Context, phone string, purpose domain.OTPPurpose) (bool, error) {
	set, err := o.cache.SetIfNotExist(ctx, ratelimitKey(phone, purpose), 1, rateLimitExpiry)
	return !set, err
}

func (o *otpRepository) DeleteRateLimit(ctx context.Context, phone string, purpose domain.OTPPurpose) {
	o.cache.Delete(ctx, ratelimitKey(phone, purpose))
}

func (o *otpRepository) Create(ctx context.Context, phone string, purpose domain.OTPPurpose) (string, string, error) {
	token, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", "", err
	}

	otp, err := generateOTP(otpLength)
	if err != nil {
		return "", "", err
	}
	hashedOTP := o.hashOTP(otp)

	pipe := o.cache.Client.TxPipeline()
	pipe.Set(ctx, otpCodeKey(phone, purpose), hashedOTP, otpExpiry)
	pipe.Set(ctx, phoneKey(token, purpose), token, otpExpiry)
	pipe.Set(ctx, attemptsKey(phone, purpose), 0, otpExpiry)

	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", err
	}

	return otp, token, nil
}

func (o *otpRepository) GetPhoneByToken(ctx context.Context, token string, purpose domain.OTPPurpose) (string, error) {
	key := phoneKey(token, purpose)
	phone, err := o.cache.Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return phone, nil
}

func (o *otpRepository) Verify(ctx context.Context, r *domain.VerifyOTPReq) error {
	otpKey := otpCodeKey(r.Phone, r.Purpose)
	atmptsKey := attemptsKey(r.Phone, r.Purpose)
	blkKey := blockKey(r.Phone, r.Purpose)
	tokKey := phoneKey(r.Token, r.Purpose)

	storeHashedOTP, err := o.cache.Get(ctx, otpKey)
	if err != nil {
		return domain.ErrInvalidOTP
	}

	inputHashedOTP := o.hashOTP(r.InputOTP)
	match := eql.IsEqual(storeHashedOTP, inputHashedOTP)

	if !match {
		attempts, err := o.cache.Incr(ctx, atmptsKey)
		if err != nil {
			return err
		}

		if attempts >= maxAttempts {
			pipe := o.cache.Client.TxPipeline()
			pipe.Set(ctx, blkKey, 1, blockExpiry)
			pipe.Del(ctx, otpKey, tokKey)
			pipe.Exec(ctx)

			return domain.ErrTooManyAttempts
		}

		return domain.ErrInvalidOTP
	}

	deleted, err := o.cache.Delete(ctx, otpKey)
	if err != nil {
		return err
	}
	if !deleted {
		return domain.ErrOTPAlreadyUsed
	}

	o.cache.DeleteAll(ctx, tokKey, atmptsKey)

	return nil
}

// hashOTP creates an HMAC-SHA256 hash of the OTP for secure storage
func (o *otpRepository) hashOTP(otp string) string {
	h := hmac.New(sha256.New, o.hmacSecret)
	h.Write([]byte(otp))
	return hex.EncodeToString(h.Sum(nil))
}

func blockKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:block:%s", purpose, phone)
}

func ratelimitKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:ratelimit:%s", purpose, phone)
}

func otpCodeKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:code:%s", purpose, phone)
}

func phoneKey(token string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:phone:%s", purpose, token)
}

func attemptsKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:attempts:%s", purpose, phone)
}
