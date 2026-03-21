// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/suite"
)

const (
	testArticleID = "article-1"
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
		ID:                 testArticleID,
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

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", testArticleID)
	s.NoError(err)
	s.Equal("test-account", retrieved.Account)
	s.Equal(testArticleID, retrieved.ID)
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
		ID:        testArticleID,
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
		ID:        testArticleID,
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

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", testArticleID)
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
		ID:        testArticleID,
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.DeleteByAccountAndID(s.ctx, "test-account", testArticleID)
	s.NoError(err)

	_, err = s.repository.GetByAccountAndID(s.ctx, "test-account", testArticleID)
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
		ID:        testArticleID,
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
		Favorite:  false,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.UpdateFavorite(s.ctx, "test-account", testArticleID, true)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", testArticleID)
	s.NoError(err)
	s.True(retrieved.Favorite)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite_NotFound() {
	err := s.repository.UpdateFavorite(s.ctx, "nonexistent", "nonexistent", true)
	s.Error(err)
	s.Equal(ErrNotFound, err)
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

func (s *SQLiteRepositoryTestSuite) TestToArticle_InvalidCreatedAt() {
	const query = `INSERT INTO articles (account, id, url, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.repository.db.ExecContext(
		s.ctx,
		query,
		"test-account",
		"error-article-1",
		"https://example.com",
		"invalid-timestamp",
	)
	s.NoError(err)

	const selectQuery = `SELECT account, id, url, created_at, title, content, author, site_name,
		source_domain, excerpt, image_url, content_type, language, error,
		word_count, reading_time_minutes, published_at, favorite
		FROM articles WHERE account = ? AND id = ?`
	row := s.repository.db.QueryRowContext(s.ctx, selectQuery, "test-account", "error-article-1")

	var a articleRow
	err = row.Scan(
		&a.Account, &a.ID, &a.URL, &a.CreatedAt,
		&a.Title, &a.Content, &a.Author, &a.SiteName,
		&a.SourceDomain, &a.Excerpt, &a.ImageURL, &a.ContentType,
		&a.Language, &a.Error, &a.WordCount, &a.ReadingTimeMinutes,
		&a.PublishedAt, &a.Favorite,
	)
	s.NoError(err)

	_, err = a.toArticle()
	s.Error(err)
	s.Contains(err.Error(), "failed to parse created_at")
}

func (s *SQLiteRepositoryTestSuite) TestToArticle_InvalidPublishedAt() {
	now := time.Now().UTC()
	article := &model.Article{
		Account:   "test-account",
		ID:        "error-article-2",
		URL:       "https://example.com/article",
		CreatedAt: now,
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	query := `UPDATE articles SET published_at = 'invalid-date' WHERE account = ? AND id = ?`
	_, err = s.repository.db.ExecContext(s.ctx, query, "test-account", "error-article-2")
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, "test-account", "error-article-2")
	s.Error(err)
	s.Contains(err.Error(), "failed to parse published_at")
	s.Nil(retrieved)
}

func (s *SQLiteRepositoryTestSuite) TestToUserProfile_InvalidBouncedEmailsJSON() {
	const query = `INSERT INTO user_profiles (account, bounced_emails) VALUES (?, ?)`
	_, err := s.repository.db.ExecContext(s.ctx, query, "test-account", "invalid-json{")
	s.NoError(err)

	const selectQuery = `SELECT account, email, device_email, auto_send, bounced_emails
		FROM user_profiles WHERE account = ?`
	row := s.repository.db.QueryRowContext(s.ctx, selectQuery, "test-account")

	var p profileRow
	err = row.Scan(&p.Account, &p.Email, &p.DeviceEmail, &p.AutoSend, &p.BouncedEmails)
	s.NoError(err)

	_, err = p.toUserProfile()
	s.Error(err)
	s.Contains(err.Error(), "failed to unmarshal bounced emails")
}

func (s *SQLiteRepositoryTestSuite) TestProfileToRow_MarshalError() {
	profile := &model.UserProfile{
		Account: "test-account",
		BouncedEmails: map[string]model.BounceInfo{
			"test@example.com": {},
		},
	}

	_, err := profileToRow(profile)
	s.NoError(err)
}

func (s *SQLiteRepositoryTestSuite) TestToSend_InvalidSentAt() {
	const query = `INSERT INTO sends (account, article_id, sent_at) VALUES (?, ?, ?)`
	_, err := s.repository.db.ExecContext(
		s.ctx,
		query,
		"test-account",
		"error-send-article",
		"invalid-timestamp",
	)
	s.NoError(err)

	const selectQuery = `SELECT account, article_id, sent_at, title, dest_email, status,
		sender_email, message_id, provider, error_response FROM sends WHERE article_id = ?`
	rows, err := s.repository.db.QueryContext(s.ctx, selectQuery, "error-send-article")
	s.NoError(err)
	defer func() { _ = rows.Close() }()

	var sr sendRow
	if rows.Next() {
		err = rows.Scan(
			&sr.Account, &sr.ArticleID, &sr.SentAt,
			&sr.Title, &sr.DestEmail, &sr.Status,
			&sr.SenderEmail, &sr.MessageID, &sr.Provider, &sr.ErrorResponse,
		)
		s.NoError(err)

		_, err = sr.toSend()
		s.Error(err)
		s.Contains(err.Error(), "failed to parse sent_at")
	}
}

func (s *SQLiteRepositoryTestSuite) TestGetByAccountAndID_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   "test-account",
		ID:        "ctx-cancel-article",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	_, err = s.repository.GetByAccountAndID(ctx, "test-account", "ctx-cancel-article")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccountAndID_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   "test-account",
		ID:        "ctx-del-article",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.DeleteByAccountAndID(ctx, "test-account", "ctx-del-article")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   "test-account",
		ID:        "ctx-fav-article",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.UpdateFavorite(ctx, "test-account", "ctx-fav-article", true)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   "test-account",
		ID:        "ctx-meta-article",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	articles, lastKey, total, err := s.repository.GetMetadataByAccount(ctx, "test-account", 1, 10, nil)
	_ = articles
	_ = lastKey
	_ = total
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: "test-account",
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	_, err = s.repository.GetUserProfile(ctx, "test-account")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetAccountIDByDeviceEmail_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account:     "test-account",
		DeviceEmail: "device@kindle.com",
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	_, err = s.repository.GetAccountIDByDeviceEmail(ctx, "device@kindle.com")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: "test-account",
	}

	err := s.repository.PutUserProfile(ctx, profile)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: "test-account",
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserProfile(ctx, "test-account")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserDeviceEmail_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account:     "test-account",
		DeviceEmail: "device@kindle.com",
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserDeviceEmail(ctx, "test-account")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCreateSendRecord_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   "test-account",
		ArticleID: "ctx-create-send",
	}

	err := s.repository.CreateSendRecord(ctx, send)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   "test-account",
		ArticleID: "ctx-update-send",
	}

	err := s.repository.UpdateSendRecord(ctx, send)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByArticleID_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   "test-account",
		ArticleID: "ctx-get-send",
	}
	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	_, err = s.repository.GetSendsByArticleID(ctx, "ctx-get-send")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByAccountDateRange_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	_, err := s.repository.GetSendsByAccountDateRange(ctx, "test-account", time.Now().Add(-time.Hour), time.Now())
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCountSendsByAccountDateRange_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	_, err := s.repository.CountSendsByAccountDateRange(ctx, "test-account", time.Now().Add(-time.Hour), time.Now())
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestQueryArticlesByAccount_ScanError() {
	const insertQuery = `INSERT INTO articles (account, id, url, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.repository.db.ExecContext(
		s.ctx,
		insertQuery,
		"test-account",
		"scan-error-article",
		"https://example.com/article",
		time.Now().UTC().Format(time.RFC3339),
	)
	s.NoError(err)

	const updateQuery = `UPDATE articles SET created_at = 'invalid-timestamp'
		WHERE account = ? AND id = ?`
	_, err = s.repository.db.ExecContext(s.ctx, updateQuery, "test-account", "scan-error-article")
	s.NoError(err)

	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	articles, lastKey, total, err := s.repository.GetMetadataByAccount(ctx, "test-account", 1, 10, nil)
	_ = articles
	_ = lastKey
	_ = total
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestQuerySends_ScanError() {
	send := &model.Send{
		Account:   "test-account",
		ArticleID: "scan-error-send",
		Title:     "Test Article",
	}
	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	query := `UPDATE sends SET sent_at = 'invalid-timestamp' WHERE article_id = ?`
	_, err = s.repository.db.ExecContext(s.ctx, query, "scan-error-send")
	s.NoError(err)

	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	_, err = s.repository.GetSendsByArticleID(ctx, "scan-error-send")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestNewSQLite_PingError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewSQLite(ctx, ":memory:")
	s.Error(err)
	s.Regexp("failed to ping .+ database", err.Error())
}

func (s *SQLiteRepositoryTestSuite) TestCreateTables_Error() {
	repo, err := NewSQLite(s.ctx, "file:/tmp/test-create-tables.db?mode=memory")
	s.NoError(err)

	_ = repo.db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.CreateTables(ctx)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCreateArticleTable_Error() {
	repo, err := NewSQLite(s.ctx, "file:/tmp/test-create-article.db?mode=memory")
	s.NoError(err)

	_ = repo.db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.createArticleTable(ctx)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCreateUserProfileTable_Error() {
	repo, err := NewSQLite(s.ctx, "file:/tmp/test-create-profile.db?mode=memory")
	s.NoError(err)

	_ = repo.db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.createUserProfileTable(ctx)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCreateSendsTable_Error() {
	repo, err := NewSQLite(s.ctx, "file:/tmp/test-create-sends.db?mode=memory")
	s.NoError(err)

	_ = repo.db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.createSendsTable(ctx)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestStore_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   "test-account",
		ID:        "store-error-article",
		URL:       "https://example.com/article",
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(ctx, article)
	s.Error(err)
}
