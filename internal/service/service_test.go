package service

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/internal/config"
	"github.com/shaftoe/savetoink/internal/consts"
	"github.com/shaftoe/savetoink/internal/model"
	"github.com/shaftoe/savetoink/internal/repository"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	articles []*model.Article
}

func (m *MockRepository) Store(_ context.Context, article *model.Article) error {
	m.articles = append(m.articles, article)
	return nil
}

func (m *MockRepository) GetByAccountAndID(_ context.Context, account, id string) (*model.Article, error) {
	for _, article := range m.articles {
		if article.Account == account && article.ID == id {
			return article, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *MockRepository) GetMetadataByAccount(
	_ context.Context,
	account string,
	page, pageSize int,
) (articles []*model.Article, lastEvaluatedKey map[string]types.AttributeValue, total int, err error) {
	var result []*model.Article
	for _, article := range m.articles {
		if article.Account == account {
			articleCopy := *article
			articleCopy.Content = ""
			result = append(result, &articleCopy)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	total = len(result)
	skip := max((page-1)*pageSize, 0)
	if skip >= total {
		return []*model.Article{}, nil, total, nil
	}
	end := min(skip+pageSize, total)

	if end < total {
		lastEvaluatedKey = map[string]types.AttributeValue{
			"account":   &types.AttributeValueMemberS{Value: account},
			"createdAt": &types.AttributeValueMemberS{Value: result[end-1].CreatedAt.Format(time.RFC3339)},
		}
	}

	return result[skip:end], lastEvaluatedKey, total, nil
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
	return initialLen - len(m.articles), nil
}

func TestGetArticlesMetadata(t *testing.T) {
	now := time.Now()
	articles := []*model.Article{
		{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1", CreatedAt: now.Add(-4 * time.Hour)},
		{Account: "user1", ID: "2", Title: "Article 2", URL: "https://example.com/2", CreatedAt: now.Add(-3 * time.Hour)},
		{Account: "user1", ID: "3", Title: "Article 3", URL: "https://example.com/3", CreatedAt: now.Add(-2 * time.Hour)},
		{Account: "user1", ID: "4", Title: "Article 4", URL: "https://example.com/4", CreatedAt: now.Add(-1 * time.Hour)},
		{Account: "user1", ID: "5", Title: "Article 5", URL: "https://example.com/5", CreatedAt: now},
	}

	tests := []struct {
		name            string
		accountID       string
		page            int
		pageSize        int
		expectedCount   int
		expectedPage    int
		expectedTotal   int
		expectedHasMore bool
	}{
		{
			name:            "first page with page size 2",
			accountID:       "user1",
			page:            1,
			pageSize:        2,
			expectedCount:   2,
			expectedPage:    1,
			expectedTotal:   5,
			expectedHasMore: true,
		},
		{
			name:            "second page with page size 2",
			accountID:       "user1",
			page:            2,
			pageSize:        2,
			expectedCount:   2,
			expectedPage:    2,
			expectedTotal:   5,
			expectedHasMore: true,
		},
		{
			name:            "last page with page size 2",
			accountID:       "user1",
			page:            3,
			pageSize:        2,
			expectedCount:   1,
			expectedPage:    3,
			expectedTotal:   5,
			expectedHasMore: false,
		},
		{
			name:            "page beyond last returns empty",
			accountID:       "user1",
			page:            10,
			pageSize:        2,
			expectedCount:   0,
			expectedPage:    10,
			expectedTotal:   5,
			expectedHasMore: false,
		},
		{
			name:            "get all articles in one page",
			accountID:       "user1",
			page:            1,
			pageSize:        100,
			expectedCount:   5,
			expectedPage:    1,
			expectedTotal:   5,
			expectedHasMore: false,
		},
		{
			name:            "no articles for account",
			accountID:       "user2",
			page:            1,
			pageSize:        10,
			expectedCount:   0,
			expectedPage:    1,
			expectedTotal:   0,
			expectedHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{articles: articles}
			svc := &Service{repo: mockRepo}

			result, err := svc.GetArticlesMetadata(context.Background(), tt.accountID, tt.page, tt.pageSize, nil)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Articles) != tt.expectedCount {
				t.Errorf("expected %d articles, got %d", tt.expectedCount, len(result.Articles))
			}

			if result.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, result.Page)
			}

			if result.PageSize != tt.pageSize {
				t.Errorf("expected page_size %d, got %d", tt.pageSize, result.PageSize)
			}

			if result.Total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, result.Total)
			}

			if result.HasMore != tt.expectedHasMore {
				t.Errorf("expected has_more %v, got %v", tt.expectedHasMore, result.HasMore)
			}
		})
	}
}

