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
	article1           = "article-1"
	testImageURL       = "https://example.com/image.jpg"
	testSiteName       = "Example Site"
	testSourceDomain   = "example.com"
	testAuthor         = "Author"
	testExcerpt        = "Excerpt"
	testContent        = "Content"
	articleTag2        = "article-tag-2"
	articleTestLang    = "en"
	article2ID         = "article-2"
	article3ID         = "article-3"
	article2Title      = "Test Article 2"
	articleTag1        = "article-tag-1"
	articleTag3        = "article-tag-3"
	articleTag4        = "article-tag-4"
)

func (s *DynamoDBRepositoryTestSuite) TestStoreAndGetArticle() {
	ctx := context.Background()
	t := s.T()

	now := time.Now()
	publishedAt := now

	article := &model.Article{
		Account:            articleTestAccount,
		ID:                 article1,
		URL:                tagArticleURL,
		Title:              tagArticleTitle,
		Content:            "This is test content",
		Author:             "Test Author",
		Excerpt:            "Test excerpt",
		ImageURL:           testImageURL,
		Language:           articleTestLang,
		PublishedAt:        &publishedAt,
		WordCount:          100,
		ReadingTimeMinutes: 1,
		SiteName:           testSiteName,
		SourceDomain:       testSourceDomain,
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
		ID:        article2ID,
		URL:       "https://example.com/article2",
		Title:     article2Title,
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
		ID:        article3ID,
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
			URL:                tagArticleURL + string(rune('0'+i)),
			Title:              "Test Article " + string(rune('0'+i)),
			Content:            "Content " + string(rune('0'+i)),
			CreatedAt:          now,
			Favorite:           i%2 == 0,
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	articles, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
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
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
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
	articles, total, err := s.repositories.GetMetadataByAccount(
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
			Author:             testAuthor,
			Excerpt:            testExcerpt,
			ImageURL:           testImageURL,
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	articles, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	articles2, total2, err := s.repositories.GetMetadataByAccount(ctx, account, 2, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total2)
	assert.Len(t, articles2, 10)

	articles3, total3, err := s.repositories.GetMetadataByAccount(ctx, account, 3, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, 25, total3)
	assert.Len(t, articles3, 5)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountEmpty() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-empty"

	articles, total, err := s.repositories.GetMetadataByAccount(ctx, account, 1, 10, nil)
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
		Content:            testContent,
		CreatedAt:          now,
		Favorite:           false,
		Author:             testAuthor,
		Excerpt:            testExcerpt,
		ImageURL:           testImageURL,
		Language:           articleTestLang,
		PublishedAt:        &publishedAt,
		WordCount:          100,
		ReadingTimeMinutes: 1,
		SiteName:           testSiteName,
		SourceDomain:       testSourceDomain,
	}

	err := s.repositories.Store(ctx, article)
	require.NoError(t, err)

	articles, total, err := s.repositories.GetMetadataByAccount(ctx, account, 100, 10, nil)
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
		URL:       tagArticleURL,
		Title:     tagArticleTitle,
		Content:   testContent,
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
		URL:       tagArticleURL,
		Title:     tagArticleTitle,
		Content:   testContent,
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
		URL:       tagArticleURL,
		Title:     tagArticleTitle,
		Content:   testContent,
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
			URL:       tagArticleURL,
			Title:     tagArticleTitle,
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
	articles, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}

	// Second page
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 2, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	for _, article := range articles {
		assert.True(t, article.Favorite)
	}

	// Third page
	articles, total, err = s.repositories.GetMetadataByAccount(
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
			URL:       tagArticleURL,
			Title:     tagArticleTitle,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			Favorite:  false,
		}
		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)
	}

	// Query for favorites - should return empty
	favorite := true
	articles, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, articles)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccountWithTagFilter() {
	ctx := context.Background()
	t := s.T()

	const tagNonexistent = "nonexistent"

	account := "test-account-tag"

	// Create articles with different tags
	testCases := []struct {
		id   string
		tags []string
		fav  bool
	}{
		{articleTag1, []string{tagTech}, false},
		{articleTag2, []string{tagTech, tagProgramming}, true},
		{articleTag3, []string{tagProgramming}, false},
		{articleTag4, []string{tagTech}, true},
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
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		// Add tags
		for _, tag := range tc.tags {
			err = s.repositories.AddTagsToArticle(ctx, account, tc.id, []string{tag}, nil)
			require.NoError(t, err)
		}
	}

	// Test filtering by tagTech tag
	tag := tagTech
	articles, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, articles, 3)

	for _, article := range articles {
		assert.Contains(t, []string{articleTag1, articleTag2, articleTag4}, article.ID)
	}

	// Test filtering by tagProgramming tag
	tag = tagProgramming
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, articles, 2)

	for _, article := range articles {
		assert.Contains(t, []string{articleTag2, articleTag3}, article.ID)
	}

	// Test filtering by non-existent tag
	tag = tagNonexistent
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, articles)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccount_WithCombinedFilters() {
	ctx := context.Background()
	t := s.T()

	account := "test-account-combined"

	// Create articles with different tag and favorite combinations
	testCases := []struct {
		id   string
		tags []string
		fav  bool
	}{
		{article1, []string{tagTech}, true},                     // tech AND favorite
		{article2ID, []string{tagTech}, false},                  // tech only
		{article3ID, []string{tagProgramming}, true},            // programming AND favorite
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
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		// Add tags
		for _, tag := range tc.tags {
			err = s.repositories.AddTagsToArticle(ctx, account, tc.id, []string{tag}, nil)
			require.NoError(t, err)
		}
	}

	// Test: favorite=true AND tag=tagTech
	tag := tagTech
	favorite := true
	articles, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, articles, 2)

	ids := make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		assert.True(t, a.Favorite)
	}
	assert.Contains(t, ids, article1)
	assert.Contains(t, ids, "article-7")

	// Test: favorite=true AND tag=tagProgramming
	tag = tagProgramming
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, articles, 2)

	ids = make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		assert.True(t, a.Favorite)
	}
	assert.Contains(t, ids, article3ID)
	assert.Contains(t, ids, "article-7")

	// Test: favorite=false AND tag=tagTech
	favorite = false
	tag = tagTech
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, articles, 2)

	ids = make([]string, len(articles))
	for i, a := range articles {
		ids[i] = a.ID
		assert.False(t, a.Favorite)
	}
	assert.Contains(t, ids, article2ID)
	assert.Contains(t, ids, "article-8")

	// Test: favorite=true AND non-existent tag
	tag = "nonexistent"
	favorite = true
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, articles)

	// Test: favorite=false AND non-existent tag
	favorite = false
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, articles)
}

