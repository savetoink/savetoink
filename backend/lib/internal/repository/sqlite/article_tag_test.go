// Package repository provides SQLite repository implementations.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTagStr      = "test"
	testTagArticle2 = "tag-article-2"
	testTagArticle3 = "tag-article-3"
	testSharedTag   = "shared-tag"
)

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-1",
		URL:       testURL,
		CreatedAt: now,
	}

	// Create the article first
	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add tags
	tags := []string{tagTech, tagProgramming, tagGolang}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, nil)
	require.NoError(t, err)

	// Verify tags were added
	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags, retrievedTags)
}

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticleWithCreatedAt() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create the article first
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-with-created",
		URL:       testURL,
		CreatedAt: now,
	}
	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add tags using the createdAt directly (no extra DB query)
	tags := []string{tagTech, tagProgramming, tagGolang}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, &article.CreatedAt)
	require.NoError(t, err)

	// Verify tags were added
	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags, retrievedTags)
}

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticle_BothMethodsSameResult() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create two articles
	article1 := &model.Article{
		Account:   testAccount,
		ID:        "tag-compare-1",
		URL:       testURL1,
		CreatedAt: now,
	}
	article2 := &model.Article{
		Account:   testAccount,
		ID:        "tag-compare-2",
		URL:       testURL2,
		CreatedAt: now,
	}
	err := s.repository.Store(ctx, article1)
	require.NoError(t, err)
	err = s.repository.Store(ctx, article2)
	require.NoError(t, err)

	// Add tags using both methods with the same createdAt
	tags := []string{tagTech, tagProgramming, tagGolang}
	err = s.repository.AddTagsToArticle(ctx, article1.Account, article1.ID, tags, nil)
	require.NoError(t, err)
	err = s.repository.AddTagsToArticle(ctx, article2.Account, article2.ID, tags, &article2.CreatedAt)
	require.NoError(t, err)

	// Both should produce the same result
	tags1, err := s.repository.GetArticleTags(ctx, article1.Account, article1.ID)
	require.NoError(t, err)
	tags2, err := s.repository.GetArticleTags(ctx, article2.Account, article2.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags, tags1)
	assert.ElementsMatch(t, tags, tags2)
}

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticle_EmptyTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-empty",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add empty tags - should be no-op
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, []string{}, nil)
	require.NoError(t, err)

	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.Empty(t, retrievedTags)
}

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticle_NonExistentArticle() {
	ctx := context.Background()
	t := s.T()

	tags := []string{tagTech, tagProgramming}
	err := s.repository.AddTagsToArticle(ctx, testAccount, "nonexistent-article", tags, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "article not found")
}

func (s *SQLiteRepositoryTestSuite) TestAddTagsToArticle_DuplicateTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-duplicate",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add tags
	tags := []string{tagTech, tagProgramming}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, nil)
	require.NoError(t, err)

	// Try adding the same tags again - should succeed but not create duplicates
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, nil)
	require.NoError(t, err)

	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	// Should still have only unique tags
	assert.ElementsMatch(t, tags, retrievedTags)
	assert.Len(t, retrievedTags, 2)
}

func (s *SQLiteRepositoryTestSuite) TestRemoveTagsFromArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-remove",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add tags
	tags := []string{tagTech, tagProgramming, tagGolang, testTagStr}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, nil)
	require.NoError(t, err)

	// Remove some tags
	removeTags := []string{tagTech, testTagStr}
	err = s.repository.RemoveTagsFromArticle(ctx, article.Account, article.ID, removeTags)
	require.NoError(t, err)

	// Verify tags were removed
	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{tagProgramming, tagGolang}, retrievedTags)
	assert.Len(t, retrievedTags, 2)
}

func (s *SQLiteRepositoryTestSuite) TestRemoveTagsFromArticle_EmptyTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-remove-empty",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Remove empty tags - should be no-op
	err = s.repository.RemoveTagsFromArticle(ctx, article.Account, article.ID, []string{})
	require.NoError(t, err)
}