func TestGetArticlesMetadataWithNilRepo(t *testing.T) {
	svc := &Service{repo: nil}

	result, err := svc.GetArticlesMetadata(context.Background(), "user1", 1, 10, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Articles == nil {
		t.Error("expected articles to be initialized, got nil")
	}

	if len(result.Articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(result.Articles))
	}

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestGetArticle(t *testing.T) {
	article := &model.Article{
		Account:   "user1",
		ID:        "test-id",
		Title:     "Test Article",
		URL:       "https://example.com/test",
		Content:   "<p>Test content</p>",
		CreatedAt: time.Now().UTC(),
	}

	mockRepo := &MockRepository{articles: []*model.Article{article}}
	svc := &Service{repo: mockRepo}

	result, err := svc.GetArticle(context.Background(), "user1", "test-id")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("expected id 'test-id', got '%s'", result.ID)
	}

	if result.Content != "<p>Test content</p>" {
		t.Errorf("expected content to be included, got '%s'", result.Content)
	}
}

func TestGetArticleNotFound(t *testing.T) {
	mockRepo := &MockRepository{articles: []*model.Article{}}
	svc := &Service{repo: mockRepo}

	_, err := svc.GetArticle(context.Background(), "user1", "non-existent")

	if err == nil {
		t.Error("expected error for non-existent article, got nil")
	}
}

func TestGetArticleEmptyID(t *testing.T) {
	svc := &Service{repo: nil}

	_, err := svc.GetArticle(context.Background(), "user1", "")

	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestGetArticlesMetadataWithDeliveryStatus(t *testing.T) {
	now := time.Now()
	articles := []*model.Article{
		{
			Account:        "user1",
			ID:             "1",
			Title:          "Article 1",
			URL:            "https://example.com/1",
			CreatedAt:      now.Add(-1 * time.Hour),
			DeliveryStatus: consts.StatusDelivered,
		},
		{
			Account:        "user1",
			ID:             "2",
			Title:          "Article 2",
			URL:            "https://example.com/2",
			CreatedAt:      now,
			DeliveryStatus: consts.StatusFailed,
			Error:          "email failed",
		},
	}

	mockRepo := &MockRepository{articles: articles}
	svc := &Service{repo: mockRepo}

	result, err := svc.GetArticlesMetadata(context.Background(), "user1", 1, 10, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(result.Articles))
	}

	if result.Articles[0].DeliveryStatus != consts.StatusFailed {
		t.Errorf("expected delivery status %v, got %v", consts.StatusFailed, result.Articles[0].DeliveryStatus)
	}

	if result.Articles[0].Error != "email failed" {
		t.Errorf("expected error 'email failed', got '%s'", result.Articles[0].Error)
	}

	if result.Articles[1].DeliveryStatus != consts.StatusDelivered {
		t.Errorf("expected delivery status %v, got %v", consts.StatusDelivered, result.Articles[1].DeliveryStatus)
	}
}

func TestDeleteArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1"},
		},
	}
	svc := &Service{repo: mockRepo}

	result, err := svc.DeleteArticle(context.Background(), "user1", "1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", result.Deleted)
	}
}

func TestDeleteArticle_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1"},
		},
	}
	svc := &Service{repo: mockRepo}

	result, err := svc.DeleteArticle(context.Background(), "user1", "non-existent")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted for not found, got %d", result.Deleted)
	}
}

func TestDeleteArticle_EmptyID(t *testing.T) {
	svc := &Service{repo: nil}

	_, err := svc.DeleteArticle(context.Background(), "user1", "")

	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestDeleteArticle_NoRepo(t *testing.T) {
	svc := &Service{repo: nil}

	result, err := svc.DeleteArticle(context.Background(), "user1", "1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted with no repo, got %d", result.Deleted)
	}
}

func TestDeleteAllArticles_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1"},
			{Account: "user1", ID: "2", Title: "Article 2", URL: "https://example.com/2"},
			{Account: "user2", ID: "3", Title: "Article 3", URL: "https://example.com/3"},
		},
	}
	svc := &Service{repo: mockRepo}

	result, err := svc.DeleteAllArticles(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", result.Deleted)
	}
}

func TestDeleteAllArticles_NoRepo(t *testing.T) {
	svc := &Service{repo: nil}

	result, err := svc.DeleteAllArticles(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted with no repo, got %d", result.Deleted)
	}
}

func TestSend_NilResult(t *testing.T) {
	cfg := &config.Config{
		SenderEmail:      "sender@example.com",
		MailjetAPIKey:    "api-key",
		MailjetAPISecret: "api-secret",
	}
	svc := New(cfg)

	_, err := svc.Send(context.Background(), nil, "", "test@kindle.com")

	if err == nil {
		t.Error("expected error for nil result, got nil")
	}
}

func TestSend_NilArticle(t *testing.T) {
	cfg := &config.Config{
		SenderEmail:      "sender@example.com",
		MailjetAPIKey:    "api-key",
		MailjetAPISecret: "api-secret",
	}
	svc := New(cfg)

	result := NewProcessResult(nil, []byte("test"), "https://example.com")

	_, err := svc.Send(context.Background(), result, "", "test@kindle.com")

	if err == nil {
		t.Error("expected error for nil article, got nil")
	}
}

