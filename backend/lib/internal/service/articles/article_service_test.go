package articles

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"testing"
	"time"

	apperrors "github.com/shaftoe/savetoink/backend/lib/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/lib/internal/epub"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	articles     []*model.Article
	storeErr     error
	getErr       error
	metadataErr  error
	deleteErr    error
	updateFavErr error
}

func (m *MockRepository) Store(_ context.Context, article *model.Article) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	for i, existing := range m.articles {
		if existing.Account == article.Account && existing.ID == article.ID {
			m.articles[i] = article
			return nil
		}
	}
	m.articles = append(m.articles, article)
	return nil
}

func (m *MockRepository) GetByAccountAndID(_ context.Context, account, id string) (*model.Article, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, article := range m.articles {
		if article.Account == account && article.ID == id {
			return article, nil
		}
	}
	return nil, repoimpl.ErrNotFound
}

func (m *MockRepository) GetMetadataByAccount(
	_ context.Context,
	account string,
	_page, _pageSize int, //nolint:revive // unused parameters required by interface
	filter *types.ArticleFilter,
) (result []*model.Article, total int, err error) {
	if m.metadataErr != nil {
		return nil, 0, m.metadataErr
	}
	var favoriteFilter *bool
	if filter != nil {
		favoriteFilter = filter.Favorite
	}
	for _, article := range m.articles {
		if article.Account == account {
			if favoriteFilter != nil && article.Favorite != *favoriteFilter {
				continue
			}
			result = append(result, article)
		}
	}
	return result, len(result), nil
}

