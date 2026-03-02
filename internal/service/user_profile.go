// Package service provides user profile functionality.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
)

// GetUserDeviceEmail retrieves the device email and auto-send preference for a given account ID.
func (s *Service) GetUserDeviceEmail(
	ctx context.Context,
	accountID string,
) (deviceEmail string, autoSend bool, err error) {
	if s.userProfileRepo == nil {
		return "", false, errors.New("user profile repository not configured")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if err != nil {
		return "", false, fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return "", false, nil
	}

	return profile.DeviceEmail, profile.AutoSend, nil
}

// SetUserDeviceEmail sets the device email for a given account ID.
func (s *Service) SetUserDeviceEmail(ctx context.Context, accountID, deviceEmail string) error {
	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.DeviceEmail = deviceEmail

	if putErr := s.userProfileRepo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	return nil
}

// SetUserDeviceEmailWithAutoSend sets the device email and auto-send preference
// for a given account ID.
func (s *Service) SetUserDeviceEmailWithAutoSend(
	ctx context.Context,
	accountID, deviceEmail string,
	autoSend bool,
) error {
	if err := validateDeviceEmail(deviceEmail); err != nil {
		return err
	}

	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.DeviceEmail = deviceEmail
	profile.AutoSend = autoSend

	if putErr := s.userProfileRepo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	return nil
}

// DeleteUserDeviceEmail removes the device email and auto-send preference for a given account ID.
func (s *Service) DeleteUserDeviceEmail(ctx context.Context, accountID string) error {
	if err := s.userProfileRepo.DeleteUserDeviceEmail(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user device email: %w", err)
	}

	return nil
}

// GetUserProfile retrieves the full user profile for a given account ID.
func (s *Service) GetUserProfile(ctx context.Context, accountID string) (*model.UserProfile, error) {
	if s.userProfileRepo == nil {
		return nil, errors.New("user profile repository not configured")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return profile, nil
}

// SetUserEmail sets the email for a given account ID.
func (s *Service) SetUserEmail(ctx context.Context, accountID, userEmail string) error {
	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	profile, getErr := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if getErr != nil {
		return fmt.Errorf("failed to get user profile: %w", getErr)
	}

	if profile == nil {
		profile = &model.UserProfile{
			Account: accountID,
		}
	}

	profile.Email = userEmail

	if putErr := s.userProfileRepo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
	}

	return nil
}

// validateDeviceEmail validates that a device email address is properly formatted
// and has a valid domain.
func validateDeviceEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf(
			"invalid device email: must be a valid email ending with %s",
			consts.ValidDeviceEmailDomainsJoined(),
		)
	}

	domains := consts.GetValidDeviceEmailDomains()
	for _, domain := range domains {
		if strings.HasSuffix(addr.Address, domain) {
			return nil
		}
	}

	return fmt.Errorf(
		"invalid device email: must be a valid email ending with %s",
		consts.ValidDeviceEmailDomainsJoined(),
	)
}

// DeleteUserProfile deletes the user profile for a given account ID.
func (s *Service) DeleteUserProfile(ctx context.Context, accountID string) error {
	if s.userProfileRepo == nil {
		return errors.New("user profile repository not configured")
	}

	if err := s.userProfileRepo.DeleteUserProfile(ctx, accountID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	return nil
}
