// Package repository provides tests for article persistence implementations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccount = "test@example.com"
)

func isResourceNotFound(err error) bool {
	var smithyErr interface {
		ErrorCode() string
	}
	return errors.As(err, &smithyErr) && smithyErr.ErrorCode() == "ResourceNotFoundException"
}

func skipIfTableNotFound(t *testing.T, err error) {
	t.Helper()
	if isResourceNotFound(err) {
		t.Skip("DynamoDB table not found, skipping integration test")
	}
}

func TestDynamoDB_Store(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	article := &model.Article{
		Account:   testAccount,
		ID:        "test-id-1",
		URL:       "https://example.com/test",
		Title:     "Test Article",
		Content:   "<p>Test content</p>",
		SiteName:  "Example Site",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
}

func TestDynamoDB_Store_RequiresAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	article := &model.Article{
		ID:        "test-id-1",
		URL:       "https://example.com/test",
		Title:     "Test Article",
		Content:   "<p>Test content</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, article)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account field is required")
}

func TestDynamoDB_GetByAccountAndID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	expected := &model.Article{
		Account:   testAccount,
		ID:        "test-id-2",
		URL:       "https://example.com/test2",
		Title:     "Test Article 2",
		Content:   "<p>Test content 2</p>",
		SiteName:  "Example Site",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, expected)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	actual, err := repo.GetByAccountAndID(ctx, testAccount, "test-id-2")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, expected.Account, actual.Account)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.URL, actual.URL)
	assert.Equal(t, expected.Title, actual.Title)
	assert.Equal(t, expected.Content, actual.Content, "content field should be included for GetByAccountAndID")
}