func (m *MockRepository) DeleteByAccountAndID(_ context.Context, account, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, article := range m.articles {
		if article.Account == account && article.ID == id {
			m.articles = append(m.articles[:i], m.articles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockRepository) DeleteByAccount(_ context.Context, account string) (int, error) {
	if m.deleteErr != nil {
		return 0, m.deleteErr
	}
	initialLen := len(m.articles)
	var filtered []*model.Article
	for _, article := range m.articles {
		if article.Account != account {
			filtered = append(filtered, article)
		}
	}
	m.articles = filtered
	return initialLen - len(filtered), nil
}

func (m *MockRepository) UpdateFavorite(_ context.Context, account, id string, favorite bool) error {
	if m.updateFavErr != nil {
		return m.updateFavErr
	}
	for _, article := range m.articles {
		if article.Account == account && article.ID == id {
			article.Favorite = favorite
			return nil
		}
	}
	return nil
}

// MockArticleTagsRepository is a mock implementation of ArticleTagsRepository.
type MockArticleTagsRepository struct {
	tags           map[string][]string // key: "accountID:articleID", value: []string (tags)
	tagToArticles  map[string][]string // key: "accountID:tag", value: []string (article IDs)
	addTagsErr     error
	removeTagsErr  error
	setTagsErr     error
	getTagsErr     error
	getArticlesErr error
	getAllTagsErr  error
	deleteTagsErr  error
}

func NewMockArticleTagsRepository() *MockArticleTagsRepository {
	return &MockArticleTagsRepository{
		tags:          make(map[string][]string),
		tagToArticles: make(map[string][]string),
	}
}

func (m *MockArticleTagsRepository) AddTagsToArticle(
	_ context.Context,
	accountID, articleID string,
	tags []string,
	_ *time.Time,
) error {
	if m.addTagsErr != nil {
		return m.addTagsErr
	}
	key := buildArticleTagKey(accountID, articleID)
	existingTags, exists := m.tags[key]
	if !exists {
		existingTags = []string{}
	}
	// Deduplicate tags
	seen := make(map[string]bool)
	for _, tag := range existingTags {
		seen[tag] = true
	}
	for _, tag := range tags {
		if !seen[tag] {
			existingTags = append(existingTags, tag)
			seen[tag] = true
		}
	}
	m.tags[key] = existingTags
	// Update tagToArticles index
	for _, tag := range tags {
		tagKey := buildAccountTagKey(accountID, tag)
		articleIDs, keyExists := m.tagToArticles[tagKey]
		if !keyExists {
			articleIDs = []string{}
		}
		// Deduplicate article IDs
		found := false
		//nolint:modernize // simple loop for clarity
		for _, id := range articleIDs {
			if id == articleID {
				found = true
				break
			}
		}
		if !found {
			articleIDs = append(articleIDs, articleID)
		}
		m.tagToArticles[tagKey] = articleIDs
	}
	return nil
}

func (m *MockArticleTagsRepository) RemoveTagsFromArticle(
	_ context.Context,
	accountID, articleID string,
	tags []string,
) error {
	if m.removeTagsErr != nil {
		return m.removeTagsErr
	}
	key := buildArticleTagKey(accountID, articleID)
	existingTags, exists := m.tags[key]
	if !exists {
		return nil
	}
	// Remove specified tags
	var newTags []string
	for _, existingTag := range existingTags {
		found := false
		//nolint:modernize // simple loop for clarity
		for _, tagToRemove := range tags {
			if existingTag == tagToRemove {
				found = true
				break
			}
		}
		if !found {
			newTags = append(newTags, existingTag)
		}
	}
	m.tags[key] = newTags
	// Update tagToArticles index
	for _, tag := range tags {
		tagKey := buildAccountTagKey(accountID, tag)
		articleIDs, keyExists := m.tagToArticles[tagKey]
		if keyExists {
			var newArticleIDs []string
			for _, id := range articleIDs {
				if id != articleID {
					newArticleIDs = append(newArticleIDs, id)
				}
			}
			m.tagToArticles[tagKey] = newArticleIDs
		}
	}
	return nil
}

func (m *MockArticleTagsRepository) SetArticleTags(
	_ context.Context,
	accountID, articleID string,
	tags []string,
) error {
	if m.setTagsErr != nil {
		return m.setTagsErr
	}
	key := buildArticleTagKey(accountID, articleID)
	// Get existing tags to remove from index
	existingTags := m.tags[key]
	for _, tag := range existingTags {
		tagKey := buildAccountTagKey(accountID, tag)
		articleIDs, exists := m.tagToArticles[tagKey]
		if exists {
			var newArticleIDs []string
			for _, id := range articleIDs {
				if id != articleID {
					newArticleIDs = append(newArticleIDs, id)
				}
			}
			m.tagToArticles[tagKey] = newArticleIDs
		}
	}
	// Set new tags
	if len(tags) == 0 {
		delete(m.tags, key)
	} else {
		m.tags[key] = tags
	}
	// Update tagToArticles index with new tags
	for _, tag := range tags {
		tagKey := buildAccountTagKey(accountID, tag)
		articleIDs, exists := m.tagToArticles[tagKey]
		if !exists {
			articleIDs = []string{}
		}
		// Deduplicate article IDs
		found := false
		//nolint:modernize // simple loop for clarity
		for _, id := range articleIDs {
			if id == articleID {
				found = true
				break
			}
		}
		if !found {
			articleIDs = append(articleIDs, articleID)
		}
		m.tagToArticles[tagKey] = articleIDs
	}
	return nil
}

func (m *MockArticleTagsRepository) GetArticleTags(_ context.Context, accountID, articleID string) ([]string, error) {
	if m.getTagsErr != nil {
		return nil, m.getTagsErr
	}
	key := buildArticleTagKey(accountID, articleID)
	tags, exists := m.tags[key]
	if !exists {
		return []string{}, nil
	}
	result := make([]string, len(tags))
	copy(result, tags)
	sort.Strings(result)
	return result, nil
}

func (m *MockArticleTagsRepository) GetArticlesByTag(
	_ context.Context,
	accountID, tag string,
	page, pageSize int,
) (articleIDs []string, total int, err error) {
	if m.getArticlesErr != nil {
		return nil, 0, m.getArticlesErr
	}
	tagKey := buildAccountTagKey(accountID, tag)
	articleIDs, exists := m.tagToArticles[tagKey]
	if !exists {
		return []string{}, 0, nil
	}
	// Pagination
	total = len(articleIDs)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []string{}, total, nil
	}
	if end > total {
		end = total
	}
	result := make([]string, end-start)
	copy(result, articleIDs[start:end])
	return result, total, nil
}

func (m *MockArticleTagsRepository) GetAllTagsForAccount(_ context.Context, accountID string) ([]string, error) {
	if m.getAllTagsErr != nil {
		return nil, m.getAllTagsErr
	}
	tagSet := make(map[string]bool)
	for key, tags := range m.tags {
		if len(key) > len(accountID) && key[:len(accountID)] == accountID {
			for _, tag := range tags {
				tagSet[tag] = true
			}
		}
	}
	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result, nil
}

func (m *MockArticleTagsRepository) DeleteTagsForArticle(_ context.Context, accountID, articleID string) error {
	if m.deleteTagsErr != nil {
		return m.deleteTagsErr
	}
	key := buildArticleTagKey(accountID, articleID)
	existingTags := m.tags[key]
	for _, tag := range existingTags {
		tagKey := buildAccountTagKey(accountID, tag)
		articleIDs, exists := m.tagToArticles[tagKey]
		if exists {
			var newArticleIDs []string
			for _, id := range articleIDs {
				if id != articleID {
					newArticleIDs = append(newArticleIDs, id)
				}
			}
			m.tagToArticles[tagKey] = newArticleIDs
		}
	}
	delete(m.tags, key)
	return nil
}

func buildArticleTagKey(accountID, articleID string) string {
	return accountID + ":" + articleID
}

func buildAccountTagKey(accountID, tag string) string {
	return accountID + ":" + tag
}

func TestUpdateArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	initialArticle := &model.Article{
		Account:   "user1",
		ID:        "test-id",
		URL:       "https://example.com/test",
		CreatedAt: time.Now(),
	}

	if err := mockRepo.Store(ctx, initialArticle); err != nil {
		t.Fatalf("failed to store initial article: %v", err)
	}

	updatedArticle := &model.Article{
		Account:   "user1",
		ID:        "test-id",
		URL:       "https://example.com/test",
		Title:     "Updated Title",
		Content:   "<p>Updated content</p>",
		Author:    "Updated Author",
		CreatedAt: time.Now(),
	}

	if err := svc.UpdateArticle(ctx, updatedArticle); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := mockRepo.GetByAccountAndID(ctx, "user1", "test-id")
	if err != nil {
		t.Fatalf("failed to retrieve updated article: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}

	if retrieved.Content != "<p>Updated content</p>" {
		t.Errorf("expected content update, got %s", retrieved.Content)
	}

	if retrieved.Author != "Updated Author" {
		t.Errorf("expected author 'Updated Author', got '%s'", retrieved.Author)
	}
}

func TestUpdateArticle_NilRepo(t *testing.T) {
	svc := New(nil, nil, epub.NewPublisher(), nil)

	article := &model.Article{
		Account: "user1",
		ID:      "test-id",
	}

	assert.Panics(t, func() {
		_ = svc.UpdateArticle(context.Background(), article)
	}, "expected panic with nil repo")
}

func TestUpdateArticle_MissingFields(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	tests := []struct {
		name    string
		article *model.Article
	}{
		{
			name:    "missing account",
			article: &model.Article{ID: "test-id"},
		},
		{
			name:    "missing id",
			article: &model.Article{Account: "user1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := svc.UpdateArticle(ctx, tt.article); err == nil {
				t.Error("expected error for missing required fields")
			}
		})
	}
}

func TestCreateArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	testURL, _ := url.Parse("https://example.com/test-article")
	accountID := "account-123"

	article, err := svc.CreateArticle(ctx, testURL, accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.Account != accountID {
		t.Errorf("expected account '%s', got '%s'", accountID, article.Account)
	}

	if article.ID == "" {
		t.Error("expected article ID to be set")
	}

	if article.URL != testURL.String() {
		t.Errorf("expected URL '%s', got '%s'", testURL.String(), article.URL)
	}

	if article.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateArticle_StoreError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
		storeErr: errors.New("store error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	testURL, _ := url.Parse("https://example.com/test")

	_, err := svc.CreateArticle(ctx, testURL, "account-123")
	if err == nil {
		t.Error("expected error when store fails")
	}
}

func TestGetArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{
				Account:   "account-123",
				ID:        "article-456",
				Title:     "Test Article",
				CreatedAt: time.Now(),
			},
		},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	article, err := svc.GetArticle(ctx, "account-123", "article-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.ID != "article-456" {
		t.Errorf("expected article ID 'article-456', got '%s'", article.ID)
	}

	if article.Title != "Test Article" {
		t.Errorf("expected title 'Test Article', got '%s'", article.Title)
	}
}

func TestGetArticle_EmptyID(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.GetArticle(ctx, "account-123", "")
	if err == nil {
		t.Error("expected error for empty article ID")
	}
}

func TestGetArticle_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.GetArticle(ctx, "account-123", "nonexistent-article")
	if err == nil {
		t.Error("expected not found error")
	}
}

