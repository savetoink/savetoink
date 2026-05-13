// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"testing"
	"time"

	internaltypes "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/suite"
)

const (
	testArticleID   = "article-1"
	testAccount     = "test-account"
	testURL         = "https://example.com/article"
	testTitle       = "Test Article"
	testDomain      = "example.com"
	testImageURL    = "https://example.com/image.jpg"
	testDestEmail   = "device@kindle.com"
	testSenderEmail = "sender@example.com"
	testProvider    = "mailjet"
	testUserEmail   = "user@example.com"
	testAuthor      = "Author"
	testExcerpt     = "Excerpt"
	testSiteName    = "Example Site"
	tagTech         = "tech"
	tagProgramming  = "programming"
	tagGolang       = "golang"
	tagDatabase     = "database"
	tagArticle1     = "tag-article-1"
	articleTag2     = "article-tag-2"
	testURL1        = "https://example.com/article1"
	testURL2        = "https://example.com/article2"
	testLang        = "en"
	testMsgID       = "msg-123"
	testStatus      = "sent"
	testArticleTag1 = "article-tag-1"
	testArticleTag3 = "article-tag-3"
	testArticleTag4 = "article-tag-4"
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
	_, _ = s.repository.db.ExecContext(s.ctx, "DELETE FROM article_tags")
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

	err = s.repository.db.QueryRowContext(s.ctx, query, "article_tags").Scan(&tableName)
	s.NoError(err)
	s.Equal("article_tags", tableName)
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
		Account:            testAccount,
		ID:                 testArticleID,
		URL:                testURL,
		CreatedAt:          time.Now().UTC(),
		Title:              testTitle,
		Content:            "Test Content",
		Author:             "Test Author",
		SiteName:           "Test Site",
		SourceDomain:       testDomain,
		Excerpt:            "Test excerpt",
		ImageURL:           testImageURL,
		ContentType:        "text/html",
		Language:           testLang,
		WordCount:          100,
		ReadingTimeMinutes: 1,
		PublishedAt:        &publishedAt,
		Favorite:           true,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, testAccount, testArticleID)
	s.NoError(err)
	s.Equal(testAccount, retrieved.Account)
	s.Equal(testArticleID, retrieved.ID)
	s.Equal(testURL, retrieved.URL)
	s.Equal(testTitle, retrieved.Title)
	s.Equal("Test Content", retrieved.Content)
	s.Equal("Test Author", retrieved.Author)
	s.Equal("Test Site", retrieved.SiteName)
	s.Equal(testDomain, retrieved.SourceDomain)
	s.Equal("Test excerpt", retrieved.Excerpt)
	s.Equal(testImageURL, retrieved.ImageURL)
	s.Equal("text/html", retrieved.ContentType)
	s.Equal(testLang, retrieved.Language)
	s.Equal(100, retrieved.WordCount)
	s.Equal(1, retrieved.ReadingTimeMinutes)
	s.True(retrieved.Favorite)
	s.NotNil(retrieved.PublishedAt)
}

