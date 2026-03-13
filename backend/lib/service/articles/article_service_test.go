package articles

import (
	"context"
	"testing"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/model"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/content/epub"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	articles []*model.Article
}

func (m *MockRepository) Store(_ context.Context, article *model.Article) error {
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
	favoriteFilter *bool,
) (result []*model.Article, lastEvaluatedKey any, total int, err error) {
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
	for i, article := range m.articles {
		if article.Account == account && article.ID == id {
			m.articles = append(m.articles[:i], m.articles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockRepository) DeleteByAccount(_ context.Context, account string) (int, error) {
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