func TestGetArticlesMetadata_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-1", Title: "Article 1", CreatedAt: time.Now()},
			{Account: "account-123", ID: "article-2", Title: "Article 2", CreatedAt: time.Now()},
			{Account: "account-123", ID: "article-3", Title: "Article 3", CreatedAt: time.Now()},
		},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
		return
	}

	if len(result.Articles) != 3 {
		t.Errorf("expected 3 articles, got %d", len(result.Articles))
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.PageSize != 2 {
		t.Errorf("expected page size 2, got %d", result.PageSize)
	}
}

func TestGetArticlesMetadata_WithFavoriteFilter(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-1", Title: "Article 1", Favorite: true, CreatedAt: time.Now()},
			{Account: "account-123", ID: "article-2", Title: "Article 2", Favorite: false, CreatedAt: time.Now()},
			{Account: "account-123", ID: "article-3", Title: "Article 3", Favorite: true, CreatedAt: time.Now()},
		},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	favorite := true
	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 10, &types.ArticleFilter{Favorite: &favorite})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Articles) != 2 {
		t.Errorf("expected 2 favorite articles, got %d", len(result.Articles))
	}

	for _, article := range result.Articles {
		if !article.Favorite {
			t.Errorf("expected all articles to be favorites, found one that is not")
		}
	}
}

