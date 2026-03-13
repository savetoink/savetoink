// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/suite"
)

type SQLiteRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	repository *SQLite
}

func TestSQLiteRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SQLiteRepositoryTestSuite))
}

func (s *SQLiteRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	repo, err := NewSQLite(s.ctx, ":memory:")
	if err != nil {
		s.T().Fatalf("failed to create sqlite repository: %v", err)
	}

	s.repository = repo
}

func (s *SQLiteRepositoryTestSuite) TearDownSuite() {
	if err := s.repository.Close(); err != nil {
		s.T().Logf("failed to close repository: %v", err)
	}
}

func (s *SQLiteRepositoryTestSuite) SetupTest() {
	s.cleanupTables()
}

func (s *SQLiteRepositoryTestSuite) cleanupTables() {
	_, _ = s.repository.db.ExecContext(s.ctx, "DELETE FROM sends")
	_, _ = s.repository.db.ExecContext(s.ctx, "DELETE FROM user_profiles")
	_, _ = s.repository.db.ExecContext(s.ctx, "DELETE FROM articles")
}

func (s *SQLiteRepositoryTestSuite) TestCreateTables() {
	s.NotNil(s.repository)
	s.NotNil(s.repository.db)
}

func (s *SQLiteRepositoryTestSuite) TestSchemaCreatedOnInitialization() {
	var tableName string
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"

	err := s.repository.db.QueryRowContext(s.ctx, query, "articles").Scan(&tableName)
	s.NoError(err)
	s.Equal("articles", tableName)

	err = s.repository.db.QueryRowContext(s.ctx, query, "user_profiles").Scan(&tableName)
	s.NoError(err)
	s.Equal("user_profiles", tableName)

	err = s.repository.db.QueryRowContext(s.ctx, query, "sends").Scan(&tableName)
	s.NoError(err)
	s.Equal("sends", tableName)
}

func (s *SQLiteRepositoryTestSuite) TestNewSQLite_EmptyPath() {
	_, err := NewSQLite(s.ctx, "")
	s.Error(err)
	s.Contains(err.Error(), "database path is required")
}

