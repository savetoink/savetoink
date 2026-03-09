package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/epub"
)

const (
	testUser1         = "user1"
	testDeviceEmail   = "device@kindle.com"
	testNewKindle     = "new@kindle.com"
	testUserEmail     = "user@example.com"
	testStatusSuccess = "success"
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
	page, pageSize int,
	favoriteFilter *bool,
) (articles []*model.Article, lastEvaluatedKey map[string]awstypes.AttributeValue, total int, err error) {
	var result []*model.Article
	for _, article := range m.articles {
		if article.Account == account {
			if favoriteFilter != nil && article.Favorite != *favoriteFilter {
				continue
			}
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
		lastEvaluatedKey = map[string]awstypes.AttributeValue{
			"account":   &awstypes.AttributeValueMemberS{Value: account},
			"createdAt": &awstypes.AttributeValueMemberS{Value: result[end-1].CreatedAt.Format(time.RFC3339)},
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

func (m *MockRepository) UpdateFavorite(_ context.Context, account, id string, favorite bool) error {
	for _, article := range m.articles {
		if article.Account == account && article.ID == id {
			article.Favorite = favorite
			return nil
		}
	}
	return repoimpl.ErrNotFound
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
			svc := New(&Dependencies{ArticlesRepo: mockRepo})

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
	svc := New(&Dependencies{})

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

func TestGetArticlesMetadataWithFavoriteFilter(t *testing.T) {
	now := time.Now()
	favoriteTrue := true
	favoriteFalse := false
	articles := []*model.Article{
		{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1", Favorite: true, CreatedAt: now},
		{Account: "user1", ID: "2", Title: "Article 2", URL: "https://example.com/2", Favorite: false, CreatedAt: now},
	}

	mockRepo := &MockRepository{articles: articles}
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

	t.Run("filter favorite true", func(t *testing.T) {
		result, err := svc.GetArticlesMetadata(context.Background(), "user1", 1, 10, &favoriteTrue)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Articles) != 1 {
			t.Errorf("expected 1 favorite article, got %d", len(result.Articles))
		}

		if len(result.Articles) > 0 && !result.Articles[0].Favorite {
			t.Error("expected article to be marked as favorite")
		}
	})

	t.Run("filter favorite false", func(t *testing.T) {
		result, err := svc.GetArticlesMetadata(context.Background(), "user1", 1, 10, &favoriteFalse)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Articles) != 1 {
			t.Errorf("expected 1 non-favorite article, got %d", len(result.Articles))
		}

		if len(result.Articles) > 0 && result.Articles[0].Favorite {
			t.Error("expected article not to be marked as favorite")
		}
	})
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
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

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
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

	_, err := svc.GetArticle(context.Background(), "user1", "non-existent")

	if err == nil {
		t.Error("expected error for non-existent article, got nil")
	}
}

func TestGetArticleEmptyID(t *testing.T) {
	svc := New(&Dependencies{})

	_, err := svc.GetArticle(context.Background(), "user1", "")

	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestDeleteArticle_Success(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1"},
		},
	}
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

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
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

	result, err := svc.DeleteArticle(context.Background(), "user1", "non-existent")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted for not found, got %d", result.Deleted)
	}
}

func TestDeleteArticle_EmptyID(t *testing.T) {
	svc := New(&Dependencies{})

	_, err := svc.DeleteArticle(context.Background(), "user1", "")

	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestDeleteArticle_NoRepo(t *testing.T) {
	svc := New(&Dependencies{})

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
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

	result, err := svc.DeleteAllArticles(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", result.Deleted)
	}
}

func TestDeleteAllArticles_NoRepo(t *testing.T) {
	svc := New(&Dependencies{})

	result, err := svc.DeleteAllArticles(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted with no repo, got %d", result.Deleted)
	}
}

func TestDeleteAllArticles_NoArticles(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user2", ID: "3", Title: "Article 3", URL: "https://example.com/3"},
		},
	}
	svc := New(&Dependencies{ArticlesRepo: mockRepo})

	result, err := svc.DeleteAllArticles(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted for account with no articles, got %d", result.Deleted)
	}
}

type MockSender struct {
	sendFunc func(ctx context.Context, req *email.Request) (*email.SendEmailResponse, error)
}

func (m *MockSender) SendEmail(ctx context.Context, req *email.Request) (*email.SendEmailResponse, error) {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, req)
	}
	return &email.SendEmailResponse{Status: "success", MessageID: "msg-123"}, nil
}

