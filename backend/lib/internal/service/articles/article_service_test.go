package articles

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/internal/epub"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
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
) (result []*model.Article, lastEvaluatedKey any, total int, err error) {
	if m.metadataErr != nil {
		return nil, nil, 0, m.metadataErr
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
	return result, nil, len(result), nil
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

func TestUpdateArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(nil, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

	ctx := context.Background()

	result, err := svc.GetArticlesMetadata(ctx, "account-123", 1, 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
	svc := New(mockRepo, epub.NewPublisher(), nil)

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