func (s *SQLiteRepositoryTestSuite) TestNewSQLite_InvalidPath() {
	_, err := NewSQLite(s.ctx, "/invalid/path/that/does/not/exist.db")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestStoreArticle() {
	publishedAt := time.Now().UTC()
	article := &model.Article{
		Account:            "test-account",
		ID:                 "article-1",
		URL:                "https://example.com/article",
		CreatedAt:          time.Now().UTC(),
		Title:              "Test Article",
		Content:            "Test Content",
		Author:             "Test Author",
		SiteName:           "Test Site",
		SourceDomain:       "example.com",
		Excerpt:            "Test excerpt",
		ImageURL:           "https://example.com/image.jpg",
		ContentType:        "text/html",
		Language:           "en",
		WordCount:          100,
		ReadingTimeMinutes: 1,
		PublishedAt:        &publishedAt,
		Favorite:           true,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", "article-1")
	s.NoError(err)
	s.Equal("test-account", retrieved.Account)
	s.Equal("article-1", retrieved.ID)
	s.Equal("https://example.com/article", retrieved.URL)
	s.Equal("Test Article", retrieved.Title)
	s.Equal("Test Content", retrieved.Content)
	s.Equal("Test Author", retrieved.Author)
	s.Equal("Test Site", retrieved.SiteName)
	s.Equal("example.com", retrieved.SourceDomain)
	s.Equal("Test excerpt", retrieved.Excerpt)
	s.Equal("https://example.com/image.jpg", retrieved.ImageURL)
	s.Equal("text/html", retrieved.ContentType)
	s.Equal("en", retrieved.Language)
	s.Equal(100, retrieved.WordCount)
	s.Equal(1, retrieved.ReadingTimeMinutes)
	s.True(retrieved.Favorite)
	s.NotNil(retrieved.PublishedAt)
}

func (s *SQLiteRepositoryTestSuite) TestStoreArticle_WithoutAccount() {
	article := &model.Article{
		ID:        "article-1",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(s.ctx, article)
	s.Error(err)
	s.Contains(err.Error(), "account field is required")
}

func (s *SQLiteRepositoryTestSuite) TestStoreArticle_UpdateExisting() {
	now := time.Now().UTC()
	article := &model.Article{
		Account:   "test-account",
		ID:        "article-1",
		URL:       "https://example.com/article",
		CreatedAt: now,
		Title:     "Original Title",
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	article.Title = "Updated Title"
	article.Content = "Updated Content"
	err = s.repository.Store(s.ctx, article)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", "article-1")
	s.NoError(err)
	s.Equal("Updated Title", retrieved.Title)
	s.Equal("Updated Content", retrieved.Content)
}

func (s *SQLiteRepositoryTestSuite) TestGetByAccountAndID_NotFound() {
	_, err := s.repository.GetByAccountAndID(s.ctx, "nonexistent", "nonexistent")
	s.Error(err)
	s.Equal(ErrNotFound, err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccountAndID() {
	article := &model.Article{
		Account:   "test-account",
		ID:        "article-1",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.DeleteByAccountAndID(s.ctx, "test-account", "article-1")
	s.NoError(err)

	_, err = s.repository.GetByAccountAndID(s.ctx, "test-account", "article-1")
	s.Error(err)
	s.Equal(ErrNotFound, err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccountAndID_NotFound() {
	err := s.repository.DeleteByAccountAndID(s.ctx, "nonexistent", "nonexistent")
	s.Error(err)
	s.Equal(ErrNotFound, err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite() {
	article := &model.Article{
		Account:   "test-account",
		ID:        "article-1",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
		Favorite:  false,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.UpdateFavorite(s.ctx, "test-account", "article-1", true)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", "article-1")
	s.NoError(err)
	s.True(retrieved.Favorite)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite_NotFound() {
	err := s.repository.UpdateFavorite(s.ctx, "nonexistent", "nonexistent", true)
	s.Error(err)
	s.Equal(ErrNotFound, err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccount() {
	for i := range 3 {
		article := &model.Article{
			Account:   "test-account",
			ID:        "article-" + string(rune('1'+i)),
			URL:       "https://example.com/article",
			CreatedAt: time.Now().UTC(),
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	count, err := s.repository.DeleteByAccount(s.ctx, "test-account")
	s.NoError(err)
	s.Equal(3, count)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccount_NotFound() {
	count, err := s.repository.DeleteByAccount(s.ctx, "nonexistent")
	s.NoError(err)
	s.Equal(0, count)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount() {
	now := time.Now().UTC()
	for i := range 5 {
		article := &model.Article{
			Account:   "test-account",
			ID:        "article-" + string(rune('a'+i)),
			URL:       "https://example.com/article",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  i%2 == 0,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	articles, lastKey, total, err := s.repository.GetMetadataByAccount(s.ctx, "test-account", 1, 2, nil)
	s.NoError(err)
	s.Len(articles, 2)
	s.Nil(lastKey)
	s.Equal(5, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_WithFavoriteFilter() {
	now := time.Now().UTC()
	for i := range 5 {
		article := &model.Article{
			Account:   "test-account",
			ID:        "article-" + string(rune('a'+i)),
			URL:       "https://example.com/article",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  i%2 == 0,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	favorite := true
	articles, _, total, err := s.repository.GetMetadataByAccount(s.ctx, "test-account", 1, 10, &favorite)
	s.NoError(err)
	s.Len(articles, 3)
	s.Equal(3, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_EmptyResult() {
	articles, _, total, err := s.repository.GetMetadataByAccount(s.ctx, "nonexistent", 1, 10, nil)
	s.NoError(err)
	s.Empty(articles)
	s.Equal(0, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_PageOutOfBounds() {
	now := time.Now().UTC()
	article := &model.Article{
		Account:   "pageout-account",
		ID:        "pageout-article-1",
		URL:       "https://example.com/article",
		CreatedAt: now,
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	articles, _, total, err := s.repository.GetMetadataByAccount(s.ctx, "pageout-account", 100, 10, nil)
	s.NoError(err)
	s.Empty(articles)
	s.Equal(1, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_InvalidPageSize() {
	_, _, total, err := s.repository.GetMetadataByAccount(s.ctx, "invalidpagesize-account", 1, 150, nil)
	s.NoError(err)
	s.Equal(0, total)
}

func (s *SQLiteRepositoryTestSuite) TestCreateSendRecord() {
	send := &model.Send{
		Account:     "test-account",
		ArticleID:   "create-send-article",
		Title:       "Test Article",
		DestEmail:   "device@kindle.com",
		SenderEmail: "sender@example.com",
		MessageID:   "msg-123",
		Provider:    "mailjet",
	}

	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	sends, err := s.repository.GetSendsByArticleID(s.ctx, "create-send-article")
	s.NoError(err)
	s.Len(sends, 1)
	s.Equal("test-account", sends[0].Account)
	s.Equal("create-send-article", sends[0].ArticleID)
	s.Equal("Test Article", sends[0].Title)
	s.Equal("device@kindle.com", sends[0].DestEmail)
	s.Equal("sender@example.com", sends[0].SenderEmail)
	s.Equal("msg-123", sends[0].MessageID)
	s.Equal("mailjet", sends[0].Provider)
	s.Equal("pending", sends[0].Status)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord() {
	send := &model.Send{
		Account:     "test-account",
		ArticleID:   "update-send-article",
		Title:       "Test Article",
		DestEmail:   "device@kindle.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	send.Status = "sent"
	send.MessageID = "msg-123"
	err = s.repository.UpdateSendRecord(s.ctx, send)
	s.NoError(err)

	sends, err := s.repository.GetSendsByArticleID(s.ctx, "update-send-article")
	s.NoError(err)
	s.Len(sends, 1)
	s.Equal("sent", sends[0].Status)
	s.Equal("msg-123", sends[0].MessageID)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_NotFound() {
	send := &model.Send{
		Account:   "test-account",
		ArticleID: "notfound-article",
		Status:    "sent",
	}

	err := s.repository.UpdateSendRecord(s.ctx, send)
	s.Error(err)
	s.Contains(err.Error(), "send record not found")
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_WithError() {
	send := &model.Send{
		Account:     "test-account",
		ArticleID:   "error-send-article",
		Title:       "Test Article",
		DestEmail:   "device@kindle.com",
		SenderEmail: "sender@example.com",
		Provider:    "mailjet",
	}

	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	send.ErrorResponse = "bounce error"
	send.Status = "failed"
	err = s.repository.UpdateSendRecord(s.ctx, send)
	s.NoError(err)

	sends, err := s.repository.GetSendsByArticleID(s.ctx, "error-send-article")
	s.NoError(err)
	s.Equal("failed", sends[0].Status)
	s.Equal("bounce error", sends[0].ErrorResponse)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByArticleID_Empty() {
	sends, err := s.repository.GetSendsByArticleID(s.ctx, "empty-article-id")
	s.NoError(err)
	s.Empty(sends)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByAccountDateRange() {
	now := time.Now().UTC()
	for i := range 3 {
		articleID := "daterange-article-" + string(rune('a'+i))
		send := &model.Send{
			Account:     "test-account",
			ArticleID:   articleID,
			Title:       "Test Article",
			SenderEmail: "sender@example.com",
			SentAt:      now.Add(-time.Duration(i) * time.Hour),
		}
		err := s.repository.CreateSendRecord(s.ctx, send)
		s.NoError(err)
	}

	startDate := now.Add(-3 * time.Hour)
	endDate := now

	sends, err := s.repository.GetSendsByAccountDateRange(s.ctx, "test-account", startDate, endDate)
	s.NoError(err)
	s.Len(sends, 3)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByAccountDateRange_Empty() {
	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour)
	endDate := now

	sends, err := s.repository.GetSendsByAccountDateRange(s.ctx, "empty-account", startDate, endDate)
	s.NoError(err)
	s.Empty(sends)
}

func (s *SQLiteRepositoryTestSuite) TestCountSendsByAccountDateRange() {
	now := time.Now().UTC()
	for i := range 3 {
		articleID := "count-article-" + string(rune('a'+i))
		send := &model.Send{
			Account:     "test-account",
			ArticleID:   articleID,
			Title:       "Test Article",
			SenderEmail: "sender@example.com",
			SentAt:      now.Add(-time.Duration(i) * time.Hour),
		}
		err := s.repository.CreateSendRecord(s.ctx, send)
		s.NoError(err)
	}

	startDate := now.Add(-3 * time.Hour)
	endDate := now

	count, err := s.repository.CountSendsByAccountDateRange(s.ctx, "test-account", startDate, endDate)
	s.NoError(err)
	s.Equal(3, count)
}

func (s *SQLiteRepositoryTestSuite) TestCountSendsByAccountDateRange_Empty() {
	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour)
	endDate := now

	count, err := s.repository.CountSendsByAccountDateRange(s.ctx, "empty-account", startDate, endDate)
	s.NoError(err)
	s.Equal(0, count)
}

func (s *SQLiteRepositoryTestSuite) TestGetUserProfile() {
	profile := &model.UserProfile{
		Account:     "test-account",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
		AutoSend:    true,
		BouncedEmails: map[string]model.BounceInfo{
			"bounce@example.com": {Timestamp: time.Now(), Error: "bounce"},
		},
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, "test-account")
	s.NoError(err)
	s.Equal("test-account", retrieved.Account)
	s.Equal("user@example.com", retrieved.Email)
	s.Equal("device@kindle.com", retrieved.DeviceEmail)
	s.True(retrieved.AutoSend)
	s.Len(retrieved.BouncedEmails, 1)
}

func (s *SQLiteRepositoryTestSuite) TestGetUserProfile_NotFound() {
	profile, err := s.repository.GetUserProfile(s.ctx, "nonexistent")
	s.NoError(err)
	s.Nil(profile)
}

func (s *SQLiteRepositoryTestSuite) TestGetAccountIDByDeviceEmail() {
	profile := &model.UserProfile{
		Account:     "test-account",
		DeviceEmail: "device@kindle.com",
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	account, err := s.repository.GetAccountIDByDeviceEmail(s.ctx, "device@kindle.com")
	s.NoError(err)
	s.Equal("test-account", account)
}

func (s *SQLiteRepositoryTestSuite) TestGetAccountIDByDeviceEmail_NotFound() {
	account, err := s.repository.GetAccountIDByDeviceEmail(s.ctx, "nonexistent@kindle.com")
	s.NoError(err)
	s.Empty(account)
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_WithoutAccount() {
	profile := &model.UserProfile{
		Email: "user@example.com",
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.Error(err)
	s.Contains(err.Error(), "account field is required")
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_UpdateExisting() {
	profile := &model.UserProfile{
		Account:  "test-account",
		Email:    "user@example.com",
		AutoSend: false,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	profile.Email = "newuser@example.com"
	profile.AutoSend = true
	err = s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, "test-account")
	s.NoError(err)
	s.Equal("newuser@example.com", retrieved.Email)
	s.True(retrieved.AutoSend)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserProfile() {
	profile := &model.UserProfile{
		Account: "test-account",
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserProfile(s.ctx, "test-account")
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, "test-account")
	s.NoError(err)
	s.Nil(retrieved)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserProfile_NotFound() {
	err := s.repository.DeleteUserProfile(s.ctx, "nonexistent")
	s.Error(err)
	s.Equal(ErrNotFound, err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserDeviceEmail() {
	profile := &model.UserProfile{
		Account:     "test-account",
		DeviceEmail: "device@kindle.com",
		AutoSend:    true,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserDeviceEmail(s.ctx, "test-account")
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, "test-account")
	s.NoError(err)
	s.Empty(retrieved.DeviceEmail)
	s.False(retrieved.AutoSend)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserDeviceEmail_NotFound() {
	err := s.repository.DeleteUserDeviceEmail(s.ctx, "nonexistent")
	s.Error(err)
	s.Equal(ErrNotFound, err)
}
