package repository

import (
	"context"
	"time"

	internaltypes "github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	articleTestAccount = "test-account"
)

func (s *DynamoDBRepositoryTestSuite) TestStoreAndGetArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	publishedAt := now

	article := &model.Article{
		Account:            articleTestAccount,
		ID:                 "article-1",
		URL:                "https://example.com/article",
		Title:              "Test Article",
		Content:            "This is test content",
		Author:             "Test Author",
		Excerpt:            "Test excerpt",
		ImageURL:           "https://example.com/image.jpg",
		Language:           "en",
		PublishedAt:        &publishedAt,
		WordCount:          100,
		ReadingTimeMinutes: 1,
		SiteName:           "Example Site",
		SourceDomain:       "example.com",
		Favorite:           false,
		CreatedAt:          now,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, article.Account, retrieved.Account)
	assert.Equal(t, article.ID, retrieved.ID)
	assert.Equal(t, article.URL, retrieved.URL)
	assert.Equal(t, article.Title, retrieved.Title)
	assert.Equal(t, article.Content, retrieved.Content)
	assert.Equal(t, article.Author, retrieved.Author)
	assert.Equal(t, article.Excerpt, retrieved.Excerpt)
	assert.Equal(t, article.ImageURL, retrieved.ImageURL)
	assert.Equal(t, article.Language, retrieved.Language)
	assert.Equal(t, article.WordCount, retrieved.WordCount)
	assert.Equal(t, article.ReadingTimeMinutes, retrieved.ReadingTimeMinutes)
	assert.Equal(t, article.SiteName, retrieved.SiteName)
	assert.Equal(t, article.SourceDomain, retrieved.SourceDomain)
	assert.Equal(t, article.Favorite, retrieved.Favorite)
}

func (s *DynamoDBRepositoryTestSuite) TestGetNonExistentArticle() {
	ctx := context.Background()
	t := s.T()

	article, err := s.repositories.GetByAccountAndID(ctx, "nonexistent", "article")
	require.Error(t, err)
	assert.Nil(t, article)
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateFavorite() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	article := &model.Article{
		Account:   articleTestAccount,
		ID:        "article-2",
		URL:       "https://example.com/article2",
		Title:     "Test Article 2",
		Content:   "Content 2",
		CreatedAt: now,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	err = s.repositories.UpdateFavorite(ctx, article.Account, article.ID, true)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.Favorite)
}