type MockUserProfileRepository struct {
	profiles map[string]*model.UserProfile
}

func NewMockUserProfileRepository() *MockUserProfileRepository {
	return &MockUserProfileRepository{
		profiles: make(map[string]*model.UserProfile),
	}
}

func (m *MockUserProfileRepository) GetUserProfile(_ context.Context, account string) (*model.UserProfile, error) {
	profile, exists := m.profiles[account]
	if !exists {
		return nil, nil
	}
	return profile, nil
}

func (m *MockUserProfileRepository) GetAccountIDByDeviceEmail(_ context.Context, deviceEmail string) (string, error) {
	for account, profile := range m.profiles {
		if profile.DeviceEmail == deviceEmail {
			return account, nil
		}
	}
	return "", repoimpl.ErrNotFound
}

func (m *MockUserProfileRepository) PutUserProfile(_ context.Context, profile *model.UserProfile) error {
	m.profiles[profile.Account] = profile
	return nil
}

func (m *MockUserProfileRepository) DeleteUserProfile(_ context.Context, account string) error {
	delete(m.profiles, account)
	return nil
}

func (m *MockUserProfileRepository) DeleteUserDeviceEmail(_ context.Context, account string) error {
	if profile, exists := m.profiles[account]; exists {
		profile.DeviceEmail = ""
	}
	return nil
}

type MockSendsRepository struct {
	sends []*model.Send
}

func (m *MockSendsRepository) CreateSendRecord(_ context.Context, send *model.Send) error {
	newSend := *send
	newSend.SentAt = time.Now().UTC()
	newSend.Status = "pending"
	m.sends = append(m.sends, &newSend)
	return nil
}

func (m *MockSendsRepository) UpdateSendRecord(_ context.Context, send *model.Send) error {
	for _, s := range m.sends {
		if s.Account == send.Account && s.ArticleID == send.ArticleID {
			s.Status = send.Status
			s.MessageID = send.MessageID
			s.ErrorResponse = send.ErrorResponse
			return nil
		}
	}
	return nil
}

func (m *MockSendsRepository) GetSendsByArticleID(_ context.Context, articleID string) ([]*model.Send, error) {
	var result []*model.Send
	for _, send := range m.sends {
		if send.ArticleID == articleID {
			result = append(result, send)
		}
	}
	return result, nil
}

func (m *MockSendsRepository) GetSendsByAccountDateRange(
	_ context.Context,
	account string,
	startDate, endDate time.Time,
) ([]*model.Send, error) {
	var result []*model.Send
	for _, send := range m.sends {
		if send.Account == account && send.SentAt.After(startDate) && send.SentAt.Before(endDate) {
			result = append(result, send)
		}
	}
	return result, nil
}

func (m *MockSendsRepository) CountSendsByAccountDateRange(
	_ context.Context,
	account string,
	startDate, endDate time.Time,
) (int, error) {
	count := 0
	for _, send := range m.sends {
		if send.Account == account && send.SentAt.After(startDate) && send.SentAt.Before(endDate) {
			count++
		}
	}
	return count, nil
}

func TestNew(t *testing.T) {
	mockSender := &MockSender{}
	mockRepo := &MockRepository{}
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          mockSender,
		ArticlesRepo:    mockRepo,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	if svc == nil {
		t.Fatal("expected service to be created, got nil")
	}
}

