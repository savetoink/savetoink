package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/internal/model"
)

// HandleBounce processes a bounce event for a device email.
// It adds the email to the bounced emails list in the user's profile.
func (s *Service) HandleBounce(ctx context.Context, deviceEmail, errorMessage string) error {
	accountID, err := s.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return fmt.Errorf("failed to find account for device email %s: %w", deviceEmail, err)
	}

	if accountID == "" {
		return fmt.Errorf("no account found for device email %s", deviceEmail)
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return fmt.Errorf("profile not found for account %s", accountID)
	}

	if profile.BouncedEmails == nil {
		profile.BouncedEmails = make(map[string]model.BounceInfo)
	}

	if _, exists := profile.BouncedEmails[deviceEmail]; !exists {
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

	if putErr := s.userProfileRepo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to update user profile: %w", putErr)
	}

	return nil
}

// IsEmailBouncing checks if a device email is currently blocked due to bouncing.
func (s *Service) IsEmailBouncing(ctx context.Context, accountID, deviceEmail string) (bool, error) {
	if s.userProfileRepo == nil {
		return false, errors.New("user profile repository not configured")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
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

// GetAccountIDByDeviceEmail looks up an account ID by device email.
// This is used when processing bounce webhooks where we only have the device email.
func (s *Service) GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error) {
	accountID, err := s.userProfileRepo.GetAccountIDByDeviceEmail(ctx, deviceEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get account ID by device email: %w", err)
	}
	return accountID, nil
}