func (s *DynamoDBRepositoryTestSuite) TestGetMetadataByAccount_TagFilterPagination() {
	ctx := context.Background()
	t := s.T()

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
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		err = s.repositories.AddTagsToArticle(ctx, account, article.ID, []string{tagTech}, nil)
		require.NoError(t, err)
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
			Language:           articleTestLang,
			PublishedAt:        &publishedAt,
			WordCount:          100,
			ReadingTimeMinutes: 1,
			SiteName:           testSiteName,
			SourceDomain:       testSourceDomain,
		}

		err := s.repositories.Store(ctx, article)
		require.NoError(t, err)

		err = s.repositories.AddTagsToArticle(ctx, account, article.ID, []string{"other"}, nil)
		require.NoError(t, err)
	}

	// Test pagination with tag filter only
	tag := tagTech
	articles, total, err := s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, articles, 10)

	// Test pagination with tag and favorite filters
	tag = tagTech
	favorite := true
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 1, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 12, total) // Half of 25 are favorites
	assert.Len(t, articles, 10)

	// Test second page with combined filters
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 2, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 12, total)
	assert.Len(t, articles, 2)

	// Test third page (empty result) with combined filters
	articles, total, err = s.repositories.GetMetadataByAccount(
		ctx, account, 3, 10, &internaltypes.ArticleFilter{Favorite: &favorite, Tag: &tag})
	require.NoError(t, err)
	assert.Equal(t, 12, total)
	assert.Empty(t, articles)
}
