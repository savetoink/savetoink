package service

import (
	"context"
	"fmt"

	"github.com/shaftoe/savetoink/backend/lib/model"
)

// GetUserDeviceEmail delegates to UserProfileService.
func (s *Service) GetUserDeviceEmail(
	ctx context.Context,
	accountID string,
) (deviceEmail string, autoSend bool, err error) {
	deviceEmail, autoSend, err = s.profile.GetUserDeviceEmail(ctx, accountID)
	return
}

// SetUserDeviceEmail delegates to UserProfileService.
func (s *Service) SetUserDeviceEmail(ctx context.Context, accountID, deviceEmail string) error {
	if err := s.profile.SetUserDeviceEmail(ctx, accountID, deviceEmail); err != nil {
		return fmt.Errorf("failed to set user device email: %w", err)
	}
	return nil
}

// SetUserDeviceEmailWithAutoSend delegates to UserProfileService.
func (s *Service) SetUserDeviceEmailWithAutoSend(
	ctx context.Context,
	accountID, deviceEmail string,
	autoSend bool,
) error {
	if err := s.profile.SetUserDeviceEmailWithAutoSend(ctx, accountID, deviceEmail, autoSend); err != nil {
		return fmt.Errorf("failed to set user device email with auto send: %w", err)
	}
	return nil
}

// DeleteUserDeviceEmail delegates to UserProfileService.
func (s *Service) DeleteUserDeviceEmail(ctx context.Context, accountID string) error {
	if err := s.profile.DeleteUserDeviceEmail(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user device email: %w", err)
	}
	return nil
}

// GetUserProfile delegates to UserProfileService.
func (s *Service) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	userProfile, err := s.profile.GetUserProfile(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return userProfile, nil
}

// SetUserEmail delegates to UserProfileService.
func (s *Service) SetUserEmail(ctx context.Context, accountID, userEmail string) error {
	if err := s.profile.SetUserEmail(ctx, accountID, userEmail); err != nil {
		return fmt.Errorf("failed to set user email: %w", err)
	}
	return nil
}

// DeleteUserProfile delegates to UserProfileService.
func (s *Service) DeleteUserProfile(ctx context.Context, accountID string) error {
	if err := s.profile.DeleteUserProfile(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}
	return nil
}

// HandleBounce delegates to UserProfileService.
func (s *Service) HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error {
	if err := s.profile.HandleBounce(ctx, deviceEmail, errorMessage); err != nil {
		return fmt.Errorf("failed to handle bounce: %w", err)
	}
	return nil
}

// IsEmailBouncing delegates to UserProfileService.
func (s *Service) IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error) {
	bouncing, err := s.profile.IsEmailBouncing(ctx, accountID, deviceEmail)
	if err != nil {
		return false, fmt.Errorf("failed to check if email is bouncing: %w", err)
	}
	return bouncing, nil
}

// GetAccountIDByDeviceEmail delegates to UserProfileService.
func (s *Service) GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error) {
	accountID, err := s.profile.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get account id by device email: %w", err)
	}
	return accountID, nil
}