func TestGetArticlesMetadata_EmptyResults(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Articles == nil {
		t.Error("expected articles slice to be initialized, got nil")
	}

	if len(result.Articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(result.Articles))
	}

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestDeleteArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Title: "To Delete", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.DeleteArticle(ctx, "account-123", "article-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 1 {
		t.Errorf("expected deleted count 1, got %d", result.Deleted)
	}

	if len(mockRepo.articles) != 0 {
		t.Errorf("expected articles slice to be empty, got %d articles", len(mockRepo.articles))
	}
}

func TestDeleteArticle_EmptyID(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.DeleteArticle(ctx, "account-123", "")
	if err == nil {
		t.Error("expected error for empty article ID")
	}
}

func TestDeleteArticle_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.DeleteArticle(ctx, "account-123", "nonexistent-article")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected deleted count 0, got %d", result.Deleted)
	}
}

func TestToggleFavorite_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Favorite: false, CreatedAt: time.Now()},
		},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	newStatus, err := svc.ToggleFavorite(ctx, "account-123", "article-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !newStatus {
		t.Error("expected favorite status to be toggled to true")
	}

	updated, err := mockRepo.GetByAccountAndID(ctx, "account-123", "article-456")
	if err != nil {
		t.Fatalf("failed to get updated article: %v", err)
	}

	if !updated.Favorite {
		t.Error("expected article favorite status to be true in repo")
	}
}