func TestNewDependenciesFromConfig(t *testing.T) {
	cfg := &config.Config{
		BrowserlessKey:   "test-key",
		EmailProvider:    "mailjet",
		MailjetAPIKey:    "mailjet-key",
		MailjetAPISecret: "mailjet-secret",
		SenderEmail:      "sender@example.com",
		AppURL:           "https://app.example.com",
		ArticlesTable:    "articles",
		UserProfileTable: "profiles",
		SendsTable:       "sends",
	}

	deps := NewDependenciesFromConfig(cfg)

	if deps.Fetcher == nil {
		t.Error("expected Fetcher to be initialized")
	}

	if deps.Extractor == nil {
		t.Error("expected Extractor to be initialized")
	}

	if deps.Publisher == nil {
		t.Error("expected Publisher to be initialized")
	}

	if deps.Sender == nil {
		t.Error("expected Sender to be initialized")
	}

	if deps.ArticlesRepo != nil {
		t.Error("expected ArticlesRepo to be nil when AWSConfig is not set")
	}

	if deps.UserProfileRepo != nil {
		t.Error("expected UserProfileRepo to be nil when AWSConfig is not set")
	}

	if deps.SendsRepo != nil {
		t.Error("expected SendsRepo to be nil when AWSConfig is not set")
	}

	if deps.Config == nil {
		t.Error("expected Config to be initialized")
	}
}

func TestNewFromConfig(t *testing.T) {
	cfg := &config.Config{
		BrowserlessKey:   "test-key",
		EmailProvider:    "mailjet",
		MailjetAPIKey:    "mailjet-key",
		MailjetAPISecret: "mailjet-secret",
		SenderEmail:      "sender@example.com",
		AppURL:           "https://app.example.com",
		ArticlesTable:    "articles",
		UserProfileTable: "profiles",
		SendsTable:       "sends",
	}

	svc := NewFromConfig(cfg)

	if svc == nil {
		t.Fatal("expected service to be created, got nil")
	}
}

func TestFetch(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html><body>Test Content</body></html>"))
		}))
		defer server.Close()

		fetcher := content.NewFetcher("")
		deps := &Dependencies{Fetcher: fetcher}
		svc := New(deps)

		htmlBytes, fetchType, err := svc.Fetch(context.Background(), server.URL)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(htmlBytes) == 0 {
			t.Error("expected HTML content to be returned")
		}

		if fetchType != content.FetcherTypeGo {
			t.Errorf("expected fetch type %v, got %v", content.FetcherTypeGo, fetchType)
		}
	})

	t.Run("fetch error", func(t *testing.T) {
		fetcher := content.NewFetcher("")
		deps := &Dependencies{Fetcher: fetcher}
		svc := New(deps)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, err := svc.Fetch(ctx, "https://example.com")

		if err == nil {
			t.Error("expected error for canceled context")
		}
	})
}

func TestExtract(t *testing.T) {
	extractor := content.NewExtractor()
	deps := &Dependencies{Extractor: extractor}
	svc := New(deps)

	html := []byte(`<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
</head>
<body>
	<h1>Test Article</h1>
	<p>This is test content for the article.</p>
</body>
</html>`)

	article, err := svc.Extract(context.Background(), html)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if article == nil {
		t.Fatal("expected article to be returned")
	}

	if article.Title == "" {
		t.Error("expected title to be set")
	}

	if article.Content == "" {
		t.Error("expected content to be set")
	}
}

func TestGenerateEPUB(t *testing.T) {
	publisher := epub.NewPublisher()
	deps := &Dependencies{Publisher: publisher}
	svc := New(deps)

	article := &model.Article{
		ID:      "test-id",
		Title:   "Test Article",
		Content: "<p>Test content for the EPUB generation.</p>",
	}

	epubData, err := svc.GenerateEPUB(article)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(epubData) == 0 {
		t.Error("expected epub data to be returned")
	}
}

