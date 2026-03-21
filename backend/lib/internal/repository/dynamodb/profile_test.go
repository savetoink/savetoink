package repository

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	profileTestEmail = "test@example.com"
)

func (s *DynamoDBRepositoryTestSuite) TestPutAndGetUserProfile() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Account:     "test-account",
		Email:       profileTestEmail,
		DeviceEmail: "device@example.com",
		AutoSend:    true,
		BouncedEmails: map[string]model.BounceInfo{
			"bounced@example.com": {
				Timestamp: time.Now(),
				Error:     "bounced",
			},
		},
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetUserProfile(ctx, profile.Account)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, profile.Account, retrieved.Account)
	assert.Equal(t, profile.Email, retrieved.Email)
	assert.Equal(t, profile.DeviceEmail, retrieved.DeviceEmail)
	assert.Equal(t, profile.AutoSend, retrieved.AutoSend)
	assert.Len(t, retrieved.BouncedEmails, 1)
}

func (s *DynamoDBRepositoryTestSuite) TestGetNonExistentUserProfile() {
	ctx := context.Background()
	t := s.T()

	profile, err := s.repositories.GetUserProfile(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, profile)
}

func (s *DynamoDBRepositoryTestSuite) TestGetAccountIDByDeviceEmail() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Account:     "test-account-device",
		Email:       profileTestEmail,
		DeviceEmail: "unique-device@example.com",
		AutoSend:    false,
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.NoError(t, err)

	accountID, err := s.repositories.GetAccountIDByDeviceEmail(ctx, "unique-device@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test-account-device", accountID)
}

func (s *DynamoDBRepositoryTestSuite) TestGetAccountIDByNonExistentDeviceEmail() {
	ctx := context.Background()
	t := s.T()

	accountID, err := s.repositories.GetAccountIDByDeviceEmail(ctx, "nonexistent@example.com")
	require.NoError(t, err)
	assert.Empty(t, accountID)
}

func (s *DynamoDBRepositoryTestSuite) TestDeleteUserProfile() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Account:     "test-account-delete",
		Email:       "delete@example.com",
		DeviceEmail: "delete-device@example.com",
		AutoSend:    false,
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.NoError(t, err)

	err = s.repositories.DeleteUserProfile(ctx, profile.Account)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetUserProfile(ctx, profile.Account)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func (s *DynamoDBRepositoryTestSuite) TestDeleteUserDeviceEmail() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Account:     "test-account-remove-device",
		Email:       "remove@example.com",
		DeviceEmail: "remove-device@example.com",
		AutoSend:    true,
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.NoError(t, err)

	err = s.repositories.DeleteUserDeviceEmail(ctx, profile.Account)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetUserProfile(ctx, profile.Account)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Empty(t, retrieved.DeviceEmail)
	assert.False(t, retrieved.AutoSend)
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateUserProfile() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Account:     "test-account-update",
		Email:       "update@example.com",
		DeviceEmail: "update-device@example.com",
		AutoSend:    false,
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.NoError(t, err)

	updatedProfile := &model.UserProfile{
		Account:     "test-account-update",
		Email:       "updated@example.com",
		DeviceEmail: "update-device@example.com",
		AutoSend:    true,
	}

	err = s.repositories.PutUserProfile(ctx, updatedProfile)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetUserProfile(ctx, profile.Account)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "updated@example.com", retrieved.Email)
	assert.True(t, retrieved.AutoSend)
}

func (s *DynamoDBRepositoryTestSuite) TestPutUserProfileEmptyAccount() {
	ctx := context.Background()
	t := s.T()

	profile := &model.UserProfile{
		Email:       "test@example.com",
		DeviceEmail: "device@example.com",
		AutoSend:    true,
	}

	err := s.repositories.PutUserProfile(ctx, profile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account field is required")
}

func (s *DynamoDBRepositoryTestSuite) TestDeleteUserProfileNonExistent() {
	ctx := context.Background()
	t := s.T()

	err := s.repositories.DeleteUserProfile(ctx, "nonexistent-account")
	require.NoError(t, err)
}

func (s *DynamoDBRepositoryTestSuite) TestDeleteUserDeviceEmailNonExistent() {
	ctx := context.Background()
	t := s.T()

	err := s.repositories.DeleteUserDeviceEmail(ctx, "nonexistent-account")
	require.NoError(t, err)
}
