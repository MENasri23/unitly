package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unicode"
	"unitly/internal/domain"
	"unitly/internal/util/eql"
	"unitly/internal/util/hash"
	"unitly/internal/util/rand"
	"unitly/internal/util/validate"

	"github.com/google/uuid"
)

const (
	sendOtpTimeout      = 15 * time.Second
	resetPasswordExpiry = 5 * time.Minute
	refreshTokenExpiry  = 7 * 24 * time.Hour
)

var (
	authPurpose      = domain.OTPPurpose("register-login")
	resetPassPurpose = domain.OTPPurpose("password-reset")
)

type AuthResult struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	otpRepo        domain.OTPRepository
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	smsService     domain.SmsService
	tokenService   domain.TokenService
	otpRateLimiter domain.RateLimiter
	uow            domain.UnitOfWork
}

func NewAuthService(
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	smsService domain.SmsService,
	tokenService domain.TokenService,
	otpRateLimiter domain.RateLimiter,
	uow domain.UnitOfWork,
) *AuthService {
	return &AuthService{
		otpRepo:        otpRepo,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		smsService:     smsService,
		tokenService:   tokenService,
		otpRateLimiter: otpRateLimiter,
		uow:            uow,
	}
}

func (s *AuthService) RequestAuthByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	if !s.isValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.ensureOTPRequestAllowed(ctx, phone, authPurpose); err != nil {
		return "", err
	}

	return s.requestOTP(ctx, phone, authPurpose)
}

func (s *AuthService) VerifyAuthByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (*AuthResult, error) {
	req := &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  authPurpose,
	}

	result, err := s.otpRepo.Verify(ctx, req)
	if err == domain.ErrOTPAlreadyUsed {
		return nil, domain.ErrInvalidOTP
	}

	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByPhone(ctx, result.Phone)
	if err != nil {
		if !isUserNotFound(err) {
			return nil, err
		}

		user, err = s.registerUser(ctx, result.Phone)
		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		accessToken, err := s.generateAccessToken(ctx, user.GUID)
		if err != nil {
			return nil, err
		}

		refreshToken, err := s.createSession(ctx, user.GUID)
		if err != nil {
			return nil, err
		}

		result := &AuthResult{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		return result, nil
	}

	return nil, domain.ErrInternal
}

func (s *AuthService) AuthenticateByPassword(
	ctx context.Context,
	phone string,
	password string,
) (*AuthResult, error) {
	if !s.isValidPhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	if password == "" {
		return nil, domain.ErrInvalidPassword
	}

	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if isUserNotFound(err) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive || user.HashedPassword == nil || *user.HashedPassword == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if !eql.IsEqual(hash.SHA256(password), *user.HashedPassword) {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(ctx, user.GUID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createSession(ctx, user.GUID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RequestResetPasswordByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	if !validate.IsValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.ensureOTPRequestAllowed(ctx, phone, resetPassPurpose); err != nil {
		return "", err
	}

	_, err := s.userRepo.GetActiveUserIDByPhone(ctx, phone)
	if err == nil {
		return s.requestOTP(ctx, phone, resetPassPurpose)
	}

	// If the phone is not associated with an active user, do not reveal that fact.
	// Create an OTP record so timing/behaviour is consistent, but do not send an SMS.
	if !isUserNotFound(err) {
		return "", err
	}

	token, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", err
	}

	req := &domain.CreateOTPReq{
		Phone:   phone,
		Token:   token,
		Purpose: resetPassPurpose,
	}

	if _, err = s.otpRepo.Create(ctx, req); err != nil {
		s.reset(ctx, phone, resetPassPurpose)
		return "", err
	}

	// Return the generated token (same shape as success) without indicating existence.
	return token, nil
}

func (s *AuthService) VerifyResetPasswordByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (string, error) {
	req := &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  resetPassPurpose,
	}

	result, err := s.otpRepo.Verify(ctx, req)
	if err == domain.ErrOTPAlreadyUsed {
		return "", domain.ErrInvalidOTP
	}

	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByPhone(ctx, result.Phone)
	if err != nil {
		// Do not reveal whether the phone maps to a user. Return a generic OTP error for not-found.
		if isUserNotFound(err) {
			return "", domain.ErrInvalidOTP
		}
		return "", err
	}

	if !user.IsActive {
		return "", domain.ErrInvalidOTP
	}

	return s.generateResetPassToken(ctx, user.GUID)
}

func (s *AuthService) RefreshAuthByToken(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	tokenHash := hash.SHA256(refreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	user, err := s.userRepo.GetByGUID(ctx, session.UserID)
	if err != nil {
		if isUserNotFound(err) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrInvalidToken
	}

	newRefreshToken, err := rand.GenereteDefaultToken()
	if err != nil {
		return nil, err
	}

	newSession := &domain.Session{
		ID:               uuid.New(),
		UserID:           user.GUID,
		RefreshTokenHash: hash.SHA256(newRefreshToken),
		ExpiresAt:        time.Now().Add(refreshTokenExpiry),
	}

	if err := s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.sessionRepo.Revoke(txCtx, session.ID); err != nil {
			return err
		}

		return s.sessionRepo.Create(txCtx, newSession)
	}); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(ctx, user.GUID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) requestOTP(
	ctx context.Context,
	phone string,
	purpose domain.OTPPurpose,
) (string, error) {
	token, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", err
	}

	req := &domain.CreateOTPReq{
		Phone:   phone,
		Token:   token,
		Purpose: purpose,
	}
	otp, err := s.otpRepo.Create(ctx, req)
	if err != nil {
		s.reset(ctx, phone, purpose)
		return "", err
	}

	s.sendOtpAsync(ctx, phone, otp)

	return token, nil
}