func TestSendArticle(t *testing.T) {
	tests := []struct {
		name      string
		destEmail string
		epubData  []byte
		title     string
		sendFunc  func(context.Context, *email.Request) (*email.SendEmailResponse, error)
		expectErr bool
	}{
		{
			name:      "successful send",
			destEmail: "test@kindle.com",
			epubData:  []byte("epub data"),
			title:     "Test Article",
			sendFunc: func(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
				return &email.SendEmailResponse{Status: "success", MessageID: "msg-123"}, nil
			},
			expectErr: false,
		},
		{
			name:      "send error",
			destEmail: "test@kindle.com",
			epubData:  []byte("epub data"),
			title:     "Test Article",
			sendFunc: func(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
				return nil, context.Canceled
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSender := &MockSender{
				sendFunc: tt.sendFunc,
			}

			deps := &Dependencies{
				Sender: mockSender,
			}

			svc := New(deps)

			resp, err := svc.SendArticle(context.Background(), tt.destEmail, tt.epubData, tt.title)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectErr && resp == nil {
				t.Error("expected response to be returned")
			}

			if !tt.expectErr && resp != nil && resp.Status != testStatusSuccess {
				t.Errorf("expected status 'success', got %q", resp.Status)
			}
		})
	}
}

func TestSendArticleByID(t *testing.T) {
	tests := []struct {
		name             string
		setupRepo        func() *MockRepository
		setupProfile     func() *MockUserProfileRepository
		accountID        string
		articleID        string
		sendFunc         func(ctx context.Context, req *email.Request) (*email.SendEmailResponse, error)
		expectErr        bool
		errContains      string
		expectResult     bool
		expectSendRecord bool
		expectSendStatus string
	}{
		{
			name: "successful send",
			setupRepo: func() *MockRepository {
				repo := &MockRepository{}
				repo.articles = []*model.Article{
					{
						ID:      "article-123",
						Account: testUser1,
						Title:   "Test Article",
						Content: "Test content",
						URL:     "https://example.com/article",
					},
				}
				return repo
			},
			setupProfile: func() *MockUserProfileRepository {
				profileRepo := NewMockUserProfileRepository()
				profileRepo.profiles[testUser1] = &model.UserProfile{
					Account:     testUser1,
					DeviceEmail: testDeviceEmail,
				}
				return profileRepo
			},
			accountID: testUser1,
			articleID: "article-123",
			sendFunc: func(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
				return &email.SendEmailResponse{Status: "success", MessageID: "msg-123"}, nil
			},
			expectErr:        false,
			expectResult:     true,
			expectSendRecord: true,
			expectSendStatus: "success",
		},
		{
			name:      "article not found",
			setupRepo: func() *MockRepository { return &MockRepository{} },
			setupProfile: func() *MockUserProfileRepository {
				profileRepo := NewMockUserProfileRepository()
				profileRepo.profiles[testUser1] = &model.UserProfile{Account: testUser1, DeviceEmail: testDeviceEmail}
				return profileRepo
			},
			accountID:        testUser1,
			articleID:        "nonexistent",
			expectErr:        true,
			errContains:      "not found",
			expectSendRecord: false,
		},
		{
			name: "device email not configured",
			setupRepo: func() *MockRepository {
				repo := &MockRepository{}
				repo.articles = []*model.Article{{
					ID:      "article-123",
					Account: testUser1,
					Title:   "Test Article",
					Content: "Test content",
					URL:     "https://example.com/article",
				}}
				return repo
			},
			setupProfile: func() *MockUserProfileRepository {
				profileRepo := NewMockUserProfileRepository()
				profileRepo.profiles[testUser1] = &model.UserProfile{Account: testUser1, DeviceEmail: ""}
				return profileRepo
			},
			accountID:        testUser1,
			articleID:        "article-123",
			expectErr:        true,
			errContains:      "device email not configured",
			expectSendRecord: false,
		},
		{
			name: "send email error",
			setupRepo: func() *MockRepository {
				repo := &MockRepository{}
				repo.articles = []*model.Article{{
					ID:      "article-123",
					Account: testUser1,
					Title:   "Test Article",
					Content: "Test content",
					URL:     "https://example.com/article",
				}}
				return repo
			},
			setupProfile: func() *MockUserProfileRepository {
				profileRepo := NewMockUserProfileRepository()
				profileRepo.profiles[testUser1] = &model.UserProfile{Account: testUser1, DeviceEmail: testDeviceEmail}
				return profileRepo
			},
			accountID: testUser1,
			articleID: "article-123",
			sendFunc: func(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
				return nil, context.Canceled
			},
			expectErr:        true,
			expectSendRecord: true,
			expectSendStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSender := &MockSender{sendFunc: tt.sendFunc}
			mockSendsRepo := &MockSendsRepository{}

			deps := &Dependencies{
				Sender:          mockSender,
				ArticlesRepo:    tt.setupRepo(),
				UserProfileRepo: tt.setupProfile(),
				Extractor:       content.NewExtractor(),
				Publisher:       epub.NewPublisher(),
				SendsRepo:       mockSendsRepo,
				Config:          &config.Config{SenderEmail: "sender@example.com", EmailProvider: consts.EmailBackendMailjet},
			}

			svc := New(deps)
			result, err := svc.SendArticleByID(context.Background(), tt.accountID, tt.articleID)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.errContains != "" && err != nil && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
			if tt.expectResult && result == nil {
				t.Error("expected result to be returned")
			}
			if !tt.expectResult && result != nil {
				t.Error("expected no result but got one")
			}
			if result != nil {
				if result.Article == nil {
					t.Error("expected article to be in result")
				}
				if result.DeviceEmail != testDeviceEmail {
					t.Errorf("expected device email %q, got %q", testDeviceEmail, result.DeviceEmail)
				}
				if result.EmailResp == nil {
					t.Error("expected email response to be in result")
				}
			}
			verifySendRecord(t, mockSendsRepo, tt.expectSendRecord, tt.expectSendStatus)
		})
	}
}

