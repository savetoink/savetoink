package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountID = "account1"

type mockUserProfileRepository struct {
	profiles               map[string]*model.UserProfile
	getErr                 error
	putErr                 error
	deleteErr              error
	deleteDeviceEmailErr   error
	getAccountIDErr        error
	accountIDByDeviceEmail map[string]string
}

func newMockUserProfileRepository() *mockUserProfileRepository {
	return &mockUserProfileRepository{
		profiles:               make(map[string]*model.UserProfile),
		accountIDByDeviceEmail: make(map[string]string),
	}
}

func (m *mockUserProfileRepository) GetUserProfile(_ context.Context, account string) (*model.UserProfile, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.profiles[account], nil
}

func (m *mockUserProfileRepository) GetAccountIDByDeviceEmail(_ context.Context, deviceEmail string) (string, error) {
	if m.getAccountIDErr != nil {
		return "", m.getAccountIDErr
	}
	return m.accountIDByDeviceEmail[deviceEmail], nil
}

func (m *mockUserProfileRepository) PutUserProfile(_ context.Context, profile *model.UserProfile) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.profiles[profile.Account] = profile
	if profile.DeviceEmail != "" {
		m.accountIDByDeviceEmail[profile.DeviceEmail] = profile.Account
	}
	return nil
}

func (m *mockUserProfileRepository) DeleteUserProfile(_ context.Context, account string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.profiles, account)
	for email, acc := range m.accountIDByDeviceEmail {
		if acc == account {
			delete(m.accountIDByDeviceEmail, email)
		}
	}
	return nil
}

func (m *mockUserProfileRepository) DeleteUserDeviceEmail(_ context.Context, account string) error {
	if m.deleteDeviceEmailErr != nil {
		return m.deleteDeviceEmailErr
	}
	if profile, exists := m.profiles[account]; exists {
		oldEmail := profile.DeviceEmail
		profile.DeviceEmail = ""
		delete(m.accountIDByDeviceEmail, oldEmail)
	}
	return nil
}

func TestNew(t *testing.T) {
	mockRepo := newMockUserProfileRepository()
	svc := New(mockRepo)

	assert.NotNil(t, svc)
	assert.Equal(t, mockRepo, svc.repo)
}

func TestGetUserDeviceEmailAndAutoSend(t *testing.T) {
	t.Run("returns device email and auto send for existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
			AutoSend:    true,
		}
		svc := New(mockRepo)

		deviceEmail, autoSend, err := svc.GetUserDeviceEmailAndAutoSend(context.Background(), testAccountID)

		assert.NoError(t, err)
		assert.Equal(t, "test@kindle.com", deviceEmail)
		assert.True(t, autoSend)
	})

	t.Run("returns empty values for non-existent profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		deviceEmail, autoSend, err := svc.GetUserDeviceEmailAndAutoSend(context.Background(), "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, deviceEmail)
		assert.False(t, autoSend)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		_, _, err := svc.GetUserDeviceEmailAndAutoSend(context.Background(), testAccountID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})
}

func TestSetUserDeviceEmail(t *testing.T) {
	t.Run("creates new profile when none exists", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmail(context.Background(), testAccountID, "new@kindle.com")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.NotNil(t, profile)
		assert.Equal(t, testAccountID, profile.Account)
		assert.Equal(t, "new@kindle.com", profile.DeviceEmail)
	})

	t.Run("updates existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "old@kindle.com",
			AutoSend:    true,
		}
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmail(context.Background(), testAccountID, "new@kindle.com")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.Equal(t, "new@kindle.com", profile.DeviceEmail)
		assert.True(t, profile.AutoSend)
	})

	t.Run("returns error on get failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmail(context.Background(), testAccountID, "test@kindle.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})

	t.Run("returns error on put failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.putErr = errors.New("put error")
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmail(context.Background(), testAccountID, "test@kindle.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set user profile")
	})
}