func (s *SQLiteRepositoryTestSuite) TestStoreArticle_WithoutAccount() {
	article := &model.Article{
		ID:        testArticleID,
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(s.ctx, article)
	s.Error(err)
	s.Contains(err.Error(), "account field is required")
}

func (s *SQLiteRepositoryTestSuite) TestStoreArticle_UpdateExisting() {
	now := time.Now().UTC()
	article := &model.Article{
		Account:   testAccount,
		ID:        testArticleID,
		URL:       testURL,
		CreatedAt: now,
		Title:     "Original Title",
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	article.Title = "Updated Title"
	article.Content = "Updated Content"
	err = s.repository.Store(s.ctx, article)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, testAccount, testArticleID)
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
		Account:   testAccount,
		ID:        testArticleID,
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.DeleteByAccountAndID(s.ctx, testAccount, testArticleID)
	s.NoError(err)

	_, err = s.repository.GetByAccountAndID(s.ctx, testAccount, testArticleID)
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
		Account:   testAccount,
		ID:        testArticleID,
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
		Favorite:  false,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.UpdateFavorite(s.ctx, testAccount, testArticleID, true)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, testAccount, testArticleID)
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
			Account:   testAccount,
			ID:        "article-" + string(rune('a'+i)),
			URL:       testURL,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  i%2 == 0,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	articles, total, err := s.repository.GetMetadataByAccount(s.ctx, testAccount, 1, 2, nil)
	s.NoError(err)
	s.Len(articles, 2)
	s.Equal(5, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_WithFavoriteFilter() {
	now := time.Now().UTC()
	for i := range 5 {
		article := &model.Article{
			Account:   testAccount,
			ID:        "article-" + string(rune('a'+i)),
			URL:       testURL,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  i%2 == 0,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	favorite := true
	articles, total, err := s.repository.GetMetadataByAccount(
		s.ctx, testAccount, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	s.NoError(err)
	s.Len(articles, 3)
	s.Equal(3, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_EmptyResult() {
	articles, total, err := s.repository.GetMetadataByAccount(s.ctx, "nonexistent", 1, 10, nil)
	s.NoError(err)
	s.Empty(articles)
	s.Equal(0, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_PageOutOfBounds() {
	now := time.Now().UTC()
	article := &model.Article{
		Account:   "pageout-account",
		ID:        "pageout-article-1",
		URL:       testURL,
		CreatedAt: now,
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	articles, total, err := s.repository.GetMetadataByAccount(s.ctx, "pageout-account", 100, 10, nil)
	s.NoError(err)
	s.Empty(articles)
	s.Equal(1, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_InvalidPageSize() {
	_, total, err := s.repository.GetMetadataByAccount(s.ctx, "invalidpagesize-account", 1, 150, nil)
	s.NoError(err)
	s.Equal(0, total)
}

func (s *SQLiteRepositoryTestSuite) TestCreateSendRecord() {
	send := &model.Send{
		Account:     testAccount,
		ArticleID:   "create-send-article",
		Title:       testTitle,
		DestEmail:   testDestEmail,
		SenderEmail: testSenderEmail,
		MessageID:   testMsgID,
		Provider:    testProvider,
	}

	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	sends, err := s.repository.GetSendsByArticleID(s.ctx, "create-send-article")
	s.NoError(err)
	s.Len(sends, 1)
	s.Equal(testAccount, sends[0].Account)
	s.Equal("create-send-article", sends[0].ArticleID)
	s.Equal(testTitle, sends[0].Title)
	s.Equal(testDestEmail, sends[0].DestEmail)
	s.Equal(testSenderEmail, sends[0].SenderEmail)
	s.Equal(testMsgID, sends[0].MessageID)
	s.Equal(testProvider, sends[0].Provider)
	s.Equal("pending", sends[0].Status)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord() {
	send := &model.Send{
		Account:     testAccount,
		ArticleID:   "update-send-article",
		Title:       testTitle,
		DestEmail:   testDestEmail,
		SenderEmail: testSenderEmail,
		Provider:    testProvider,
	}

	err := s.repository.CreateSendRecord(s.ctx, send)
	s.NoError(err)

	send.Status = testStatus
	send.MessageID = testMsgID
	err = s.repository.UpdateSendRecord(s.ctx, send)
	s.NoError(err)

	sends, err := s.repository.GetSendsByArticleID(s.ctx, "update-send-article")
	s.NoError(err)
	s.Len(sends, 1)
	s.Equal(testStatus, sends[0].Status)
	s.Equal(testMsgID, sends[0].MessageID)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_NotFound() {
	send := &model.Send{
		Account:   testAccount,
		ArticleID: "notfound-article",
		Status:    testStatus,
	}

	err := s.repository.UpdateSendRecord(s.ctx, send)
	s.Error(err)
	s.Contains(err.Error(), "send record not found")
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_WithError() {
	send := &model.Send{
		Account:     testAccount,
		ArticleID:   "error-send-article",
		Title:       testTitle,
		DestEmail:   testDestEmail,
		SenderEmail: testSenderEmail,
		Provider:    testProvider,
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
			Account:     testAccount,
			ArticleID:   articleID,
			Title:       testTitle,
			SenderEmail: testSenderEmail,
			SentAt:      now.Add(-time.Duration(i) * time.Hour),
		}
		err := s.repository.CreateSendRecord(s.ctx, send)
		s.NoError(err)
	}

	startDate := now.Add(-3 * time.Hour)
	endDate := now

	sends, err := s.repository.GetSendsByAccountDateRange(s.ctx, testAccount, startDate, endDate)
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
			Account:     testAccount,
			ArticleID:   articleID,
			Title:       testTitle,
			SenderEmail: testSenderEmail,
			SentAt:      now.Add(-time.Duration(i) * time.Hour),
		}
		err := s.repository.CreateSendRecord(s.ctx, send)
		s.NoError(err)
	}

	startDate := now.Add(-3 * time.Hour)
	endDate := now

	count, err := s.repository.CountSendsByAccountDateRange(s.ctx, testAccount, startDate, endDate)
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
		Account:     testAccount,
		Email:       testUserEmail,
		DeviceEmail: testDestEmail,
		AutoSend:    true,
		BouncedEmails: map[string]model.BounceInfo{
			"bounce@example.com": {Timestamp: time.Now(), Error: "bounce"},
		},
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, testAccount)
	s.NoError(err)
	s.Equal(testAccount, retrieved.Account)
	s.Equal(testUserEmail, retrieved.Email)
	s.Equal(testDestEmail, retrieved.DeviceEmail)
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
		Account:     testAccount,
		DeviceEmail: testDestEmail,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	account, err := s.repository.GetAccountIDByDeviceEmail(s.ctx, testDestEmail)
	s.NoError(err)
	s.Equal(testAccount, account)
}

func (s *SQLiteRepositoryTestSuite) TestGetAccountIDByDeviceEmail_NotFound() {
	account, err := s.repository.GetAccountIDByDeviceEmail(s.ctx, "nonexistent@kindle.com")
	s.NoError(err)
	s.Empty(account)
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_WithoutAccount() {
	profile := &model.UserProfile{
		Email: testUserEmail,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.Error(err)
	s.Contains(err.Error(), "account field is required")
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_UpdateExisting() {
	profile := &model.UserProfile{
		Account:  testAccount,
		Email:    testUserEmail,
		AutoSend: false,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	profile.Email = "newuser@example.com"
	profile.AutoSend = true
	err = s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, testAccount)
	s.NoError(err)
	s.Equal("newuser@example.com", retrieved.Email)
	s.True(retrieved.AutoSend)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserProfile() {
	profile := &model.UserProfile{
		Account: testAccount,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserProfile(s.ctx, testAccount)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, testAccount)
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
		Account:     testAccount,
		DeviceEmail: testDestEmail,
		AutoSend:    true,
	}

	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserDeviceEmail(s.ctx, testAccount)
	s.NoError(err)

	retrieved, err := s.repository.GetUserProfile(s.ctx, testAccount)
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
		testAccount,
		"error-article-1",
		"https://example.com",
		"invalid-timestamp",
	)
	s.NoError(err)

	const selectQuery = `SELECT account, id, url, created_at, title, content, author, site_name,
		source_domain, excerpt, image_url, content_type, language, error,
		word_count, reading_time_minutes, published_at, favorite
		FROM articles WHERE account = ? AND id = ?`
	row := s.repository.db.QueryRowContext(s.ctx, selectQuery, testAccount, "error-article-1")

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
		Account:   testAccount,
		ID:        "error-article-2",
		URL:       testURL,
		CreatedAt: now,
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	query := `UPDATE articles SET published_at = 'invalid-date' WHERE account = ? AND id = ?`
	_, err = s.repository.db.ExecContext(s.ctx, query, testAccount, "error-article-2")
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, testAccount, "error-article-2")
	s.Error(err)
	s.Contains(err.Error(), "failed to parse published_at")
	s.Nil(retrieved)
}

func (s *SQLiteRepositoryTestSuite) TestToUserProfile_InvalidBouncedEmailsJSON() {
	const query = `INSERT INTO user_profiles (account, bounced_emails) VALUES (?, ?)`
	_, err := s.repository.db.ExecContext(s.ctx, query, testAccount, "invalid-json{")
	s.NoError(err)

	const selectQuery = `SELECT account, email, device_email, auto_send, bounced_emails
		FROM user_profiles WHERE account = ?`
	row := s.repository.db.QueryRowContext(s.ctx, selectQuery, testAccount)

	var p profileRow
	err = row.Scan(&p.Account, &p.Email, &p.DeviceEmail, &p.AutoSend, &p.BouncedEmails)
	s.NoError(err)

	_, err = p.toUserProfile()
	s.Error(err)
	s.Contains(err.Error(), "failed to unmarshal bounced emails")
}

func (s *SQLiteRepositoryTestSuite) TestProfileToRow_MarshalError() {
	profile := &model.UserProfile{
		Account: testAccount,
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
		testAccount,
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
		Account:   testAccount,
		ID:        "ctx-cancel-article",
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	_, err = s.repository.GetByAccountAndID(ctx, testAccount, "ctx-cancel-article")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteByAccountAndID_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   testAccount,
		ID:        "ctx-del-article",
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.DeleteByAccountAndID(ctx, testAccount, "ctx-del-article")
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   testAccount,
		ID:        "ctx-fav-article",
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	err = s.repository.UpdateFavorite(ctx, testAccount, "ctx-fav-article", true)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	article := &model.Article{
		Account:   testAccount,
		ID:        "ctx-meta-article",
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}
	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	articles, total, err := s.repository.GetMetadataByAccount(ctx, testAccount, 1, 10, nil)
	_ = articles
	_ = total
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: testAccount,
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	_, err = s.repository.GetUserProfile(ctx, testAccount)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetAccountIDByDeviceEmail_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account:     testAccount,
		DeviceEmail: testDestEmail,
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	_, err = s.repository.GetAccountIDByDeviceEmail(ctx, testDestEmail)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestPutUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: testAccount,
	}

	err := s.repository.PutUserProfile(ctx, profile)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserProfile_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account: testAccount,
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserProfile(ctx, testAccount)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteUserDeviceEmail_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	profile := &model.UserProfile{
		Account:     testAccount,
		DeviceEmail: testDestEmail,
	}
	err := s.repository.PutUserProfile(s.ctx, profile)
	s.NoError(err)

	err = s.repository.DeleteUserDeviceEmail(ctx, testAccount)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCreateSendRecord_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   testAccount,
		ArticleID: "ctx-create-send",
	}

	err := s.repository.CreateSendRecord(ctx, send)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateSendRecord_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   testAccount,
		ArticleID: "ctx-update-send",
	}

	err := s.repository.UpdateSendRecord(ctx, send)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestGetSendsByArticleID_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	send := &model.Send{
		Account:   testAccount,
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

	_, err := s.repository.GetSendsByAccountDateRange(ctx, testAccount, time.Now().Add(-time.Hour), time.Now())
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestCountSendsByAccountDateRange_DatabaseError() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	_, err := s.repository.CountSendsByAccountDateRange(ctx, testAccount, time.Now().Add(-time.Hour), time.Now())
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestQueryArticlesByAccount_ScanError() {
	const insertQuery = `INSERT INTO articles (account, id, url, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.repository.db.ExecContext(
		s.ctx,
		insertQuery,
		testAccount,
		"scan-error-article",
		testURL,
		time.Now().UTC().Format(time.RFC3339),
	)
	s.NoError(err)

	const updateQuery = `UPDATE articles SET created_at = 'invalid-timestamp'
		WHERE account = ? AND id = ?`
	_, err = s.repository.db.ExecContext(s.ctx, updateQuery, testAccount, "scan-error-article")
	s.NoError(err)

	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	articles, total, err := s.repository.GetMetadataByAccount(ctx, testAccount, 1, 10, nil)
	_ = articles
	_ = total
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestQuerySends_ScanError() {
	send := &model.Send{
		Account:   testAccount,
		ArticleID: "scan-error-send",
		Title:     testTitle,
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
		Account:   testAccount,
		ID:        "store-error-article",
		URL:       testURL,
		CreatedAt: time.Now().UTC(),
	}

	err := s.repository.Store(ctx, article)
	s.Error(err)
}

func (s *SQLiteRepositoryTestSuite) TestUpdateFavorite_ToggleOnThenOff() {
	now := time.Now().UTC()

	article := &model.Article{
		Account:   testAccount,
		ID:        "toggle-fav-article",
		URL:       testURL,
		CreatedAt: now,
		Favorite:  false,
	}

	err := s.repository.Store(s.ctx, article)
	s.NoError(err)

	retrieved, err := s.repository.GetByAccountAndID(s.ctx, testAccount, "toggle-fav-article")
	s.NoError(err)
	s.False(retrieved.Favorite)

	err = s.repository.UpdateFavorite(s.ctx, testAccount, "toggle-fav-article", true)
	s.NoError(err)

	retrieved, err = s.repository.GetByAccountAndID(s.ctx, testAccount, "toggle-fav-article")
	s.NoError(err)
	s.True(retrieved.Favorite)

	err = s.repository.UpdateFavorite(s.ctx, testAccount, "toggle-fav-article", false)
	s.NoError(err)

	retrieved, err = s.repository.GetByAccountAndID(s.ctx, testAccount, "toggle-fav-article")
	s.NoError(err)
	s.False(retrieved.Favorite)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_FavoritesPagination() {
	now := time.Now().UTC()
	account := "test-account-fav-pag"

	for i := range 25 {
		article := &model.Article{
			Account:   account,
			ID:        "article-" + string(rune('a'+(i%26))),
			URL:       testURL,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  true,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	favorite := true

	articles, total, err := s.repository.GetMetadataByAccount(
		s.ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	s.NoError(err)
	s.Len(articles, 10)
	s.Equal(25, total)

	for _, article := range articles {
		s.True(article.Favorite)
	}

	articles, total, err = s.repository.GetMetadataByAccount(
		s.ctx, account, 2, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	s.NoError(err)
	s.Len(articles, 10)
	s.Equal(25, total)

	articles, total, err = s.repository.GetMetadataByAccount(
		s.ctx, account, 3, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	s.NoError(err)
	s.Len(articles, 5)
	s.Equal(25, total)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_NonFavoritesFilter() {
	now := time.Now().UTC()
	account := "test-account-non-fav"

	for i := range 10 {
		article := &model.Article{
			Account:   account,
			ID:        "article-" + string(rune('a'+i)),
			URL:       testURL,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
			Favorite:  i%2 == 0,
		}
		err := s.repository.Store(s.ctx, article)
		s.NoError(err)
	}

	favorite := false
	articles, total, err := s.repository.GetMetadataByAccount(
		s.ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	s.NoError(err)
	s.Len(articles, 5)
	s.Equal(5, total)

	for _, article := range articles {
		s.False(article.Favorite)
	}
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_WithTagFilter() {
	ctx := context.Background()

	const (
		tagTech        = tagTech
		tagProgramming = tagProgramming
		tagNonexistent = "nonexistent"
	)

	account := "test-account-tag"

	// Create articles with different tags
	testCases := []struct {
		id   string
		tags []string
		fav  bool
	}{
		{testArticleTag1, []string{tagTech}, false},
		{articleTag2, []string{tagTech, tagProgramming}, true},
		{testArticleTag3, []string{tagProgramming}, false},
		{testArticleTag4, []string{tagTech}, true},
		{"article-tag-5", []string{}, false},
	}

	for _, tc := range testCases {
		now := time.Now()
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 tc.id,
			URL:                "https://example.com/" + tc.id,
			Title:              "Article " + tc.id,
			Content:            "Content " + tc.id,
			CreatedAt:          now,
			Favorite:           tc.fav,
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           testLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testDomain,
		}

		err := s.repository.Store(ctx, article)
		s.NoError(err)

		// Add tags
		err = s.repository.AddTagsToArticle(ctx, account, tc.id, tc.tags, &article.CreatedAt)
		s.NoError(err)
	}

	// Test filtering by tagTech tag
	tag := tagTech
	articles, total, err := s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	s.NoError(err)
	s.Equal(3, total)
	s.Len(articles, 3)

	for _, article := range articles {
		s.Contains([]string{testArticleTag1, articleTag2, testArticleTag4}, article.ID)
	}

	// Test filtering by tagProgramming tag
	tag = tagProgramming
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	s.NoError(err)
	s.Equal(2, total)
	s.Len(articles, 2)

	for _, article := range articles {
		s.Contains([]string{articleTag2, testArticleTag3}, article.ID)
	}

	// Test filtering by non-existent tag
	tag = tagNonexistent
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	s.NoError(err)
	s.Equal(0, total)
	s.Empty(articles)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_WithCombinedFilters() {
	ctx := context.Background()

	const (
		tagTech        = tagTech
		tagProgramming = tagProgramming
	)

	account := "test-account-combined"

	// Create articles with different tag and favorite combinations
	testCases := []struct {
		id   string
		tags []string
		fav  bool
	}{
		{"article-1", []string{tagTech}, true},                  // tech AND favorite
		{"article-2", []string{tagTech}, false},                 // tech only
		{"article-3", []string{tagProgramming}, true},           // programming AND favorite
		{"article-4", []string{tagProgramming}, false},          // programming only
		{"article-5", []string{}, true},                         // favorite only
		{"article-6", []string{}, false},                        // neither
		{"article-7", []string{tagTech, tagProgramming}, true},  // both tags AND favorite
		{"article-8", []string{tagTech, tagProgramming}, false}, // both tags only
	}

	for _, tc := range testCases {
		now := time.Now()
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 tc.id,
			URL:                "https://example.com/" + tc.id,
			Title:              "Article " + tc.id,
			Content:            "Content " + tc.id,
			CreatedAt:          now,
			Favorite:           tc.fav,
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           testLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testDomain,
		}

		err := s.repository.Store(ctx, article)
		s.NoError(err)

		// Add tags
		err = s.repository.AddTagsToArticle(ctx, account, tc.id, tc.tags, &article.CreatedAt)
		s.NoError(err)
	}

	// Test: favorite=true AND tag=tagTech
	tag := tagTech
	favorite := true
	articles, total, err := s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(2, total)
	s.Len(articles, 2)

	ids := make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		s.True(a.Favorite)
	}
	s.Contains(ids, "article-1")
	s.Contains(ids, "article-7")

	// Test: favorite=true AND tag=tagProgramming
	tag = tagProgramming
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(2, total)
	s.Len(articles, 2)

	ids = make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		s.True(a.Favorite)
	}
	s.Contains(ids, "article-3")
	s.Contains(ids, "article-7")

	// Test: favorite=false AND tag=tagTech
	favorite = false
	tag = tagTech
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(2, total)
	s.Len(articles, 2)

	ids = make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		s.False(a.Favorite)
	}
	s.Contains(ids, "article-2")
	s.Contains(ids, "article-8")

	// Test: favorite=true AND non-existent tag
	tag = "nonexistent"
	favorite = true
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(0, total)
	s.Empty(articles)

	// Test: favorite=false AND non-existent tag
	favorite = false
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(0, total)
	s.Empty(articles)
}

func (s *SQLiteRepositoryTestSuite) TestGetMetadataByAccount_TagFilterPagination() {
	ctx := context.Background()

	const tagTech = tagTech

	account := "test-account-tag-pag"

	// Create 25 articles with tagTech tag
	for i := 1; i <= 25; i++ {
		now := time.Now().Add(-time.Duration(i) * time.Hour)
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 "article-tech-" + string(rune('a'+(i-1)%26)),
			URL:                "https://example.com/tech" + string(rune('a'+(i-1)%26)),
			Title:              "Tech Article " + string(rune('a'+(i-1)%26)),
			Content:            "Tech Content " + string(rune('a'+(i-1)%26)),
			CreatedAt:          now,
			Favorite:           i%2 == 0, // Half are favorites
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           testLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testDomain,
		}

		err := s.repository.Store(ctx, article)
		s.NoError(err)

		err = s.repository.AddTagsToArticle(ctx, account, article.ID, []string{tagTech}, &article.CreatedAt)
		s.NoError(err)
	}

	// Create 5 articles with "other" tag
	for i := 1; i <= 5; i++ {
		now := time.Now().Add(-time.Duration(i) * time.Hour)
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 "article-other-" + string(rune('a'+i-1)),
			URL:                "https://example.com/other" + string(rune('a'+i-1)),
			Title:              "Other Article " + string(rune('a'+i-1)),
			Content:            "Other Content " + string(rune('a'+i-1)),
			CreatedAt:          now,
			Favorite:           false,
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           testLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testDomain,
		}

		err := s.repository.Store(ctx, article)
		s.NoError(err)

		err = s.repository.AddTagsToArticle(ctx, account, article.ID, []string{"other"}, &article.CreatedAt)
		s.NoError(err)
	}

	// Test pagination with tag filter only
	tag := tagTech
	articles, total, err := s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	s.NoError(err)
	s.Equal(25, total)
	s.Len(articles, 10)

	// Test pagination with tag and favorite filters
	tag = tagTech
	favorite := true
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(12, total) // Half of 25 are favorites
	s.Len(articles, 10)

	// Test second page with combined filters
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 2, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(12, total)
	s.Len(articles, 2)

	// Test third page (empty result) with combined filters
	articles, total, err = s.repository.GetMetadataByAccount(
		ctx, account, 3, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	s.NoError(err)
	s.Equal(12, total)
	s.Empty(articles)
}