func verifySendRecord(t *testing.T, repo *MockSendsRepository, expectRecord bool, expectStatus string) {
	t.Helper()
	if expectRecord {
		if len(repo.sends) != 1 {
			t.Errorf("expected 1 send record, got %d", len(repo.sends))
			return
		}
		if repo.sends[0].Status != expectStatus {
			t.Errorf("expected send status %q, got %q", expectStatus, repo.sends[0].Status)
		}
		if expectStatus == "success" && repo.sends[0].MessageID != "msg-123" {
			t.Errorf("expected message ID %q, got %q", "msg-123", repo.sends[0].MessageID)
		}
		return
	}
	if len(repo.sends) != 0 {
		t.Errorf("expected 0 send records, got %d", len(repo.sends))
	}
}

func TestCreateArticle(t *testing.T) {
	mockRepo := &MockRepository{}
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		ArticlesRepo:    mockRepo,
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	article, err := svc.CreateArticle(context.Background(), "https://example.com/article", "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article == nil {
		t.Fatal("expected article to be returned")
	}

	if article.Account != testUser1 {
		t.Errorf("expected account 'user1', got %q", article.Account)
	}

	if article.ID == "" {
		t.Error("expected article ID to be set")
	}

	if article.URL == "" {
		t.Error("expected URL to be set")
	}

	if article.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateArticleInvalidURL(t *testing.T) {
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	_, err := svc.CreateArticle(context.Background(), "invalid-url", "user1")

	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestUpdateArticle(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "test-id", URL: "https://example.com/test"},
		},
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		ArticlesRepo:    mockRepo,
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	updatedArticle := &model.Article{
		Account:   "user1",
		ID:        "test-id",
		URL:       "https://example.com/test",
		Title:     "Updated Title",
		Content:   "<p>Updated content</p>",
		CreatedAt: time.Now().UTC(),
	}

	err := svc.UpdateArticle(context.Background(), updatedArticle)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := mockRepo.GetByAccountAndID(context.Background(), "user1", "test-id")
	if err != nil {
		t.Fatalf("failed to retrieve updated article: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", retrieved.Title)
	}
}

func TestUpdateArticleNilRepo(t *testing.T) {
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	article := &model.Article{
		Account: "user1",
		ID:      "test-id",
	}

	err := svc.UpdateArticle(context.Background(), article)

	if err != nil {
		t.Errorf("unexpected error with nil repo: %v", err)
	}
}