func TestToggleFavorite_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.ToggleFavorite(ctx, "account-123", "nonexistent-article")
	if err == nil {
		t.Error("expected not found error")
	}
}

func TestToggleFavorite_UpdateError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Favorite: false, CreatedAt: time.Now()},
		},
		updateFavErr: errors.New("update error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.ToggleFavorite(ctx, "account-123", "article-456")
	if err == nil {
		t.Error("expected update error")
	}
}

func TestDeleteArticle_GetError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Title: "To Delete", CreatedAt: time.Now()},
		},
		getErr: errors.New("get error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.DeleteArticle(ctx, "account-123", "article-456")
	if err == nil {
		t.Error("expected get error")
	}
}

func TestDeleteArticle_DeleteError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Title: "To Delete", CreatedAt: time.Now()},
		},
		deleteErr: errors.New("delete error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.DeleteArticle(ctx, "account-123", "article-456")
	if err == nil {
		t.Error("expected delete error")
	}
}

func TestGetArticlesMetadata_Error(t *testing.T) {
	mockRepo := &MockRepository{
		articles:    []*model.Article{},
		metadataErr: errors.New("metadata error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 10, nil)
	if err == nil {
		t.Error("expected metadata error")
	}
}

func TestGetArticle_RepoError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
		getErr:   errors.New("repository error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.GetArticle(ctx, "account-123", "article-456")
	if err == nil {
		t.Error("expected repository error")
	}
}

func TestUpdateArticle_RepoError(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
		storeErr: errors.New("store error"),
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	article := &model.Article{
		Account:   "user1",
		ID:        "test-id",
		URL:       "https://example.com/test",
		Title:     "Updated Title",
		Content:   "<p>Updated content</p>",
		Author:    "Updated Author",
		CreatedAt: time.Now(),
	}

	err := svc.UpdateArticle(ctx, article)
	if err == nil {
		t.Error("expected store error")
	}
}

func TestGetArticlesMetadata_HasMore_FirstPageWithMore(t *testing.T) { //nolint:dupl // intentional duplicate
	articles := make([]*model.Article, 22)
	for i := range 22 {
		articles[i] = &model.Article{
			Account:   "account-123",
			ID:        "article-" + string(rune(i)),
			Title:     "Article " + string(rune(i)),
			CreatedAt: time.Now(),
		}
	}
	mockRepo := &MockRepository{
		articles: articles,
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 20, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 22 {
		t.Errorf("expected total 22, got %d", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.PageSize != 20 {
		t.Errorf("expected page size 20, got %d", result.PageSize)
	}

	if !result.HasMore {
		t.Error("expected has_more to be true for page 1 with 22 total items and page size 20")
	}
}

func TestGetArticlesMetadata_HasMore_SecondPageWithMore(t *testing.T) { //nolint:dupl // intentional duplicate
	articles := make([]*model.Article, 22)
	for i := range 22 {
		articles[i] = &model.Article{
			Account:   "account-123",
			ID:        "article-" + string(rune(i)),
			Title:     "Article " + string(rune(i)),
			CreatedAt: time.Now(),
		}
	}
	mockRepo := &MockRepository{
		articles: articles,
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 2, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 22 {
		t.Errorf("expected total 22, got %d", result.Total)
	}

	if result.Page != 2 {
		t.Errorf("expected page 2, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", result.PageSize)
	}

	if !result.HasMore {
		t.Error("expected has_more to be true for page 2 with 22 total items and page size 10 (there are items on page 3)")
	}
}

func TestGetArticlesMetadata_HasMore_LastPage(t *testing.T) {
	articles := make([]*model.Article, 22)
	for i := range 22 {
		articles[i] = &model.Article{
			Account:   "account-123",
			ID:        "article-" + string(rune(i)),
			Title:     "Article " + string(rune(i)),
			CreatedAt: time.Now(),
		}
	}
	mockRepo := &MockRepository{
		articles: articles,
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 3, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 22 {
		t.Errorf("expected total 22, got %d", result.Total)
	}

	if result.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", result.PageSize)
	}

	if result.HasMore {
		t.Error("expected has_more to be false for page 3 with 22 total items and page size 10 (last page)")
	}
}

func TestGetArticlesMetadata_HasMore_ExactPageBoundary(t *testing.T) {
	articles := make([]*model.Article, 20)
	for i := range 20 {
		articles[i] = &model.Article{
			Account:   "account-123",
			ID:        "article-" + string(rune(i)),
			Title:     "Article " + string(rune(i)),
			CreatedAt: time.Now(),
		}
	}
	mockRepo := &MockRepository{
		articles: articles,
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 20, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 20 {
		t.Errorf("expected total 20, got %d", result.Total)
	}

	if result.HasMore {
		t.Error("expected has_more to be false when page size equals total (exact boundary)")
	}
}

func TestGetArticlesMetadata_HasMore_EmptyResults(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HasMore {
		t.Error("expected has_more to be false for empty results")
	}
}

func TestGetArticlesMetadata_HasMore_WithFavoriteFilter(t *testing.T) {
	articles := make([]*model.Article, 25)
	for i := range 25 {
		articles[i] = &model.Article{
			Account:   "account-123",
			ID:        "article-" + string(rune(i)),
			Title:     "Article " + string(rune(i)),
			CreatedAt: time.Now(),
		}
	}
	for i := range 15 {
		articles[i].Favorite = true
	}
	mockRepo := &MockRepository{
		articles: articles,
	}
	svc := New(mockRepo, nil, epub.NewPublisher(), nil)

	ctx := context.Background()

	favorite := true
	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 10, &types.ArticleFilter{Favorite: &favorite})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 15 {
		t.Errorf("expected total 15 (favorite articles only), got %d", result.Total)
	}

	if !result.HasMore {
		t.Error("expected has_more to be true for page 1 with 15 favorite items and page size 10")
	}
}

// Tag-related tests

func TestAddArticleTags_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags := []string{"tech", "programming"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", tags)
	require.NoError(t, err)

	// Verify tags were added
	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"programming", "tech"}, retrievedTags)
}

func TestAddArticleTags_ArticleNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	err := svc.AddArticleTags(ctx, "account-123", "article-456", []string{"tech"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
}

func TestAddArticleTags_EmptyTags(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	err := svc.AddArticleTags(ctx, "account-123", "article-456", []string{})
	assert.Error(t, err)
}

func TestAddArticleTags_TooManyTags(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags := make([]string, 11)
	for i := range tags {
		tags[i] = "tag" + string(rune('a'+i))
	}

	err := svc.AddArticleTags(ctx, "account-123", "article-456", tags)
	assert.Error(t, err)
}

func TestAddArticleTags_InvalidTagCharacters(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	err := svc.AddArticleTags(ctx, "account-123", "article-456", []string{"tag@with#special"})
	assert.Error(t, err)
}

func TestAddArticleTags_TagTooLong(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	longTag := string(make([]byte, 51)) // 51 characters, exceeds maxTagLength
	err := svc.AddArticleTags(ctx, "account-123", "article-456", []string{longTag})
	assert.Error(t, err)
}

func TestAddArticleTags_DuplicateTags(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags := []string{"tech", "tech", "programming"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", tags)
	require.NoError(t, err)

	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"programming", "tech"}, retrievedTags)
}

func TestAddArticleTags_TagNormalization(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags := []string{"  TECH  ", "Programming", "golang "}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", tags)
	require.NoError(t, err)

	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"golang", "programming", "tech"}, retrievedTags)
}

func TestRemoveArticleTags_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	// First add some tags
	initialTags := []string{"tech", "programming", "golang"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", initialTags)
	require.NoError(t, err)

	// Remove one tag
	err = svc.RemoveArticleTags(ctx, "account-123", "article-456", []string{"programming"})
	require.NoError(t, err)

	// Verify tag was removed
	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"golang", "tech"}, retrievedTags)
}