func (s *SQLiteRepositoryTestSuite) TestRemoveTagsFromArticle_NonExistentArticle() {
	ctx := context.Background()
	t := s.T()

	tags := []string{tagTech}
	err := s.repository.RemoveTagsFromArticle(ctx, testAccount, "nonexistent-article", tags)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "article not found")
}

func (s *SQLiteRepositoryTestSuite) TestSetArticleTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-set",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Set initial tags
	tags1 := []string{tagTech, tagProgramming}
	err = s.repository.SetArticleTags(ctx, article.Account, article.ID, tags1)
	require.NoError(t, err)

	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags1, retrievedTags)

	// Replace with new tags
	tags2 := []string{tagGolang, tagDatabase, "aws"}
	err = s.repository.SetArticleTags(ctx, article.Account, article.ID, tags2)
	require.NoError(t, err)

	retrievedTags, err = s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags2, retrievedTags)
	assert.Len(t, retrievedTags, 3)
}

func (s *SQLiteRepositoryTestSuite) TestSetArticleTags_EmptyTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-set-empty",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Set initial tags
	tags := []string{tagTech, tagProgramming}
	err = s.repository.SetArticleTags(ctx, article.Account, article.ID, tags)
	require.NoError(t, err)

	// Set to empty - should remove all tags
	err = s.repository.SetArticleTags(ctx, article.Account, article.ID, []string{})
	require.NoError(t, err)

	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.Empty(t, retrievedTags)
}

func (s *SQLiteRepositoryTestSuite) TestGetArticleTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-get",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Test getting tags for article with no tags
	tags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.Empty(t, tags)

	// Add tags
	expectedTags := []string{tagTech, tagProgramming, tagGolang}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, expectedTags, nil)
	require.NoError(t, err)

	// Get tags
	tags, err = s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedTags, tags)
}

func (s *SQLiteRepositoryTestSuite) TestGetArticlesByTag() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create multiple articles with different tags
	articles := []struct {
		id   string
		tags []string
	}{
		{tagArticle1, []string{tagTech, tagProgramming}},
		{testTagArticle2, []string{tagTech, tagGolang}},
		{testTagArticle3, []string{tagProgramming, tagDatabase}},
	}

	for _, art := range articles {
		article := &model.Article{
			Account:   testAccount,
			ID:        art.id,
			URL:       "https://example.com/" + art.id,
			CreatedAt: now,
		}
		err := s.repository.Store(ctx, article)
		require.NoError(t, err)
		err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, art.tags, nil)
		require.NoError(t, err)
	}

	// Get articles by tagTech tag
	articleIDs, total, err := s.repository.GetArticlesByTag(ctx, testAccount, tagTech, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.ElementsMatch(t, []string{tagArticle1, testTagArticle2}, articleIDs)

	// Get articles by tagProgramming tag
	articleIDs, total, err = s.repository.GetArticlesByTag(ctx, testAccount, tagProgramming, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.ElementsMatch(t, []string{tagArticle1, testTagArticle3}, articleIDs)
}

func (s *SQLiteRepositoryTestSuite) TestGetArticlesByTag_Pagination() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create multiple articles with the same tag
	articleCount := 5
	articleIDs := make([]string, articleCount)
	for i := range articleCount {
		id := fmt.Sprintf("tag-page-%d", i)
		articleIDs[i] = id
		article := &model.Article{
			Account:   testAccount,
			ID:        id,
			URL:       "https://example.com/" + id,
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		}
		err := s.repository.Store(ctx, article)
		require.NoError(t, err)
		err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, []string{"pagetag"}, nil)
		require.NoError(t, err)
	}

	// First page
	ids, total, err := s.repository.GetArticlesByTag(ctx, testAccount, "pagetag", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, articleCount, total)
	assert.Len(t, ids, 2)

	// Second page
	ids, total, err = s.repository.GetArticlesByTag(ctx, testAccount, "pagetag", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, articleCount, total)
	assert.Len(t, ids, 2)

	// Third page (last page)
	ids, total, err = s.repository.GetArticlesByTag(ctx, testAccount, "pagetag", 3, 2)
	require.NoError(t, err)
	assert.Equal(t, articleCount, total)
	assert.Len(t, ids, 1)
}

