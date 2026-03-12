// Package profile provides user profile and device email management.
package profile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/logging"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/repository"
	"github.com/shaftoe/savetoink/backend/lib/validation"
)

// UserProfileService manages user profiles and device email settings.
type UserProfileService struct {
	repo repository.UserProfileRepository
}

// New creates a new UserProfileService instance.
func New(repo repository.UserProfileRepository) *UserProfileService {
	return &UserProfileService{
		repo: repo,
	}
}

// GetUserDeviceEmailAndAutoSend retrieves the user's device email and auto-send preference.
func (s *UserProfileService) GetUserDeviceEmailAndAutoSend(
	ctx context.Context,
	accountID string,
) (deviceEmail string, autoSend bool, err error) {
	if s.repo == nil {
		return "", false, errors.New("user profile repository not configured")
	}

	profile, err := s.repo.GetUserProfile(ctx, accountID)
	if err != nil {
		return "", false, fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return "", false, nil
	}

	return profile.DeviceEmail, profile.AutoSend, nil
}

// SetUserDeviceEmail sets the user's device email address.
func (s *UserProfileService) SetUserDeviceEmail(ctx context.Context, accountID, deviceEmail string) error {
	if s.repo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.repo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.DeviceEmail = deviceEmail

	if putErr := s.repo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	return nil
}

// SetUserDeviceEmailWithAutoSend sets the user's device email and auto-send preference.
func (s *UserProfileService) SetUserDeviceEmailWithAutoSend(
	ctx context.Context,
	accountID, deviceEmail string,
	autoSend bool,
) error {
	if err := s.validateDeviceEmail(deviceEmail); err != nil {
		return fmt.Errorf("%w: %s", apperrors.ErrInvalid, err.Error())
	}

	if s.repo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.repo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	oldEmail := ""
	if profile != nil {
		oldEmail = profile.DeviceEmail
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.DeviceEmail = deviceEmail
	profile.AutoSend = autoSend

	if putErr := s.repo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	logging.AddLogAttr(ctx, slog.String("old_device_email", oldEmail))
	logging.AddLogAttr(ctx, slog.String("new_device_email", deviceEmail))
	logging.AddLogAttr(ctx, slog.Bool("auto_send", autoSend))

	return nil
}

// DeleteUserDeviceEmail removes the user's device email.
func (s *UserProfileService) DeleteUserDeviceEmail(ctx context.Context, accountID string) error {
	if s.repo == nil {
		return errors.New("user profile repository not configured")
	}

	oldDeviceEmail, _, _ := s.GetUserDeviceEmailAndAutoSend(ctx, accountID)

	if err := s.repo.DeleteUserDeviceEmail(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user device email: %w", err)
	}

	logging.AddLogAttr(ctx, slog.String("old_device_email", oldDeviceEmail))
	logging.AddLogAttr(ctx, slog.String("action", "delete_device_email"))

	return nil
}

// GetUserProfile retrieves the user's profile.
func (s *UserProfileService) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	if s.repo == nil {
		return nil, errors.New("user profile repository not configured")
	}

	profile, err := s.repo.GetUserProfile(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return profile, nil
}

// SetUserEmail sets the user's email address.
func (s *UserProfileService) SetUserEmail(ctx context.Context, accountID, userEmail string) error {
	if err := validation.ValidateEmail(userEmail); err != nil {
		return fmt.Errorf("%w: %s", apperrors.ErrInvalid, err.Error())
	}

	if s.repo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.repo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.Email = userEmail

	if putErr := s.repo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	return nil
}

// DeleteUserProfile removes the user's profile.
func (s *UserProfileService) DeleteUserProfile(ctx context.Context, accountID string) error {
	if s.repo == nil {
		return errors.New("user profile repository not configured")
	}

	if err := s.repo.DeleteUserProfile(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	return nil
}

// HandleBounce records an email bounce for a device email.
func (s *UserProfileService) HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error {
	accountID, err := s.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return fmt.Errorf("failed to find account for device email %s: %w", deviceEmail, err)
	}

	if accountID == "" {
		return apperrors.ErrNotFound
	}

	profile, err := s.repo.GetUserProfile(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return apperrors.ErrNotFound
	}

	if profile.BouncedEmails == nil {
		profile.BouncedEmails = make(map[string]model.BounceInfo)
	}

	hardBounce := false
	if _, exists := profile.BouncedEmails[deviceEmail]; !exists {
		hardBounce = true
		profile.BouncedEmails[deviceEmail] = model.BounceInfo{
			Timestamp: time.Now().UTC(),
			Error:     errorMessage,
		}
	} else {
		existingBounce := profile.BouncedEmails[deviceEmail]
		profile.BouncedEmails[deviceEmail] = model.BounceInfo{
			Timestamp: existingBounce.Timestamp,
			Error:     errorMessage,
		}
	}

	if putErr := s.repo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to update user profile: %w", putErr)
	}

	logging.AddLogAttr(ctx, slog.String("bounced_email", deviceEmail))
	logging.AddLogAttr(ctx, slog.String("bounce_error", errorMessage))
	logging.AddLogAttr(ctx, slog.Bool("hard_bounce", hardBounce))
	logging.AddLogAttr(ctx, slog.Time("bounce_timestamp", time.Now().UTC()))

	return nil
}

// IsEmailBouncing checks if a device email is marked as bouncing.
func (s *UserProfileService) IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error) {
	if s.repo == nil {
		return false, errors.New("user profile repository not configured")
	}

	profile, err := s.repo.GetUserProfile(ctx, accountID)
	if err != nil {
		return false, fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return false, nil
	}

	if profile.BouncedEmails == nil {
		return false, nil
	}

	_, exists := profile.BouncedEmails[deviceEmail]
	return exists, nil
}

// GetAccountIDByDeviceEmail retrieves the account ID for a device email.
func (s *UserProfileService) GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error) {
	accountID, err := s.repo.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get account ID by device email: %w", err)
	}
	return accountID, nil
}

func (s *UserProfileService) validateDeviceEmail(email string) error {
	if err := validation.ValidateDeviceEmail(email); err != nil {
		return fmt.Errorf("invalid device email: %w", err)
	}
	return nil
}
