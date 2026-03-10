package repository

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *DynamoDBRepositoryTestSuite) TestCreateAndGetSendRecord() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	send := &model.Send{
		Account:     "test-account",
		ArticleID:   "article-1",
		SentAt:      now,
		Title:       "Test Article",
		DestEmail:   "dest@example.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repositories.CreateSendRecord(ctx, send)
	require.NoError(t, err)

	sends, err := s.repositories.GetSendsByArticleID(ctx, send.ArticleID)
	require.NoError(t, err)
	require.Len(t, sends, 1)

	retrieved := sends[0]
	assert.Equal(t, send.Account, retrieved.Account)
	assert.Equal(t, send.ArticleID, retrieved.ArticleID)
	assert.Equal(t, "pending", retrieved.Status)
	assert.Equal(t, send.Title, retrieved.Title)
	assert.Equal(t, send.DestEmail, retrieved.DestEmail)
	assert.Equal(t, send.SenderEmail, retrieved.SenderEmail)
	assert.Equal(t, send.Provider, retrieved.Provider)
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateSendRecord() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	send := &model.Send{
		Account:     "test-account",
		ArticleID:   "article-update",
		SentAt:      now,
		Title:       "Test Article Update",
		DestEmail:   "dest@example.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repositories.CreateSendRecord(ctx, send)
	require.NoError(t, err)

	updateSend := &model.Send{
		Account:       "test-account",
		ArticleID:     "article-update",
		Status:        "sent",
		MessageID:     "message-id-123",
		ErrorResponse: "",
	}

	err = s.repositories.UpdateSendRecord(ctx, updateSend)
	require.NoError(t, err)

	sends, err := s.repositories.GetSendsByArticleID(ctx, send.ArticleID)
	require.NoError(t, err)
	require.Len(t, sends, 1)

	retrieved := sends[0]
	assert.Equal(t, "sent", retrieved.Status)
	assert.Equal(t, "message-id-123", retrieved.MessageID)
}

func (s *DynamoDBRepositoryTestSuite) TestGetSendsByNonExistentArticleID() {
	ctx := context.Background()
	t := s.T()

	sends, err := s.repositories.GetSendsByArticleID(ctx, "nonexistent-article")
	require.NoError(t, err)
	assert.Empty(t, sends)
}

func (s *DynamoDBRepositoryTestSuite) TestGetSendsByAccountDateRange() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	account := "test-account-sends"

	for i := range 3 {
		sendTime := now.Add(-time.Duration(i) * 24 * time.Hour)
		send := &model.Send{
			Account:     account,
			ArticleID:   "article-" + string(rune('0'+i)),
			SentAt:      sendTime,
			Title:       "Article " + string(rune('0'+i)),
			DestEmail:   "dest" + string(rune('0'+i)) + "@example.com",
			SenderEmail: "sender@example.com",
			Provider:    "mailjet",
		}

		err := s.repositories.CreateSendRecord(ctx, send)
		require.NoError(t, err)
	}

	startDate := now.Add(-2 * 24 * time.Hour).Add(-time.Second)
	endDate := now.Add(time.Second)

	sends, err := s.repositories.GetSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Len(t, sends, 3)
}

func (s *DynamoDBRepositoryTestSuite) TestCountSendsByAccountDateRange() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	account := "test-account-count"

	for i := range 5 {
		sendTime := now.Add(-time.Duration(i) * 24 * time.Hour)
		send := &model.Send{
			Account:     account,
			ArticleID:   "article-count-" + string(rune('0'+i)),
			SentAt:      sendTime,
			Title:       "Count Article " + string(rune('0'+i)),
			DestEmail:   "count" + string(rune('0'+i)) + "@example.com",
			SenderEmail: "sender@example.com",
			Provider:    "mailjet",
		}

		err := s.repositories.CreateSendRecord(ctx, send)
		require.NoError(t, err)
	}

	startDate := now.Add(-4 * 24 * time.Hour)
	endDate := now

	count, err := s.repositories.CountSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 4, count)
}

func (s *DynamoDBRepositoryTestSuite) TestCountSendsByAccountDateRangePartial() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	account := "test-account-count-partial"

	for i := range 10 {
		sendTime := now.Add(-time.Duration(i) * 24 * time.Hour)
		send := &model.Send{
			Account:     account,
			ArticleID:   "article-count-p-" + string(rune('0'+i)),
			SentAt:      sendTime,
			Title:       "Count Partial " + string(rune('0'+i)),
			DestEmail:   "partial" + string(rune('0'+i)) + "@example.com",
			SenderEmail: "sender@example.com",
			Provider:    "mailjet",
		}

		err := s.repositories.CreateSendRecord(ctx, send)
		require.NoError(t, err)
	}

	startDate := now.Add(-5 * 24 * time.Hour).Add(-time.Second)
	endDate := now.Add(-1 * 24 * time.Hour).Add(time.Second)

	count, err := s.repositories.CountSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}