func TestRemoveArticleTags_ArticleNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	err := svc.RemoveArticleTags(ctx, "account-123", "article-456", []string{"tech"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
}

func TestSetArticleTags_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	// First add some tags
	initialTags := []string{"tech", "programming"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", initialTags)
	require.NoError(t, err)

	// Replace all tags
	newTags := []string{"golang", "rust"}
	err = svc.SetArticleTags(ctx, "account-123", "article-456", newTags)
	require.NoError(t, err)

	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"golang", "rust"}, retrievedTags)
}

func TestSetArticleTags_ClearAllTags(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	// First add some tags
	initialTags := []string{"tech", "programming"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", initialTags)
	require.NoError(t, err)

	// Clear all tags
	err = svc.SetArticleTags(ctx, "account-123", "article-456", []string{})
	require.NoError(t, err)

	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{}, retrievedTags)
}

func TestSetArticleTags_ArticleNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	err := svc.SetArticleTags(ctx, "account-123", "article-456", []string{"tech"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
}

func TestGetArticleTags_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags := []string{"tech", "programming", "golang"}
	err := svc.AddArticleTags(ctx, "account-123", "article-456", tags)
	require.NoError(t, err)

	retrievedTags, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{"golang", "programming", "tech"}, retrievedTags)
}

func TestGetArticleTags_ArticleNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	_, err := svc.GetArticleTags(ctx, "account-123", "article-456")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
}