func TestToggleFavorite(t *testing.T) {
	now := time.Now()
	mockRepo := &MockRepository{
		articles: []*model.Article{
			{Account: "user1", ID: "1", Title: "Article 1", URL: "https://example.com/1", Favorite: false, CreatedAt: now},
		},
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		ArticlesRepo:    mockRepo,
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	t.Run("toggle to true", func(t *testing.T) {
		result, err := svc.ToggleFavorite(context.Background(), "user1", "1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result {
			t.Error("expected favorite to be true, got false")
		}

		article, _ := mockRepo.GetByAccountAndID(context.Background(), "user1", "1")
		if !article.Favorite {
			t.Error("expected article favorite to be true")
		}
	})

	t.Run("toggle to false", func(t *testing.T) {
		article, _ := mockRepo.GetByAccountAndID(context.Background(), "user1", "1")
		article.Favorite = true

		result, err := svc.ToggleFavorite(context.Background(), "user1", "1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result {
			t.Error("expected favorite to be false, got true")
		}
	})
}

func TestToggleFavoriteNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		articles: []*model.Article{},
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		ArticlesRepo:    mockRepo,
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	_, err := svc.ToggleFavorite(context.Background(), "user1", "non-existent")

	if err == nil {
		t.Error("expected error for non-existent article, got nil")
	}
}

func TestToggleFavoriteNilRepo(t *testing.T) {
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()
	mockSendsRepo := &MockSendsRepository{}

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	_, err := svc.ToggleFavorite(context.Background(), "user1", "1")

	if err == nil {
		t.Error("expected error with nil repo, got nil")
	}
}

func TestCountSendsByAccountDateRange(t *testing.T) {
	now := time.Now()
	mockSendsRepo := &MockSendsRepository{
		sends: []*model.Send{
			{Account: "user1", ArticleID: "1", Title: "Article 1", SentAt: now.Add(-1 * time.Hour)},
			{Account: "user1", ArticleID: "2", Title: "Article 2", SentAt: now.Add(-2 * time.Hour)},
			{Account: "user2", ArticleID: "3", Title: "Article 3", SentAt: now.Add(-30 * time.Minute)},
		},
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
		SendsRepo:       mockSendsRepo,
	}

	svc := New(deps)

	count, err := svc.CountSendsByAccountDateRange(
		context.Background(),
		"user1",
		now.Add(-3*time.Hour),
		now,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count of 2, got %d", count)
	}
}

func TestCountSendsByAccountDateRangeNilRepo(t *testing.T) {
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()
	mockProfileRepo := NewMockUserProfileRepository()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	count, err := svc.CountSendsByAccountDateRange(
		context.Background(),
		"user1",
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count of 0 with nil repo, got %d", count)
	}
}

func TestGetUserDeviceEmail(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
		AutoSend:    true,
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	deviceEmail, autoSend, err := svc.GetUserDeviceEmail(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deviceEmail != testDeviceEmail {
		t.Errorf("expected device email 'device@kindle.com', got %q", deviceEmail)
	}

	if !autoSend {
		t.Error("expected auto send to be true")
	}
}

func TestGetUserDeviceEmailNotFound(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	deviceEmail, autoSend, err := svc.GetUserDeviceEmail(context.Background(), "non-existent")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deviceEmail != "" {
		t.Errorf("expected empty device email for non-existent user, got %q", deviceEmail)
	}

	if autoSend {
		t.Error("expected auto send to be false for non-existent user")
	}
}

func TestSetUserDeviceEmail(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.SetUserDeviceEmail(context.Background(), "user1", "new@kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.DeviceEmail != testNewKindle {
		t.Errorf("expected device email 'new@kindle.com', got %q", profile.DeviceEmail)
	}
}

func TestSetUserDeviceEmailUpdateExisting(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "old@kindle.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.SetUserDeviceEmail(context.Background(), "user1", "new@kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.DeviceEmail != testNewKindle {
		t.Errorf("expected device email 'new@kindle.com', got %q", profile.DeviceEmail)
	}

	if profile.Email != testUserEmail {
		t.Error("expected existing email to be preserved")
	}
}

func TestSetUserDeviceEmailWithAutoSend(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), "user1", "device@kindle.com", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.DeviceEmail != "device@kindle.com" {
		t.Errorf("expected device email 'device@kindle.com', got %q", profile.DeviceEmail)
	}

	if !profile.AutoSend {
		t.Error("expected auto send to be true")
	}
}

