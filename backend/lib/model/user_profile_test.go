package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testProfileDeviceEmail = "device@example.com"
	testProfileBounced     = "email bounced: 550 5.7.1"
	testNameWithErr        = "with error"
	testNameWithoutErr     = "without error"
	testNameFullProfile    = "full profile with bounced emails"
	testNameMinProfile     = "minimal profile"
)

func TestBounceInfo_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		bounce   BounceInfo
		wantJSON string
	}{
		{
			name: testNameWithErr,
			bounce: BounceInfo{
				Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Error:     testProfileBounced,
			},
			wantJSON: `{"timestamp":"2024-03-15T10:30:00Z","error":"email bounced: 550 5.7.1"}`,
		},
		{
			name: testNameWithoutErr,
			bounce: BounceInfo{
				Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Error:     "",
			},
			wantJSON: `{"timestamp":"2024-03-15T10:30:00Z","error":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.bounce)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled BounceInfo
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.True(t, tt.bounce.Timestamp.Equal(unmarshaled.Timestamp))
			assert.Equal(t, tt.bounce.Error, unmarshaled.Error)
		})
	}
}

func TestBounceInfo_DynamoDBAttributeMapping(t *testing.T) {
	tests := []struct {
		name   string
		bounce BounceInfo
	}{
		{
			name: testNameWithErr,
			bounce: BounceInfo{
				Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Error:     testProfileBounced,
			},
		},
		{
			name: testNameWithoutErr,
			bounce: BounceInfo{
				Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Error:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := attributevalue.Marshal(tt.bounce)
			require.NoError(t, err)

			var unmarshaled BounceInfo
			err = attributevalue.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.True(t, tt.bounce.Timestamp.Equal(unmarshaled.Timestamp))
			assert.Equal(t, tt.bounce.Error, unmarshaled.Error)
		})
	}
}

func TestUserProfile_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		profile  UserProfile
		wantJSON string
	}{
		{
			name: testNameFullProfile,
			profile: UserProfile{
				Account:     testAccount,
				Email:       testUserEmail,
				DeviceEmail: testProfileDeviceEmail,
				AutoSend:    true,
				BouncedEmails: map[string]BounceInfo{
					testProfileDeviceEmail: {
						Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
						Error:     testProfileBounced,
					},
				},
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"device@example.com","auto_send":true,"bounced_emails":` +
				`{"device@example.com":{"timestamp":"2024-03-15T10:30:00Z",` +
				`"error":"email bounced: 550 5.7.1"}}}`,
		},
		{
			name: "profile without bounced emails (nil)",
			profile: UserProfile{
				Account:     testAccount,
				Email:       testUserEmail,
				DeviceEmail: testProfileDeviceEmail,
				AutoSend:    false,
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"device@example.com","auto_send":false,"bounced_emails":null}`,
		},
		{
			name: testNameMinProfile,
			profile: UserProfile{
				Account: testAccount,
				Email:   testUserEmail,
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"","auto_send":false,"bounced_emails":null}`,
		},
		{
			name: "profile with empty bounced emails map",
			profile: UserProfile{
				Account:       testAccount,
				Email:         testUserEmail,
				DeviceEmail:   testProfileDeviceEmail,
				AutoSend:      true,
				BouncedEmails: map[string]BounceInfo{},
			},
			wantJSON: `{"account":"test-account","email":"user@example.com",` +
				`"device_email":"device@example.com","auto_send":true,"bounced_emails":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.profile)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled UserProfile
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.profile.Account, unmarshaled.Account)
			assert.Equal(t, tt.profile.Email, unmarshaled.Email)
			assert.Equal(t, tt.profile.DeviceEmail, unmarshaled.DeviceEmail)
			assert.Equal(t, tt.profile.AutoSend, unmarshaled.AutoSend)
			assert.Equal(t, len(tt.profile.BouncedEmails), len(unmarshaled.BouncedEmails))
		})
	}
}

func TestUserProfile_DynamoDBAttributeMapping(t *testing.T) {
	tests := []struct {
		name    string
		profile UserProfile
	}{
		{
			name: testNameFullProfile,
			profile: UserProfile{
				Account:     testAccount,
				Email:       testUserEmail,
				DeviceEmail: testProfileDeviceEmail,
				AutoSend:    true,
				BouncedEmails: map[string]BounceInfo{
					testProfileDeviceEmail: {
						Timestamp: time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
						Error:     testProfileBounced,
					},
					"another@example.com": {
						Timestamp: time.Date(2024, 3, 16, 11, 30, 0, 0, time.UTC),
						Error:     "mailbox full",
					},
				},
			},
		},
		{
			name: "profile without bounced emails",
			profile: UserProfile{
				Account:     testAccount,
				Email:       testUserEmail,
				DeviceEmail: testProfileDeviceEmail,
				AutoSend:    false,
			},
		},
		{
			name: testNameMinProfile,
			profile: UserProfile{
				Account: testAccount,
				Email:   testUserEmail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := attributevalue.Marshal(tt.profile)
			require.NoError(t, err)

			var unmarshaled UserProfile
			err = attributevalue.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.profile.Account, unmarshaled.Account)
			assert.Equal(t, tt.profile.Email, unmarshaled.Email)
			assert.Equal(t, tt.profile.DeviceEmail, unmarshaled.DeviceEmail)
			assert.Equal(t, tt.profile.AutoSend, unmarshaled.AutoSend)
			assert.Equal(t, len(tt.profile.BouncedEmails), len(unmarshaled.BouncedEmails))

			for email, bounce := range tt.profile.BouncedEmails {
				unmarshaledBounce, exists := unmarshaled.BouncedEmails[email]
				require.True(t, exists, "bounced email should exist in unmarshaled data")
				assert.True(t, bounce.Timestamp.Equal(unmarshaledBounce.Timestamp))
				assert.Equal(t, bounce.Error, unmarshaledBounce.Error)
			}
		})
	}
}

func TestUserProfile_EmptyFields(t *testing.T) {
	profile := UserProfile{
		Account: testAccount,
		Email:   testUserEmail,
	}

	data, err := json.Marshal(profile)
	require.NoError(t, err)

	var unmarshaled UserProfile
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Empty(t, unmarshaled.DeviceEmail)
	assert.False(t, unmarshaled.AutoSend)
	assert.Nil(t, unmarshaled.BouncedEmails)
}

func TestUserProfile_NilBouncedEmails(t *testing.T) {
	profile := UserProfile{
		Account:       testAccount,
		Email:         testUserEmail,
		BouncedEmails: nil,
	}

	marshaled, err := attributevalue.Marshal(profile)
	require.NoError(t, err)

	var unmarshaled UserProfile
	err = attributevalue.Unmarshal(marshaled, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.BouncedEmails)
}

func TestUserProfile_BouncedEmailsEmptyMap(t *testing.T) {
	profile := UserProfile{
		Account:       testAccount,
		Email:         testUserEmail,
		BouncedEmails: map[string]BounceInfo{},
	}

	marshaled, err := attributevalue.Marshal(profile)
	require.NoError(t, err)

	var unmarshaled UserProfile
	err = attributevalue.Unmarshal(marshaled, &unmarshaled)
	require.NoError(t, err)

	assert.NotNil(t, unmarshaled.BouncedEmails)
	assert.Empty(t, unmarshaled.BouncedEmails)
}