func TestSend_NoSenderConfigured(t *testing.T) {
	cfg := &config.Config{}
	svc := New(cfg)

	article := &model.Article{
		Title: "Test Article",
		URL:   "https://example.com",
	}
	result := NewProcessResult(article, []byte("test"), "https://example.com")

	_, err := svc.Send(context.Background(), result, "", "test@kindle.com")

	if err == nil {
		t.Error("expected error when sender not configured, got nil")
	}
}

func TestWriteToFile_NilResult(t *testing.T) {
	svc := &Service{}

	err := svc.WriteToFile(nil, "/tmp/test.epub")

	if err == nil {
		t.Error("expected error for nil result, got nil")
	}
}

func TestWriteToFile_NilArticle(t *testing.T) {
	svc := &Service{}

	result := NewProcessResult(nil, []byte("test"), "https://example.com")

	err := svc.WriteToFile(result, "/tmp/test.epub")

	if err == nil {
		t.Error("expected error for nil article, got nil")
	}
}

func TestWriteToFile_EmptyPath(t *testing.T) {
	svc := &Service{}

	article := &model.Article{
		Title: "Test Article",
	}
	result := NewProcessResult(article, []byte("test"), "https://example.com")

	err := svc.WriteToFile(result, "")

	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestService_AddTags(t *testing.T) {
	svc := &Service{}
	article := &model.Article{
		Account:   "test-account",
		ID:        "test-id",
		Tags:      []string{"existing-tag"},
		CreatedAt: time.Now(),
	}

	svc.repo = &MockRepository{articles: []*model.Article{article}}

	err := svc.AddTags(context.Background(), "test-account", "test-id", []string{"new-tag1", "new-tag2"})
	require.NoError(t, err)

	require.Contains(t, article.Tags, "new-tag1")
	require.Contains(t, article.Tags, "new-tag2")
}

func TestService_AddTags_EmptyTags(t *testing.T) {
	svc := &Service{}
	article := &model.Article{
		Account:   "test-account",
		ID:        "test-id",
		Tags:      []string{"existing-tag"},
		CreatedAt: time.Now(),
	}

	svc.repo = &MockRepository{articles: []*model.Article{article}}

	err := svc.AddTags(context.Background(), "test-account", "test-id", []string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "tags cannot be empty")
}

func TestService_AddTags_ArticleNotFound(t *testing.T) {
	svc := &Service{}

	svc.repo = &MockRepository{articles: []*model.Article{}}

	err := svc.AddTags(context.Background(), "test-account", "test-id", []string{"new-tag"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "article not found")
}

func TestService_RemoveTag(t *testing.T) {
	svc := &Service{}
	article := &model.Article{
		Account:   "test-account",
		ID:        "test-id",
		Tags:      []string{"tag-to-remove", "keep-tag"},
		CreatedAt: time.Now(),
	}

	svc.repo = &MockRepository{articles: []*model.Article{article}}

	err := svc.RemoveTag(context.Background(), "test-account", "test-id", "tag-to-remove")
	require.NoError(t, err)

	require.NotContains(t, article.Tags, "tag-to-remove")
	require.Contains(t, article.Tags, "keep-tag")
}

func TestService_RemoveTag_TagNotFound(t *testing.T) {
	svc := &Service{}
	article := &model.Article{
		Account:   "test-account",
		ID:        "test-id",
		Tags:      []string{"tag-to-remove"},
		CreatedAt: time.Now(),
	}

	svc.repo = &MockRepository{articles: []*model.Article{article}}

	err := svc.RemoveTag(context.Background(), "test-account", "test-id", "not-found-tag")
	require.Error(t, err)
	require.Contains(t, err.Error(), "tag not found on article")
}

func TestService_RemoveTag_ArticleNotFound(t *testing.T) {
	svc := &Service{}

	svc.repo = &MockRepository{articles: []*model.Article{}}

	err := svc.RemoveTag(context.Background(), "test-account", "test-id", "tag")
	require.Error(t, err)
	require.Contains(t, err.Error(), "article not found")
}

func TestService_NormalizeTags(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "normal tags",
			input:    []string{"Golang", "AWS"},
			expected: []string{"golang", "aws"},
		},
		{
			name:     "trim whitespace",
			input:    []string{"  Golang  ", " AWS "},
			expected: []string{"golang", "aws"},
		},
		{
			name:     "truncate long tags",
			input:    []string{strings.Repeat("a", 100)},
			expected: []string{strings.Repeat("a", 50)},
		},
		{
			name:     "remove duplicates",
			input:    []string{"golang", "AWS", "golang"},
			expected: []string{"aws", "golang"},
		},
		{
			name:     "remove empty tags",
			input:    []string{"  ", "", "golang"},
			expected: []string{"golang"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.normalizeTags(tt.input)
			if tt.expected == nil {
				require.Nil(t, result)
				return
			}
			require.ElementsMatch(t, tt.expected, result)
		})
	}
}
