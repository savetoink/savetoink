// Package service provides user profile functionality.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/shaftoe/savetoink/internal/model"
)

// GetUserKindleEmail retrieves the kindle email for a given account ID.
func (s *Service) GetUserKindleEmail(ctx context.Context, accountID string) (string, error) {
	if s.userProfileRepo == nil {
		return "", errors.New("user profile repository not configured")
	}

	profile, err := s.userProfileRepo.GetUserProfile(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}

	if profile == nil {
		return "", nil
	}

	return profile.KindleEmail, nil
}

// SetUserKindleEmail sets the kindle email for a given account ID.
func (s *Service) SetUserKindleEmail(ctx context.Context, accountID, kindleEmail string) error {
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

	profile.KindleEmail = kindleEmail

	if putErr := s.userProfileRepo.PutUserProfile(ctx, profile); putErr != nil {
		return fmt.Errorf("failed to set user profile: %w", putErr)
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