func TestDynamoDB_GetByAccountAndID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	_, err := repo.GetByAccountAndID(ctx, testAccount, "non-existent-id")
	skipIfTableNotFound(t, err)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestDynamoDB_GetByAccountAndID_WrongAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	article := &model.Article{
		Account:   "user1@example.com",
		ID:        "test-id-3",
		URL:       "https://example.com/test3",
		Title:     "Test Article 3",
		Content:   "<p>Test content 3</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	_, err = repo.GetByAccountAndID(ctx, "user2@example.com", "test-id-3")
	skipIfTableNotFound(t, err)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestDynamoDB_GetMetadataByAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := testAccount
	articles := []*model.Article{
		{
			Account:   account,
			ID:        "test-id-4",
			URL:       "https://example.com/test4",
			Title:     "Test Article 4",
			Content:   "<p>Test content 4</p>",
			CreatedAt: time.Now().UTC(),
		},
		{
			Account:   account,
			ID:        "test-id-5",
			URL:       "https://example.com/test5",
			Title:     "Test Article 5",
			Content:   "<p>Test content 5</p>",
			CreatedAt: time.Now().UTC(),
		},
		{
			Account:   "other@example.com",
			ID:        "test-id-6",
			URL:       "https://example.com/test6",
			Title:     "Test Article 6",
			Content:   "<p>Test content 6</p>",
			CreatedAt: time.Now().UTC(),
		},
	}

	for _, article := range articles {
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	retrieved, _, _, err := repo.GetMetadataByAccount(ctx, account, 1, 20, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)

	accountIDs := make(map[string]bool)
	for _, article := range retrieved {
		accountIDs[article.ID] = true
		assert.Empty(t, article.Content, "content field should be excluded")
	}
	assert.True(t, accountIDs["test-id-4"])
	assert.True(t, accountIDs["test-id-5"])
	assert.False(t, accountIDs["test-id-6"])
}

func TestDynamoDB_GetByAccount_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	retrieved, _, _, err := repo.GetMetadataByAccount(ctx, "non-existent@example.com", 1, 20, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestDynamoDB_DeleteByAccountAndID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	article := &model.Article{
		Account:   testAccount,
		ID:        "test-id-7",
		URL:       "https://example.com/test7",
		Title:     "Test Article 7",
		Content:   "<p>Test content 7</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	err = repo.DeleteByAccountAndID(ctx, testAccount, "test-id-7")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	_, err = repo.GetByAccountAndID(ctx, testAccount, "test-id-7")
	skipIfTableNotFound(t, err)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestDynamoDB_DeleteByAccountAndID_WrongAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	article := &model.Article{
		Account:   "user1@example.com",
		ID:        "test-id-8",
		URL:       "https://example.com/test8",
		Title:     "Test Article 8",
		Content:   "<p>Test content 8</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	err = repo.DeleteByAccountAndID(ctx, "user2@example.com", "test-id-8")
	skipIfTableNotFound(t, err)
	assert.Error(t, err)

	_, err = repo.GetByAccountAndID(ctx, "user1@example.com", "test-id-8")
	skipIfTableNotFound(t, err)
	assert.NoError(t, err)
}

func TestDynamoDB_DeleteByAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := testAccount
	articles := []*model.Article{
		{
			Account:   account,
			ID:        "test-id-10",
			URL:       "https://example.com/test10",
			Title:     "Test Article 10",
			Content:   "<p>Test content 10</p>",
			CreatedAt: time.Now().UTC(),
		},
		{
			Account:   account,
			ID:        "test-id-11",
			URL:       "https://example.com/test11",
			Title:     "Test Article 11",
			Content:   "<p>Test content 11</p>",
			CreatedAt: time.Now().UTC(),
		},
		{
			Account:   "other@example.com",
			ID:        "test-id-12",
			URL:       "https://example.com/test12",
			Title:     "Test Article 12",
			Content:   "<p>Test content 12</p>",
			CreatedAt: time.Now().UTC(),
		},
	}

	for _, article := range articles {
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	deleted, err := repo.DeleteByAccount(ctx, account)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 2, deleted)

	retrieved, _, _, err := repo.GetMetadataByAccount(ctx, account, 1, 20, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Empty(t, retrieved)

	otherArticles, _, _, err := repo.GetMetadataByAccount(ctx, "other@example.com", 1, 20, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Len(t, otherArticles, 1)
}

func TestDynamoDB_DeleteByAccount_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	deleted, err := repo.DeleteByAccount(ctx, "non-existent@example.com")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_MultiplePages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "pagination@example.com"
	numArticles := 25

	for i := range numArticles {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("pagination-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/pagination%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	page1, _, total, err := repo.GetMetadataByAccount(ctx, account, 1, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 10, len(page1))
	assert.Equal(t, numArticles, total)

	page2, _, total, err := repo.GetMetadataByAccount(ctx, account, 2, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 10, len(page2))
	assert.Equal(t, numArticles, total)

	page3, _, total, err := repo.GetMetadataByAccount(ctx, account, 3, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 5, len(page3))
	assert.Equal(t, numArticles, total)

	allIDs := make(map[string]bool)
	for _, article := range page1 {
		allIDs[article.ID] = true
	}
	for _, article := range page2 {
		assert.False(t, allIDs[article.ID], "page 2 should not contain articles from page 1")
		allIDs[article.ID] = true
	}
	for _, article := range page3 {
		assert.False(t, allIDs[article.ID], "page 3 should not contain articles from pages 1-2")
		allIDs[article.ID] = true
	}
	assert.Len(t, allIDs, numArticles)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_LastPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "lastpage@example.com"
	numArticles := 15

	for i := range numArticles {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("lastpage-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/lastpage%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	page1, lastKey1, total, err := repo.GetMetadataByAccount(ctx, account, 1, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 10, len(page1))
	assert.NotNil(t, lastKey1, "lastEvaluatedKey should be non-nil when there are more results")
	assert.Equal(t, numArticles, total)

	page2, lastKey2, total, err := repo.GetMetadataByAccount(ctx, account, 2, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 5, len(page2))
	assert.Nil(t, lastKey2, "lastEvaluatedKey should be nil on the last page")
	assert.Equal(t, numArticles, total)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_PageOutOfRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "outofrange@example.com"
	numArticles := 5

	for i := range numArticles {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("outofrange-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/outofrange%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	articles, lastKey, total, err := repo.GetMetadataByAccount(ctx, account, 10, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Empty(t, articles)
	assert.Nil(t, lastKey)
	assert.Equal(t, numArticles, total)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_TotalCountAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "totalcount@example.com"
	numArticles := 12

	for i := range numArticles {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("totalcount-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/totalcount%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	page1, _, total1, err := repo.GetMetadataByAccount(ctx, account, 1, 5, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 5, len(page1))
	assert.Equal(t, numArticles, total1)

	page2, _, total2, err := repo.GetMetadataByAccount(ctx, account, 2, 5, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 5, len(page2))
	assert.Equal(t, numArticles, total2)

	page3, _, total3, err := repo.GetMetadataByAccount(ctx, account, 3, 5, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, 2, len(page3))
	assert.Equal(t, numArticles, total3)

	assert.Equal(t, total1, total2)
	assert.Equal(t, total2, total3)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_LargePageSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "largepagesize@example.com"
	numArticles := 15

	for i := range numArticles {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("largepagesize-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/largepagesize%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	articles, _, total, err := repo.GetMetadataByAccount(ctx, account, 1, 100, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, numArticles, len(articles))
	assert.Equal(t, numArticles, total)
}

func TestDynamoDB_GetMetadataByAccount_Pagination_InvalidPageDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "invalidpage@example.com"

	for i := range 5 {
		article := &model.Article{
			Account:   account,
			ID:        fmt.Sprintf("invalidpage-id-%03d", i),
			URL:       fmt.Sprintf("https://example.com/invalidpage%d", i),
			Title:     fmt.Sprintf("Article %d", i),
			Content:   fmt.Sprintf("<p>Content %d</p>", i),
			CreatedAt: time.Now().UTC(),
		}
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	articles, _, _, err := repo.GetMetadataByAccount(ctx, account, 0, 10, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Len(t, articles, 5)
}

func TestDynamoDB_UpdateArticle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	original := &model.Article{
		Account:   testAccount,
		ID:        "test-id-9",
		URL:       "https://example.com/test9",
		Title:     "Test Article 9",
		Content:   "<p>Test content 9</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Store(ctx, original)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	updated := &model.Article{
		Account:            testAccount,
		ID:                 "test-id-9",
		URL:                "https://example.com/test9",
		Title:              "Updated Article 9",
		Content:            "<p>Updated content 9</p>",
		CreatedAt:          original.CreatedAt,
		DeliveryStatus:     consts.StatusDelivered,
		DeliveredFrom:      stringPtr("sender@example.com"),
		DeliveredTo:        stringPtr("kindle@example.com"),
		DeliveredEmailUUID: stringPtr("email-uuid-123"),
		DeliveredBy:        consts.EmailBackendMailjet,
	}

	err = repo.Store(ctx, updated)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	retrieved, err := repo.GetByAccountAndID(ctx, testAccount, "test-id-9")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, "Updated Article 9", retrieved.Title)
	assert.Equal(t, consts.StatusDelivered, retrieved.DeliveryStatus)
	assert.Equal(t, "sender@example.com", *retrieved.DeliveredFrom)
}

func TestDynamoDB_Store_NormalizesToUTC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	localTime := time.Date(2024, 2, 14, 10, 0, 0, 0, time.FixedZone("EST", -4*3600))
	article := &model.Article{
		Account:   testAccount,
		ID:        "utc-test-id",
		URL:       "https://example.com/utc-test",
		Title:     "UTC Test Article",
		Content:   "<p>Test content</p>",
		CreatedAt: localTime,
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	retrieved, err := repo.GetByAccountAndID(ctx, testAccount, "utc-test-id")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	assert.True(t, retrieved.CreatedAt.UTC().Equal(localTime.UTC()), "timestamp should be stored in UTC")
}

func TestDynamoDB_GetMetadataByAccount_SortsByCreatedAt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	account := "sort-test@example.com"
	now := time.Now().UTC()
	articles := []*model.Article{
		{
			Account:   account,
			ID:        "sort-id-1",
			URL:       "https://example.com/sort1",
			Title:     "Article 1",
			Content:   "<p>Content 1</p>",
			CreatedAt: now.Add(-4 * time.Hour),
		},
		{
			Account:   account,
			ID:        "sort-id-2",
			URL:       "https://example.com/sort2",
			Title:     "Article 2",
			Content:   "<p>Content 2</p>",
			CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			Account:   account,
			ID:        "sort-id-3",
			URL:       "https://example.com/sort3",
			Title:     "Article 3",
			Content:   "<p>Content 3</p>",
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			Account:   account,
			ID:        "sort-id-4",
			URL:       "https://example.com/sort4",
			Title:     "Article 4",
			Content:   "<p>Content 4</p>",
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			Account:   account,
			ID:        "sort-id-5",
			URL:       "https://example.com/sort5",
			Title:     "Article 5",
			Content:   "<p>Content 5</p>",
			CreatedAt: now,
		},
	}

	for _, article := range articles {
		err := repo.Store(ctx, article)
		skipIfTableNotFound(t, err)
		require.NoError(t, err)
	}

	retrieved, _, _, err := repo.GetMetadataByAccount(ctx, account, 1, 20, nil)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Len(t, retrieved, 5)

	for i := 0; i < len(retrieved)-1; i++ {
		assert.True(t, retrieved[i].CreatedAt.After(retrieved[i+1].CreatedAt),
			"article %d (%s) should be newer than article %d (%s)",
			i, retrieved[i].CreatedAt, i+1, retrieved[i+1].CreatedAt)
	}

	assert.Equal(t, "Article 5", retrieved[0].Title, "first article should be newest")
	assert.Equal(t, "Article 1", retrieved[4].Title, "last article should be oldest")
}

func TestDynamoDB_Store_PreservesUTC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	utcTime := time.Date(2024, 2, 14, 14, 0, 0, 0, time.UTC)
	article := &model.Article{
		Account:   testAccount,
		ID:        "utc-preserve-test",
		URL:       "https://example.com/utc-preserve",
		Title:     "UTC Preserve Test",
		Content:   "<p>Test content</p>",
		CreatedAt: utcTime,
	}

	err := repo.Store(ctx, article)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	retrieved, err := repo.GetByAccountAndID(ctx, testAccount, "utc-preserve-test")
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	assert.True(t, retrieved.CreatedAt.Equal(utcTime), "UTC timestamp should be preserved exactly")
	assert.Equal(t, time.UTC, retrieved.CreatedAt.Location(), "timestamp location should be UTC")
}

func TestDynamoDB_GetUserProfile_NotSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	profile, err := repo.GetUserProfile(ctx, testAccount)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Nil(t, profile)
}

func TestDynamoDB_PutUserProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	profile := &model.UserProfile{
		Account:     testAccount,
		KindleEmail: "user@kindle.com",
	}

	err := repo.PutUserProfile(ctx, profile)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
}

func TestDynamoDB_GetUserProfile_AfterPut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	expected := &model.UserProfile{
		Account:     testAccount,
		KindleEmail: "user@kindle.com",
	}

	err := repo.PutUserProfile(ctx, expected)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	actual, err := repo.GetUserProfile(ctx, testAccount)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, expected.Account, actual.Account)
	assert.Equal(t, expected.KindleEmail, actual.KindleEmail)
}

func TestDynamoDB_PutUserProfile_UpdatesExisting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	profile1 := &model.UserProfile{
		Account:     testAccount,
		KindleEmail: "user@kindle.com",
	}

	err := repo.PutUserProfile(ctx, profile1)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	profile2 := &model.UserProfile{
		Account:     testAccount,
		KindleEmail: "updated@kindle.com",
	}

	err = repo.PutUserProfile(ctx, profile2)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)

	actual, err := repo.GetUserProfile(ctx, testAccount)
	skipIfTableNotFound(t, err)
	require.NoError(t, err)
	assert.Equal(t, "updated@kindle.com", actual.KindleEmail)
}

func TestDynamoDB_PutUserProfile_RequiresAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo := setupTestDynamoDB(t)
	ctx := context.Background()

	profile := &model.UserProfile{
		KindleEmail: "user@kindle.com",
	}

	err := repo.PutUserProfile(ctx, profile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account field is required")
}

func stringPtr(s string) *string {
	return &s
}

func setupTestDynamoDB(t *testing.T) *DynamoDB {
	t.Helper()

	articlesTableName := "test-savetoink-articles"
	profileTableName := "test-savetoink-user-profiles"
	repo := NewDynamoDB(nil, articlesTableName, profileTableName)

	t.Cleanup(func() {
		ctx := context.Background()
		articles, _, _, err := repo.GetMetadataByAccount(ctx, testAccount, 1, 20, nil)
		if err != nil {
			return
		}

		for _, article := range articles {
			_ = repo.DeleteByAccountAndID(ctx, article.Account, article.ID)
		}
	})

	return repo
}
