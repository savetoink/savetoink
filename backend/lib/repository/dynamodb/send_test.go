package repository

import (
	"context"
	"fmt"
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

func (s *DynamoDBRepositoryTestSuite) TestUpdateSendRecordNotFound() {
	ctx := context.Background()
	t := s.T()

	updateSend := &model.Send{
		Account:       "test-account",
		ArticleID:     "nonexistent-article",
		Status:        "sent",
		MessageID:     "message-id-123",
		ErrorResponse: "",
	}

	err := s.repositories.UpdateSendRecord(ctx, updateSend)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send record not found")
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateSendRecordAccountMismatch() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	send := &model.Send{
		Account:     "test-account-original",
		ArticleID:   "article-mismatch",
		SentAt:      now,
		Title:       "Test Article",
		DestEmail:   "dest@example.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repositories.CreateSendRecord(ctx, send)
	require.NoError(t, err)

	updateSend := &model.Send{
		Account:   "different-account",
		ArticleID: "article-mismatch",
		Status:    "sent",
		MessageID: "message-id-123",
	}

	err = s.repositories.UpdateSendRecord(ctx, updateSend)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send record account mismatch")
}

func (s *DynamoDBRepositoryTestSuite) TestCreateSendRecordWithTimestamp() {
	ctx := context.Background()
	t := s.T()

	now := time.Now().Truncate(time.Second)
	send := &model.Send{
		Account:     "test-account-timestamp",
		ArticleID:   "article-timestamp",
		SentAt:      now,
		Title:       "Test Timestamp Article",
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
	assert.WithinDuration(t, now, retrieved.SentAt, time.Second)
}

func (s *DynamoDBRepositoryTestSuite) TestCreateSendRecordTimestampNotZero() {
	ctx := context.Background()
	t := s.T()

	now := time.Now().UTC().Truncate(time.Second)
	cutoff := now.Add(-5 * time.Minute)

	send := &model.Send{
		Account:     "test-account-notzero",
		ArticleID:   "article-notzero",
		SentAt:      now,
		Title:       "Test Not Zero Article",
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
	assert.True(t, retrieved.SentAt.After(cutoff), "SentAt should not be zero value")
	assert.WithinDuration(t, now, retrieved.SentAt, time.Second)
}

func (s *DynamoDBRepositoryTestSuite) TestGetSendsByAccountDateRangeRespectsTimestamp() {
	ctx := context.Background()
	t := s.T()

	now := time.Now().UTC().Truncate(time.Second)
	account := "test-account-daterange-ts"

	sends := []struct {
		daysAgo int
		title   string
	}{
		{0, "Today Article"},
		{1, "Yesterday Article"},
		{2, "Two Days Ago Article"},
		{5, "Five Days Ago Article"},
		{10, "Ten Days Ago Article"},
	}

	for _, item := range sends {
		send := &model.Send{
			Account:     account,
			ArticleID:   "article-" + item.title,
			SentAt:      now.Add(-time.Duration(item.daysAgo) * 24 * time.Hour),
			Title:       item.title,
			DestEmail:   "dest@example.com",
			SenderEmail: "sender@example.com",
			Provider:    "mailjet",
		}

		err := s.repositories.CreateSendRecord(ctx, send)
		require.NoError(t, err)
	}

	startDate := now.Add(-6 * 24 * time.Hour)
	endDate := now.Add(time.Second)

	retrievedSends, err := s.repositories.GetSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)

	assert.Len(t, retrievedSends, 4, "should return sends from 0, 1, 2, and 5 days ago")

	titles := make([]string, len(retrievedSends))
	for i, send := range retrievedSends {
		titles[i] = send.Title
	}

	expectedTitles := []string{"Five Days Ago Article", "Two Days Ago Article", "Yesterday Article", "Today Article"}
	for _, expected := range expectedTitles {
		assert.Contains(t, titles, expected, "should include expected send")
	}
}

func (s *DynamoDBRepositoryTestSuite) TestCountSendsByAccountDateRangeRespectsTimestamp() {
	ctx := context.Background()
	t := s.T()

	now := time.Now().UTC().Truncate(time.Second)
	account := "test-account-count-ts"

	for i := range 10 {
		send := &model.Send{
			Account:     account,
			ArticleID:   fmt.Sprintf("article-count-ts-%d", i),
			SentAt:      now.Add(-time.Duration(i) * 24 * time.Hour),
			Title:       fmt.Sprintf("Count TS Article %d", i),
			DestEmail:   "countts@example.com",
			SenderEmail: "sender@example.com",
			Provider:    "mailjet",
		}

		err := s.repositories.CreateSendRecord(ctx, send)
		require.NoError(t, err)
	}

	startDate := now.Add(-5 * 24 * time.Hour)
	endDate := now.Add(time.Second)

	count, err := s.repositories.CountSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 6, count, "should count sends from 0, 1, 2, 3, 4, and 5 days ago")
}

func (s *DynamoDBRepositoryTestSuite) TestGetSendsByAccountDateRangeEmptyResult() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	account := "test-account-empty-range"

	startDate := now.Add(-10 * 24 * time.Hour)
	endDate := now.Add(-5 * 24 * time.Hour)

	sends, err := s.repositories.GetSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Empty(t, sends)
}

func (s *DynamoDBRepositoryTestSuite) TestCountSendsByAccountDateRangeEmptyResult() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	account := "test-account-empty-count"

	startDate := now.Add(-10 * 24 * time.Hour)
	endDate := now.Add(-5 * 24 * time.Hour)

	count, err := s.repositories.CountSendsByAccountDateRange(ctx, account, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateSendRecordWithErrorResponse() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	send := &model.Send{
		Account:     "test-account-error",
		ArticleID:   "article-error",
		SentAt:      now,
		Title:       "Test Article Error",
		DestEmail:   "dest@example.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repositories.CreateSendRecord(ctx, send)
	require.NoError(t, err)

	updateSend := &model.Send{
		Account:       "test-account-error",
		ArticleID:     "article-error",
		Status:        "failed",
		MessageID:     "",
		ErrorResponse: "SMTP error: 550 5.7.1",
	}

	err = s.repositories.UpdateSendRecord(ctx, updateSend)
	require.NoError(t, err)

	sends, err := s.repositories.GetSendsByArticleID(ctx, send.ArticleID)
	require.NoError(t, err)
	require.Len(t, sends, 1)

	retrieved := sends[0]
	assert.Equal(t, "failed", retrieved.Status)
	assert.Equal(t, "SMTP error: 550 5.7.1", retrieved.ErrorResponse)
}
