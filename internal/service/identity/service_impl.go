package identity

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unitly/internal/domain"
	"unitly/internal/util/eql"
	"unitly/internal/util/hash"

	"github.com/google/uuid"
)

type identityService struct {
	userRepo domain.UserRepository
}

func NewIdentityService(userRepo domain.UserRepository) IdentityService {
	return &identityService{userRepo: userRepo}
}

func (s *identityService) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return s.userRepo.GetByPhone(ctx, phone)
}

func (s *identityService) HasPassword(ctx context.Context, phone string) (bool, error) {
	user, err := s.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}

	return user.IsActive && user.IsPasswordSet(), nil
}

func (s *identityService) GetOrCreateUser(ctx context.Context, phone string) (*domain.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	guid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	user = &domain.User{
		GUID:     guid,
		Phone:    phone,
		IsActive: true,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrAlreadyExist) {
			return s.userRepo.GetByPhone(ctx, phone)
		}
		return nil, err
	}

	return user, nil
}

func (s *identityService) Authenticate(ctx context.Context, phone, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive || user.HashedPassword == nil || *user.HashedPassword == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if !eql.IsEqual(hash.SHA256(password), *user.HashedPassword) {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

func (s *identityService) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	if err := s.validatePasswordComplexity(newPassword); err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, newPassword)
}

func (s *identityService) validatePasswordComplexity(password string) error {
	if len(password) < 8 || len(password) > 64 {
		return domain.ErrInvalidPassword
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
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

func (s *identityService) UpdateFullName(ctx context.Context, user *domain.User) error {
	trimString(user.FirstName)
	trimString(user.LastName)

	if isFullNameMissing(user) {
		return ErrEmptyFullName
	}

	return s.userRepo.UpdateFullName(ctx, user)
}

func trimString(s *string) {
	if s != nil {
		*s = strings.TrimSpace(*s)
	}
}

func isFullNameMissing(user *domain.User) bool {
	return isEmpty(user.FirstName) && isEmpty(user.LastName)
}

func isEmpty(s *string) bool {
	return s == nil || *s == ""
}