func (s *AuthService) ensureOTPRequestAllowed(
	ctx context.Context,
	phone string,
	purpose domain.OTPPurpose,
) error {
	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil {
		return err
	}

	if blocked {
		return domain.ErrTooManyAttempts
	}

	if err = s.checkRatelimit(ctx, phone, purpose); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) checkRatelimit(ctx context.Context, phone string, purpose domain.OTPPurpose) error {
	ratelimitAllowed, err := s.otpRateLimiter.Allow(ctx, ratelimitKey(phone, purpose))
	if err != nil {
		return err
	}

	if !ratelimitAllowed {
		return domain.ErrRateLimit
	}

	return nil
}

func (s *AuthService) reset(ctx context.Context, phone string, purpose domain.OTPPurpose) {
	s.otpRateLimiter.Reset(ctx, ratelimitKey(phone, purpose))
}

func (s *AuthService) sendOtpAsync(
	ctx context.Context,
	phone string,
	otp string,
) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sendOtpTimeout)
		defer cancel()
		if err := s.smsService.SendOTP(bgCtx, phone, otp); err != nil {
			log.Printf("failed to send OTP to %s: %v", phone, err)
		}
	}()
}

func (s *AuthService) isValidPhone(phone string) bool {
	return validate.IsValidPhone(phone)
}

func isUserNotFound(err error) bool {
	return errors.Is(err, domain.ErrUserNotFound)
}

func (s *AuthService) registerUser(ctx context.Context, phone string) (*domain.User, error) {
	user, err := s.createUser(ctx, phone)
	if err == nil {
		return user, nil
	}

	// Possible race condition: user may have been created by another goroutine.
	// Attempt to retrieve the user. If successful, proceed; otherwise return original error.
	if errors.Is(err, domain.ErrAlreadyExist) {
		user, lookupErr := s.userRepo.GetByPhone(ctx, phone)
		if lookupErr == nil {
			return user, nil
		}
	}

	return nil, err
}

func (s *AuthService) createUser(ctx context.Context, phone string) (*domain.User, error) {
	guid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		GUID:     guid,
		Phone:    phone,
		IsActive: true,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) createSession(ctx context.Context, userID uuid.UUID) (string, error) {
	refreshToken, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", err
	}

	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: hash.SHA256(refreshToken),
		ExpiresAt:        time.Now().Add(refreshTokenExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *AuthService) generateAccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := domain.AccessTokenClaims{
		UserID: userID.String(),
	}
	return s.tokenService.GenerateAccessToken(ctx, claims)
}

func (s *AuthService) generateResetPassToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := domain.AccessTokenClaims{
		UserID: userID.String(),
		TokenClaims: domain.TokenClaims{
			Expiry: resetPasswordExpiry,
		},
	}
	return s.tokenService.GenerateResetPasswordToken(ctx, claims)
}

func ratelimitKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:ratelimit:%s", purpose, phone)
}

func (s *AuthService) FinalizeResetPasswordByOTP(
	ctx context.Context, sign, newPassword string,
) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	claims, err := s.tokenService.ValidateResetPasswordToken(ctx, sign)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return err
	}

	if s.uow == nil || s.sessionRepo == nil {
		return domain.ErrInternal
	}

	return s.uow.Do(ctx, func(ctx context.Context) error {
		if err := s.userRepo.UpdatePassword(ctx, userUUID, newPassword); err != nil {
			return err
		}

		if err := s.sessionRepo.RevokeAll(ctx, userUUID); err != nil {
			return err
		}

		return nil
	})
}

func validatePassword(password string) error {
	const (
		minPassLength = 8
		maxPassLength = 64
	)

	if len(password) < minPassLength || len(password) > maxPassLength {
		return domain.ErrInvalidPassword
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return domain.ErrInvalidPassword
	}

	return nil
}
