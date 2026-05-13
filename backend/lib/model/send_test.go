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
	testSendTitle     = "Test Article"
	testSendSender    = "sender@example.com"
	testSendMsgID     = "msg-123"
	testSendProvider  = "mailjet"
	testSendBounced   = "email bounced: 550 5.7.1"
	testArticleID     = "test-article-id"
	testUserEmail     = "user@example.com"
	testNameSuccSend  = "successful send"
	testStatusSuccess = "success"
	testStatusFailed  = "failed"
)

func TestSend_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		send     Send
		wantJSON string
	}{
		{
			name: testNameSuccSend,
			send: Send{
				Account:       testAccount,
				ArticleID:     testArticleID,
				SentAt:        time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:         testSendTitle,
				DestEmail:     testUserEmail,
				Status:        testStatusSuccess,
				SenderEmail:   testSendSender,
				MessageID:     testSendMsgID,
				Provider:      testSendProvider,
				ErrorResponse: "",
			},
			wantJSON: `{"Account":"test-account","ArticleID":"test-article-id",` +
				`"SentAt":"2024-03-15T10:30:00Z","Title":"Test Article","DestEmail":"user@example.com",` +
				`"Status":"success","SenderEmail":"sender@example.com","MessageID":"msg-123",` +
				`"Provider":"mailjet","ErrorResponse":""}`,
		},
		{
			name: "failed send",
			send: Send{
				Account:       testAccount,
				ArticleID:     testArticleID,
				SentAt:        time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:         testSendTitle,
				DestEmail:     testUserEmail,
				Status:        testStatusFailed,
				SenderEmail:   testSendSender,
				MessageID:     "",
				Provider:      testSendProvider,
				ErrorResponse: "email bounced",
			},
			wantJSON: `{"Account":"test-account","ArticleID":"test-article-id",` +
				`"SentAt":"2024-03-15T10:30:00Z","Title":"Test Article","DestEmail":"user@example.com",` +
				`"Status":"failed","SenderEmail":"sender@example.com","MessageID":"",` +
				`"Provider":"mailjet","ErrorResponse":"email bounced"}`,
		},
		{
			name: "minimal send",
			send: Send{
				Account:   testAccount,
				ArticleID: testArticleID,
				SentAt:    time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			},
			wantJSON: `{"Account":"test-account","ArticleID":"test-article-id",` +
				`"SentAt":"2024-03-15T10:30:00Z","Title":"","DestEmail":"","Status":"",` +
				`"SenderEmail":"","MessageID":"","Provider":"","ErrorResponse":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.send)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled Send
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.send.Account, unmarshaled.Account)
			assert.Equal(t, tt.send.ArticleID, unmarshaled.ArticleID)
			assert.True(t, tt.send.SentAt.Equal(unmarshaled.SentAt))
		})
	}
}

func TestSend_DynamoDBAttributeMapping(t *testing.T) {
	tests := []struct {
		name string
		send Send
	}{
		{
			name: testNameSuccSend,
			send: Send{
				Account:       testAccount,
				ArticleID:     testArticleID,
				SentAt:        time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:         testSendTitle,
				DestEmail:     testUserEmail,
				Status:        testStatusSuccess,
				SenderEmail:   testSendSender,
				MessageID:     testSendMsgID,
				Provider:      testSendProvider,
				ErrorResponse: "",
			},
		},
		{
			name: "failed send with error",
			send: Send{
				Account:       testAccount,
				ArticleID:     testArticleID,
				SentAt:        time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
				Title:         testSendTitle,
				DestEmail:     testUserEmail,
				Status:        testStatusFailed,
				SenderEmail:   testSendSender,
				MessageID:     "",
				Provider:      testSendProvider,
				ErrorResponse: testSendBounced,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := attributevalue.Marshal(tt.send)
			require.NoError(t, err)

			var unmarshaled Send
			err = attributevalue.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.send.Account, unmarshaled.Account)
			assert.Equal(t, tt.send.ArticleID, unmarshaled.ArticleID)
			assert.True(t, tt.send.SentAt.Equal(unmarshaled.SentAt))
			assert.Equal(t, tt.send.Title, unmarshaled.Title)
			assert.Equal(t, tt.send.DestEmail, unmarshaled.DestEmail)
			assert.Equal(t, tt.send.Status, unmarshaled.Status)
			assert.Equal(t, tt.send.SenderEmail, unmarshaled.SenderEmail)
			assert.Equal(t, tt.send.MessageID, unmarshaled.MessageID)
			assert.Equal(t, tt.send.Provider, unmarshaled.Provider)
			assert.Equal(t, tt.send.ErrorResponse, unmarshaled.ErrorResponse)
		})
	}
}

func TestSend_EmptyFields(t *testing.T) {
	send := Send{
		Account:   testAccount,
		ArticleID: testArticleID,
		SentAt:    time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(send)
	require.NoError(t, err)

	var unmarshaled Send
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Empty(t, unmarshaled.Title)
	assert.Empty(t, unmarshaled.DestEmail)
	assert.Empty(t, unmarshaled.Status)
	assert.Empty(t, unmarshaled.SenderEmail)
	assert.Empty(t, unmarshaled.MessageID)
	assert.Empty(t, unmarshaled.Provider)
	assert.Empty(t, unmarshaled.ErrorResponse)
}