func (s *SQLiteRepositoryTestSuite) TestGetAllTagsForAccount() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create articles with various tags
	articles := []struct {
		id   string
		tags []string
	}{
		{"tag-all-1", []string{tagTech, tagProgramming}},
		{"tag-all-2", []string{tagTech, tagGolang}},
		{"tag-all-3", []string{tagDatabase, "sql"}},
	}

	for _, art := range articles {
		article := &model.Article{
			Account:   testAccount,
			ID:        art.id,
			URL:       "https://example.com/" + art.id,
			CreatedAt: now,
		}
		err := s.repository.Store(ctx, article)
		require.NoError(t, err)
		err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, art.tags, nil)
		require.NoError(t, err)
	}

	// Get all tags for account
	tags, err := s.repository.GetAllTagsForAccount(ctx, testAccount)
	require.NoError(t, err)
	assert.Contains(t, tags, tagTech)
	assert.Contains(t, tags, tagProgramming)
	assert.Contains(t, tags, tagGolang)
	assert.Contains(t, tags, tagDatabase)
	assert.Contains(t, tags, "sql")
}

func (s *SQLiteRepositoryTestSuite) TestDeleteTagsForArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-delete",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Add tags
	tags := []string{tagTech, tagProgramming, tagGolang}
	err = s.repository.AddTagsToArticle(ctx, article.Account, article.ID, tags, nil)
	require.NoError(t, err)

	// Verify tags exist
	retrievedTags, err := s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tags, retrievedTags)

	// Delete all tags
	err = s.repository.DeleteTagsForArticle(ctx, article.Account, article.ID)
	require.NoError(t, err)

	// Verify tags are deleted
	retrievedTags, err = s.repository.GetArticleTags(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.Empty(t, retrievedTags)
}

func (s *SQLiteRepositoryTestSuite) TestDeleteTagsForArticle_NoTags() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	article := &model.Article{
		Account:   testAccount,
		ID:        "tag-test-delete-no-tags",
		URL:       testURL,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article)
	require.NoError(t, err)

	// Delete tags from article that has no tags - should succeed
	err = s.repository.DeleteTagsForArticle(ctx, article.Account, article.ID)
	require.NoError(t, err)
}

func (s *SQLiteRepositoryTestSuite) TestTagIsolation() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	// Create articles for different accounts
	article1 := &model.Article{
		Account:   "account-1",
		ID:        "isolation-1",
		URL:       testURL1,
		CreatedAt: now,
	}
	article2 := &model.Article{
		Account:   "account-2",
		ID:        "isolation-2",
		URL:       testURL2,
		CreatedAt: now,
	}

	err := s.repository.Store(ctx, article1)
	require.NoError(t, err)
	err = s.repository.Store(ctx, article2)
	require.NoError(t, err)

	// Add tags to both articles with same tag name
	err = s.repository.AddTagsToArticle(ctx, article1.Account, article1.ID, []string{testSharedTag}, nil)
	require.NoError(t, err)
	err = s.repository.AddTagsToArticle(ctx, article2.Account, article2.ID, []string{testSharedTag}, nil)
	require.NoError(t, err)

	// Verify both accounts can have the same tag independently
	tags1, err := s.repository.GetArticleTags(ctx, article1.Account, article1.ID)
	require.NoError(t, err)
	assert.Contains(t, tags1, testSharedTag)

	tags2, err := s.repository.GetArticleTags(ctx, article2.Account, article2.ID)
	require.NoError(t, err)
	assert.Contains(t, tags2, testSharedTag)

	// Get all tags for each account - should only see their own articles
	allTags1, err := s.repository.GetAllTagsForAccount(ctx, article1.Account)
	require.NoError(t, err)
	assert.Contains(t, allTags1, testSharedTag)
	assert.Equal(t, 1, len(allTags1))

	allTags2, err := s.repository.GetAllTagsForAccount(ctx, article2.Account)
	require.NoError(t, err)
	assert.Contains(t, allTags2, testSharedTag)
	assert.Equal(t, 1, len(allTags2))
}