func TestSetUserDeviceEmailWithAutoSend(t *testing.T) {
	t.Run("creates new profile with valid device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "test@kindle.com", true)

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.NotNil(t, profile)
		assert.Equal(t, "test@kindle.com", profile.DeviceEmail)
		assert.True(t, profile.AutoSend)
	})

	t.Run("creates new profile with all valid device domains", func(t *testing.T) {
		domains := []string{"@kindle.com", "@free.kindle.com", "@send.kobo.com", "@pbsync.com", "@mytolino.com"}

		for _, domain := range domains {
			mockRepo := newMockUserProfileRepository()
			svc := New(mockRepo)
			email := "test" + domain

			err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, email, false)

			assert.NoError(t, err)
			profile := mockRepo.profiles[testAccountID]
			assert.NotNil(t, profile)
			assert.Equal(t, email, profile.DeviceEmail)
			assert.False(t, profile.AutoSend)
		}
	})

	t.Run("updates existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "old@kindle.com",
			AutoSend:    false,
		}
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "new@kindle.com", true)

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.Equal(t, "new@kindle.com", profile.DeviceEmail)
		assert.True(t, profile.AutoSend)
	})

	t.Run("returns error for invalid device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "invalid@gmail.com", true)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error for empty device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "", true)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error for malformed email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "not-an-email", true)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error on get failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "test@kindle.com", true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})

	t.Run("returns error on put failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.putErr = errors.New("put error")
		svc := New(mockRepo)

		err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), testAccountID, "test@kindle.com", true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set user profile")
	})
}

func TestDeleteUserDeviceEmail(t *testing.T) {
	t.Run("successfully deletes device email from existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
			AutoSend:    true,
		}
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		svc := New(mockRepo)

		err := svc.DeleteUserDeviceEmail(context.Background(), testAccountID)

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.Equal(t, "", profile.DeviceEmail)
		_, exists := mockRepo.accountIDByDeviceEmail["test@kindle.com"]
		assert.False(t, exists)
	})

	t.Run("handles deletion when profile has no device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "",
			AutoSend:    false,
		}
		svc := New(mockRepo)

		err := svc.DeleteUserDeviceEmail(context.Background(), testAccountID)

		assert.NoError(t, err)
	})

	t.Run("handles deletion when profile does not exist", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.DeleteUserDeviceEmail(context.Background(), "nonexistent")

		assert.NoError(t, err)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.deleteDeviceEmailErr = errors.New("delete error")
		svc := New(mockRepo)

		err := svc.DeleteUserDeviceEmail(context.Background(), testAccountID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete user device email")
	})
}

func TestGetUserProfile(t *testing.T) {
	t.Run("returns existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		expectedProfile := &model.UserProfile{
			Account:     testAccountID,
			Email:       "user@example.com",
			DeviceEmail: "test@kindle.com",
			AutoSend:    true,
		}
		mockRepo.profiles[testAccountID] = expectedProfile
		svc := New(mockRepo)

		profile, err := svc.GetUserProfile(context.Background(), testAccountID)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, profile)
	})

	t.Run("returns nil for non-existent profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		profile, err := svc.GetUserProfile(context.Background(), "nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, profile)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		_, err := svc.GetUserProfile(context.Background(), testAccountID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})
}

func TestSetUserEmail(t *testing.T) {
	t.Run("creates new profile with valid email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "user@example.com")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.NotNil(t, profile)
		assert.Equal(t, "user@example.com", profile.Email)
	})

	t.Run("updates existing profile email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account: testAccountID,
			Email:   "old@example.com",
		}
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "new@example.com")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.Equal(t, "new@example.com", profile.Email)
	})

	t.Run("returns error for invalid email format", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "invalid-email")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error for empty email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error for email with angle brackets", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "User <user@example.com>")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalid)
	})

	t.Run("returns error on get failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "user@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})

	t.Run("returns error on put failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.putErr = errors.New("put error")
		svc := New(mockRepo)

		err := svc.SetUserEmail(context.Background(), testAccountID, "user@example.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set user profile")
	})
}