func TestGetAllTagsForAccount_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-1", CreatedAt: time.Now()},
			{Account: "account-123", ID: "article-2", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	_ = mockTagsRepo.AddTagsToArticle(ctx, "account-123", "article-1", []string{"tech", "programming"}, nil)
	_ = mockTagsRepo.AddTagsToArticle(ctx, "account-123", "article-2", []string{"tech", "golang"}, nil)

	tags, err := svc.GetAllTagsForAccount(ctx, "account-123")
	require.NoError(t, err)
	assert.Equal(t, []string{"golang", "programming", "tech"}, tags)
}

func TestGetAllTagsForAccount_Empty(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	tags, err := svc.GetAllTagsForAccount(ctx, "account-123")
	require.NoError(t, err)
	assert.Equal(t, []string{}, tags)
}

func TestDeleteArticle_CascadesToTags(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Title: "To Delete", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	// Add tags to the article
	_ = mockTagsRepo.AddTagsToArticle(ctx, "account-123", "article-456", []string{"tech", "programming"}, nil)

	// Delete the article
	result, err := svc.DeleteArticle(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, 1, result.Deleted)

	// Verify tags were deleted
	retrievedTags, err := mockTagsRepo.GetArticleTags(ctx, "account-123", "article-456")
	require.NoError(t, err)
	assert.Equal(t, []string{}, retrievedTags)
}

func TestDeleteArticle_TagsErrorFailsDelete(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "account-123", ID: "article-456", Title: "To Delete", CreatedAt: time.Now()},
		},
	}
	mockTagsRepo := NewMockArticleTagsRepository()
	mockTagsRepo.deleteTagsErr = errors.New("tags delete error")
	svc := New(mockRepo, mockTagsRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	// Add tags to the article
	_ = mockTagsRepo.AddTagsToArticle(ctx, "account-123", "article-456", []string{"tech"}, nil)

	// Delete the article - should fail if tag deletion fails
	_, err := svc.DeleteArticle(ctx, "account-123", "article-456")
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to delete tags for article")
}