func TestDeleteUserDeviceEmail(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.DeleteUserDeviceEmail(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.DeviceEmail != "" {
		t.Errorf("expected empty device email, got %q", profile.DeviceEmail)
	}
}

func TestGetUserProfile(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
		AutoSend:    true,
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	profile, err := svc.GetUserProfile(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Account != testUser1 {
		t.Errorf("expected account 'user1', got %q", profile.Account)
	}

	if profile.Email != testUserEmail {
		t.Errorf("expected email 'user@example.com', got %q", profile.Email)
	}
}

func TestSetUserEmail(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.SetUserEmail(context.Background(), "user1", "newemail@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.Email != "newemail@example.com" {
		t.Errorf("expected email 'newemail@example.com', got %q", profile.Email)
	}
}

func TestDeleteUserProfile(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account: "user1",
		Email:   "user@example.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.DeleteUserProfile(context.Background(), "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, err := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile != nil {
		t.Error("expected profile to be deleted")
	}
	if err != nil {
		t.Errorf("expected no error when getting deleted profile, got: %v", err)
	}
}

func TestHandleBounce(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	err := svc.HandleBounce(context.Background(), "device@kindle.com", "bounce error")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, _ := mockProfileRepo.GetUserProfile(context.Background(), "user1")
	if profile.BouncedEmails == nil {
		t.Error("expected bounced emails to be set")
	}

	if _, exists := profile.BouncedEmails["device@kindle.com"]; !exists {
		t.Error("expected device email to be in bounced emails")
	}
}

func TestIsEmailBouncing(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
		BouncedEmails: map[string]model.BounceInfo{
			"device@kindle.com": {
				Timestamp: time.Now().UTC(),
				Error:     "bounce error",
			},
		},
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	t.Run("email is bouncing", func(t *testing.T) {
		bouncing, err := svc.IsEmailBouncing(context.Background(), "user1", "device@kindle.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !bouncing {
			t.Error("expected email to be bouncing")
		}
	})

	t.Run("email not bouncing", func(t *testing.T) {
		bouncing, err := svc.IsEmailBouncing(context.Background(), "user1", "other@kindle.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bouncing {
			t.Error("expected email not to be bouncing")
		}
	})
}

func TestIsEmailBouncingNoProfile(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	bouncing, err := svc.IsEmailBouncing(context.Background(), "user1", "device@kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bouncing {
		t.Error("expected email not to be bouncing for non-existent user")
	}
}

func TestIsEmailBouncingWithNilBouncedEmails(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	bouncing, err := svc.IsEmailBouncing(context.Background(), "user1", "device@kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bouncing {
		t.Error("expected email not to be bouncing when BouncedEmails is nil")
	}
}

func TestGetAccountIDByDeviceEmail(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockProfileRepo.profiles["user1"] = &model.UserProfile{
		Account:     "user1",
		Email:       "user@example.com",
		DeviceEmail: "device@kindle.com",
	}

	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	accountID, err := svc.GetAccountIDByDeviceEmail(context.Background(), "device@kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "user1" {
		t.Errorf("expected account ID 'user1', got %q", accountID)
	}
}

func TestGetAccountIDByDeviceEmailNotFound(t *testing.T) {
	mockProfileRepo := NewMockUserProfileRepository()
	mockExtractor := content.NewExtractor()
	mockPublisher := epub.NewPublisher()

	deps := &Dependencies{
		Extractor:       mockExtractor,
		Publisher:       mockPublisher,
		UserProfileRepo: mockProfileRepo,
	}

	svc := New(deps)

	_, err := svc.GetAccountIDByDeviceEmail(context.Background(), "nonexistent@kindle.com")

	if err == nil {
		t.Error("expected error for non-existent device email")
	}
}