func TestDeleteUserProfile(t *testing.T) {
	t.Run("successfully deletes existing profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
		}
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		svc := New(mockRepo)

		err := svc.DeleteUserProfile(context.Background(), testAccountID)

		assert.NoError(t, err)
		_, exists := mockRepo.profiles[testAccountID]
		assert.False(t, exists)
		_, emailExists := mockRepo.accountIDByDeviceEmail["test@kindle.com"]
		assert.False(t, emailExists)
	})

	t.Run("handles deletion of non-existent profile", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.DeleteUserProfile(context.Background(), "nonexistent")

		assert.NoError(t, err)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.deleteErr = errors.New("delete error")
		svc := New(mockRepo)

		err := svc.DeleteUserProfile(context.Background(), testAccountID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete user profile")
	})
}

func TestHandleBounce(t *testing.T) {
	t.Run("records first bounce as hard bounce", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
		}
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error message")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.NotNil(t, profile.BouncedEmails)
		bounceInfo, exists := profile.BouncedEmails["test@kindle.com"]
		assert.True(t, exists)
		assert.Equal(t, "bounce error message", bounceInfo.Error)
		assert.False(t, bounceInfo.Timestamp.IsZero())
	})

	t.Run("updates existing bounce info for repeated bounce", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
			BouncedEmails: map[string]model.BounceInfo{
				"test@kindle.com": {
					Timestamp: time.Now().UTC().Add(-1 * time.Hour),
					Error:     "old error",
				},
			},
		}
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "new error message")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		bounceInfo := profile.BouncedEmails["test@kindle.com"]
		assert.Equal(t, "new error message", bounceInfo.Error)
	})

	t.Run("initializes bounced emails map when nil", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:       testAccountID,
			DeviceEmail:   "test@kindle.com",
			BouncedEmails: nil,
		}
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error")

		assert.NoError(t, err)
		profile := mockRepo.profiles[testAccountID]
		assert.NotNil(t, profile.BouncedEmails)
		assert.Len(t, profile.BouncedEmails, 1)
	})

	t.Run("returns not found error for unregistered device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "unknown@kindle.com", "bounce error")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("returns not found error when profile does not exist", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("returns error when finding account fails", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getAccountIDErr = errors.New("database error")
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find account")
	})

	t.Run("returns error when getting profile fails", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})

	t.Run("returns error when updating profile fails", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
		}
		mockRepo.putErr = errors.New("put error")
		svc := New(mockRepo)

		err := svc.HandleBounce(context.Background(), "test@kindle.com", "bounce error")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user profile")
	})
}

func TestIsEmailBouncing(t *testing.T) {
	t.Run("returns true for bouncing email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
			BouncedEmails: map[string]model.BounceInfo{
				"test@kindle.com": {
					Timestamp: time.Now().UTC(),
					Error:     "bounce error",
				},
			},
		}
		svc := New(mockRepo)

		isBouncing, err := svc.IsEmailBouncing(context.Background(), testAccountID, "test@kindle.com")

		assert.NoError(t, err)
		assert.True(t, isBouncing)
	})

	t.Run("returns false for non-bouncing email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
			BouncedEmails: map[string]model.BounceInfo{
				"other@kindle.com": {
					Timestamp: time.Now().UTC(),
					Error:     "bounce error",
				},
			},
		}
		svc := New(mockRepo)

		isBouncing, err := svc.IsEmailBouncing(context.Background(), testAccountID, "test@kindle.com")

		assert.NoError(t, err)
		assert.False(t, isBouncing)
	})

	t.Run("returns false when profile has no bounced emails", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:       testAccountID,
			DeviceEmail:   "test@kindle.com",
			BouncedEmails: nil,
		}
		svc := New(mockRepo)

		isBouncing, err := svc.IsEmailBouncing(context.Background(), testAccountID, "test@kindle.com")

		assert.NoError(t, err)
		assert.False(t, isBouncing)
	})

	t.Run("returns false when profile does not exist", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		isBouncing, err := svc.IsEmailBouncing(context.Background(), "nonexistent", "test@kindle.com")

		assert.NoError(t, err)
		assert.False(t, isBouncing)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getErr = errors.New("database error")
		svc := New(mockRepo)

		_, err := svc.IsEmailBouncing(context.Background(), testAccountID, "test@kindle.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user profile")
	})
}