func (s *DynamoDBRepositoryTestSuite) TestDeleteArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	article := &model.Article{
		Account:   articleTestAccount,
		ID:        "article-3",
		URL:       "https://example.com/article3",
		Title:     "Test Article 3",
		Content:   "Content 3",
		CreatedAt: now,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	err = s.repositories.DeleteByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)

	_, err = s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.Error(t, err)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccount() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-query"

	for i := 1; i <= 5; i++ {
		now := time.Now().Add(-time.Duration(i) * time.Hour)
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 "article-" + string(rune('0'+i)),
			URL:                "https://example.com/article" + string(rune('0'+i)),
			Title:              "Test Article " + string(rune('0'+i)),
			Content:            "Content " + string(rune('0'+i)),
			CreatedAt:          now,
			Favorite:           i%2 == 0,
			Author:             "Author",
			Excerpt:            "Excerpt",
			ImageURL:           "https://example.com/image.jpg",
			Language:           "en",
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           "Example Site",
			SourceDomain:       "example.com",
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	articles, _, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, articles, 5)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountWithFavoriteFilter() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-fav"

	for i := 1; i <= 4; i++ {
		now := time.Now().Add(-time.Duration(i) * time.Hour)
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 "article-fav-" + string(rune('0'+i)),
			URL:                "https://example.com/fav" + string(rune('0'+i)),
			Title:              "Fav Article " + string(rune('0'+i)),
			Content:            "Fav Content " + string(rune('0'+i)),
			CreatedAt:          now,
			Favorite:           false, // Start as non-favorite
			Author:             "Author",
			Excerpt:            "Excerpt",
			ImageURL:           "https://example.com/image.jpg",
			Language:           "en",
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           "Example Site",
			SourceDomain:       "example.com",
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		// Toggle favorites
		if i%2 == 0 {
			err = s.repositories.UpdateFavorite(ctx, account, "article-fav-"+string(rune('0'+i)), true)
			require.NoError(t, err)
		}
	}

	favorite := true
	articles, _, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, articles, 2)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountPagination() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-pag"

	for i := 1; i <= 25; i++ {
		now := time.Now().Add(-time.Duration(i) * time.Hour)
		publishedAt := now

		article := &model.Article{
			Account:            account,
			ID:                 "article-pag-" + string(rune('0'+i)),
			URL:                "https://example.com/pag" + string(rune('0'+i)),
			Title:              "Pag Article " + string(rune('0'+i)),
			Content:            "Pag Content " + string(rune('0'+i)),
			CreatedAt:          now,
			Favorite:           false,
			Author:             "Author",
			Excerpt:            "Excerpt",
			ImageURL:           "https://example.com/image.jpg",
			Language:           "en",
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           "Example Site",
			SourceDomain:       "example.com",
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	articles, _, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	articles2, _, total2, err := s.repositories.GetMetadataByAccount(ctx, account, 2, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total2)
	assert.Len(t, articles2, 10)

	articles3, _, total3, err := s.repositories.GetMetadataByAccount(ctx, account, 3, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total3)
	assert.Len(t, articles3, 5)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountEmpty() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-empty"

	articles, _, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
	require.NoError(t, err)
	assert.Empty(t, articles)
	assert.Equal(t, 0, total)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountOffsetOutOfBounds() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-offset"

	now := time.Now()
	publishedAt := now

	article := &model.Article{
		Account:            account,
		ID:                 "article-offset",
		URL:                "https://example.com/offset",
		Title:              "Offset Article",
		Content:            "Content",
		CreatedAt:          now,
		Favorite:           false,
		Author:             "Author",
		Excerpt:            "Excerpt",
		ImageURL:           "https://example.com/image.jpg",
		Language:           "en",
		PublishedAt:        &publishedAt,
		WordCount:          100,
		ReadingTimeMinutes: 1,
		SiteName:           "Example Site",
		SourceDomain:       "example.com",
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	articles, _, total, err := s.repositories.GetMetadataByAccount(ctx, account, 100, 10, nil)
	require.NoError(t, err)
	assert.Empty(t, articles)
	assert.Equal(t, 1, total)
}

func (s *DynamoDBRepositoryTestSuite) TestStoreArticleEmptyAccount() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	article := &model.Article{
		ID:        "article-empty-account",
		URL:       "https://example.com/article",
		Title:     "Test Article",
		Content:   "Content",
		CreatedAt: now,
	}

	err := s.repositories.Store(ctx, article)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account field is required")
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateFavorite_SetsAccountFavorite() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	article := &model.Article{
		Account:   articleTestAccount,
		ID:        "article-fav-on",
		URL:       "https://example.com/article",
		Title:     "Test Article",
		Content:   "Content",
		CreatedAt: now,
		Favorite:  false,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	err = s.repositories.UpdateFavorite(ctx, article.Account, article.ID, true)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.Favorite)
}

func (s *DynamoDBRepositoryTestSuite) TestUpdateFavorite_RemovesAccountFavorite() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()

	article := &model.Article{
		Account:   articleTestAccount,
		ID:        "article-fav-off",
		URL:       "https://example.com/article",
		Title:     "Test Article",
		Content:   "Content",
		CreatedAt: now,
		Favorite:  true,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	retrieved, err := s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.Favorite)

	err = s.repositories.UpdateFavorite(ctx, article.Account, article.ID, false)
	require.NoError(t, err)

	retrieved, err = s.repositories.GetByAccountAndID(ctx, article.Account, article.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.Favorite)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccount_FavoritesPagination() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-fav-pag"

	// Create 25 articles, all favorites
	for i := range 25 {
		article := &model.Article{
			Account:   account,
			ID:        "article-" + string(rune('a'+(i%26))),
			URL:       "https://example.com/article",
			Title:     "Test Article",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			Favorite:  false, // Start as non-favorite
		}
		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		// Mark as favorite
		err = s.repositories.UpdateFavorite(ctx, account, article.ID, true)
		require.NoError(t, err)
	}

	// First page
	favorite := true
	articles, _, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}

	// Second page
	articles, _, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 2, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}

	// Third page
	articles, _, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 3, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 5)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccount_FavoritesEmptyResult() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-fav-empty"

	// Create 5 non-favorite articles
	for i := range 5 {
		article := &model.Article{
			Account:   account,
			ID:        "article-" + string(rune('a'+i)),
			URL:       "https://example.com/article",
			Title:     "Test Article",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			Favorite:  false,
		}
		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	// Query for favorites - should return empty
	favorite := true
	articles, _, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, articles)
}