func TestGetAccountIDByDeviceEmail(t *testing.T) {
	t.Run("returns account ID for existing device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		svc := New(mockRepo)

		accountID, err := svc.GetAccountIDByDeviceEmail(context.Background(), "test@kindle.com")

		assert.NoError(t, err)
		assert.Equal(t, testAccountID, accountID)
	})

	t.Run("returns empty string for non-existent device email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)

		accountID, err := svc.GetAccountIDByDeviceEmail(context.Background(), "unknown@kindle.com")

		assert.NoError(t, err)
		assert.Empty(t, accountID)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.getAccountIDErr = errors.New("database error")
		svc := New(mockRepo)

		_, err := svc.GetAccountIDByDeviceEmail(context.Background(), "test@kindle.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get account ID by device email")
	})
}

func TestValidateDeviceEmail(t *testing.T) {
	t.Run("accepts valid device email domains", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)
		ctx := context.Background()

		domains := []string{"@kindle.com", "@free.kindle.com", "@send.kobo.com", "@pbsync.com", "@mytolino.com"}

		for _, domain := range domains {
			email := "user" + domain
			err := svc.SetUserDeviceEmailWithAutoSend(ctx, testAccountID, email, true)
			assert.NoError(t, err, "should accept email with domain %s", domain)
		}
	})

	t.Run("rejects invalid device email domains", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)
		ctx := context.Background()

		invalidEmails := []string{
			"user@gmail.com",
			"user@yahoo.com",
			"user@outlook.com",
		}

		for _, email := range invalidEmails {
			err := svc.SetUserDeviceEmailWithAutoSend(ctx, testAccountID, email, true)
			assert.Error(t, err, "should reject email %s", email)
			assert.ErrorIs(t, err, apperrors.ErrInvalid)
		}
	})

	t.Run("rejects malformed email addresses", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)
		ctx := context.Background()

		invalidEmails := []string{
			"not-an-email",
			"@kindle.com",
			"user@",
			"",
		}

		for _, email := range invalidEmails {
			err := svc.SetUserDeviceEmailWithAutoSend(ctx, testAccountID, email, true)
			assert.Error(t, err, "should reject malformed email %s", email)
			assert.ErrorIs(t, err, apperrors.ErrInvalid)
		}
	})
}

func TestIntegration_SetUserDeviceEmailWithAutoSendAndGet(t *testing.T) {
	t.Run("round trip: set and get device email with auto send", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)
		ctx := context.Background()

		err := svc.SetUserDeviceEmailWithAutoSend(ctx, testAccountID, "test@kindle.com", true)
		require.NoError(t, err)

		deviceEmail, autoSend, err := svc.GetUserDeviceEmailAndAutoSend(ctx, testAccountID)
		require.NoError(t, err)

		assert.Equal(t, "test@kindle.com", deviceEmail)
		assert.True(t, autoSend)
	})
}

func TestIntegration_SetUserEmailAndGet(t *testing.T) {
	t.Run("round trip: set and get user email", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		svc := New(mockRepo)
		ctx := context.Background()

		err := svc.SetUserEmail(ctx, testAccountID, "user@example.com")
		require.NoError(t, err)

		profile, err := svc.GetUserProfile(ctx, testAccountID)
		require.NoError(t, err)

		assert.Equal(t, "user@example.com", profile.Email)
	})
}

func TestIntegration_HandleBounceAndCheck(t *testing.T) {
	t.Run("round trip: handle bounce and check if bouncing", func(t *testing.T) {
		mockRepo := newMockUserProfileRepository()
		mockRepo.accountIDByDeviceEmail["test@kindle.com"] = testAccountID
		mockRepo.profiles[testAccountID] = &model.UserProfile{
			Account:     testAccountID,
			DeviceEmail: "test@kindle.com",
		}
		svc := New(mockRepo)
		ctx := context.Background()

		err := svc.HandleBounce(ctx, "test@kindle.com", "bounce error")
		require.NoError(t, err)

		isBouncing, err := svc.IsEmailBouncing(ctx, testAccountID, "test@kindle.com")
		require.NoError(t, err)

		assert.True(t, isBouncing)
	})
}
