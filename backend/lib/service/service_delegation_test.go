package service

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/internal/email"
	"github.com/shaftoe/savetoink/backend/lib/internal/epub"
	"github.com/shaftoe/savetoink/backend/lib/internal/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/internal/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const (
	testAccountID = "account-1"
	testEmail     = "device@example.com"
	testMessageID = "msg-123"
)

type testArticlesRepo struct {
	articles     []*model.Article
	storeErr     error
	getErr       error
	metadataErr  error
	deleteErr    error
	updateFavErr error
}

func (r *testArticlesRepo) Store(_ context.Context, article *model.Article) error {
	if r.storeErr != nil {
		return r.storeErr
	}
	r.articles = append(r.articles, article)
	return nil
}

func (r *testArticlesRepo) GetByAccountAndID(_ context.Context, account, id string) (*model.Article, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, a := range r.articles {
		if a.Account == account && a.ID == id {
			return a, nil
		}
	}
	return nil, repoimpl.ErrNotFound
}

func (r *testArticlesRepo) GetMetadataByAccount(
	_ context.Context,
	account string,
	_, _ int,
	favoriteFilter *bool,
) (articles []*model.Article, lastKey any, total int, err error) {
	if r.metadataErr != nil {
		return nil, nil, 0, r.metadataErr
	}
	var result []*model.Article
	for _, a := range r.articles {
		if a.Account == account {
			if favoriteFilter != nil && a.Favorite != *favoriteFilter {
				continue
			}
			result = append(result, &model.Article{
				ID:        a.ID,
				Title:     a.Title,
				URL:       a.URL,
				Account:   a.Account,
				CreatedAt: a.CreatedAt,
			})
		}
	}
	return result, nil, len(result), nil
}

func (r *testArticlesRepo) DeleteByAccountAndID(_ context.Context, account, id string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	for i, a := range r.articles {
		if a.Account == account && a.ID == id {
			r.articles = append(r.articles[:i], r.articles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *testArticlesRepo) DeleteByAccount(_ context.Context, account string) (int, error) {
	if r.deleteErr != nil {
		return 0, r.deleteErr
	}
	initial := len(r.articles)
	var remaining []*model.Article
	for _, a := range r.articles {
		if a.Account != account {
			remaining = append(remaining, a)
		}
	}
	r.articles = remaining
	return initial - len(r.articles), nil
}

func (r *testArticlesRepo) UpdateFavorite(_ context.Context, _, _ string, _ bool) error {
	return r.updateFavErr
}

type testUserProfileRepo struct {
	profile      *model.UserProfile
	getErr       error
	putErr       error
	deleteErr    error
	accountIDErr error
}

func (r *testUserProfileRepo) GetUserProfile(_ context.Context, _ string) (*model.UserProfile, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.profile, nil
}

func (r *testUserProfileRepo) PutUserProfile(_ context.Context, _ *model.UserProfile) error {
	return r.putErr
}

func (r *testUserProfileRepo) DeleteUserProfile(_ context.Context, _ string) error {
	return r.deleteErr
}

func (r *testUserProfileRepo) GetAccountIDByDeviceEmail(_ context.Context, _ string) (string, error) {
	if r.accountIDErr != nil {
		return "", r.accountIDErr
	}
	if r.profile != nil {
		return r.profile.Account, nil
	}
	return "", nil
}

func (r *testUserProfileRepo) DeleteUserDeviceEmail(_ context.Context, _ string) error {
	return r.deleteErr
}

type testSendsRepo struct {
	createErr   error
	countErr    error
	countResult int
}

func (r *testSendsRepo) CreateSendRecord(_ context.Context, _ *model.Send) error {
	return r.createErr
}

func (r *testSendsRepo) UpdateSendRecord(_ context.Context, _ *model.Send) error {
	return nil
}

func (r *testSendsRepo) GetSendsByArticleID(_ context.Context, _ string) ([]*model.Send, error) {
	return nil, nil
}

func (r *testSendsRepo) GetSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) ([]*model.Send, error) {
	return nil, nil
}

func (r *testSendsRepo) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return r.countResult, r.countErr
}

func TestCreateArticle_Success(t *testing.T) {
	repo := &testArticlesRepo{articles: []*model.Article{}}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	article, err := svc.CreateArticle(context.Background(), &url.URL{Scheme: "https", Host: "example.com"}, "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.Account != testAccountID {
		t.Errorf("expected account '%s', got '%s'", testAccountID, article.Account)
	}
}

func TestCreateArticle_Error(t *testing.T) {
	repo := &testArticlesRepo{storeErr: errors.New("store error")}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.CreateArticle(context.Background(), &url.URL{Scheme: "https", Host: "example.com"}, "account-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateArticle_Success(t *testing.T) {
	repo := &testArticlesRepo{articles: []*model.Article{
		{Account: "account-1", ID: "1", Title: "Original"},
	}}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.UpdateArticle(context.Background(), &model.Article{ID: "1", Account: "account-1", Title: "Updated"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateArticle_Error(t *testing.T) {
	repo := &testArticlesRepo{articles: []*model.Article{}}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	article := &model.Article{ID: "1"}
	err := svc.UpdateArticle(context.Background(), article)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestToggleFavorite_Success(t *testing.T) {
	repo := &testArticlesRepo{articles: []*model.Article{
		{Account: "account-1", ID: "article-1", Favorite: false},
	}}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.ToggleFavorite(context.Background(), "account-1", "article-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleFavorite_Error(t *testing.T) {
	repo := &testArticlesRepo{updateFavErr: errors.New("update error")}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.ToggleFavorite(context.Background(), "account-1", "article-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCountSendsByAccountDateRange_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{countResult: 5}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	count, err := svc.CountSendsByAccountDateRange(
		context.Background(),
		"account-1",
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestCountSendsByAccountDateRange_NoSendsRepo(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       nil,
	})

	count, err := svc.CountSendsByAccountDateRange(
		context.Background(),
		"account-1",
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0 (no sends repo), got %d", count)
	}
}

func TestCountSendsByAccountDateRange_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{countErr: errors.New("count error")}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.CountSendsByAccountDateRange(
		context.Background(),
		"account-1",
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUserDeviceEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: testEmail,
			AutoSend:    true,
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	eml, autoSend, sendErr := svc.GetUserDeviceEmailAndAutoSend(context.Background(), "account-1")

	if sendErr != nil {
		t.Fatalf("unexpected error: %v", sendErr)
	}

	if eml != testEmail {
		t.Errorf("expected email 'device@example.com', got '%s'", eml)
	}

	if !autoSend {
		t.Error("expected autoSend to be true")
	}
}

func TestSetUserDeviceEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmail(context.Background(), "account-1", testEmail)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetUserDeviceEmail_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{putErr: errors.New("put error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmail(context.Background(), "account-1", testEmail)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetUserDeviceEmailWithAutoSend_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), "account-1", "device@free.kindle.com", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetUserDeviceEmailWithAutoSend_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{putErr: errors.New("put error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), "account-1", testEmail, true)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUserDeviceEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.DeleteUserDeviceEmail(context.Background(), "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUserDeviceEmail_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{deleteErr: errors.New("delete error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.DeleteUserDeviceEmail(context.Background(), "account-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUserProfile_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: testEmail,
		},
	}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       &testSendsRepo{},
	})

	profile, err := svc.GetUserProfile(context.Background(), "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Account != testAccountID {
		t.Errorf("expected account '%s', got '%s'", testAccountID, profile.Account)
	}
}

func TestSendArticle_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	reader := &testReadCloser{data: []byte("epub content")}
	resp, err := svc.SendArticle(context.Background(), "test@example.com", reader, "Test Article")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.MessageID != testMessageID {
		t.Errorf("expected message ID 'msg-123', got '%s'", resp.MessageID)
	}
}

func TestSendArticle_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	sender := &errorSenderMock{}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          sender,
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	reader := &testReadCloser{data: []byte("epub content")}
	_, err := svc.SendArticle(context.Background(), "test@example.com", reader, "Test Article")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSendArticleByID_Success(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: testEmail,
			AutoSend:    true,
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          &config.Config{SenderEmail: "sender@example.com", EmailProvider: consts.EmailBackendMailjet},
	})

	result, err := svc.SendArticleByID(context.Background(), "account-1", "article-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Article.ID != "article-1" {
		t.Errorf("expected article ID 'article-1', got '%s'", result.Article.ID)
	}

	if result.DeviceEmail != testEmail {
		t.Errorf("expected device email 'device@example.com', got '%s'", result.DeviceEmail)
	}

	if result.EmailResp.MessageID != testMessageID {
		t.Errorf("expected message ID 'msg-123', got '%s'", result.EmailResp.MessageID)
	}
}

func TestSendArticleByID_NoDeviceEmail(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:  "account-1",
			AutoSend: true,
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          &config.Config{SenderEmail: "sender@example.com", EmailProvider: consts.EmailBackendMailjet},
	})

	_, err := svc.SendArticleByID(context.Background(), "account-1", "article-1")

	if err == nil {
		t.Fatal("expected error for missing device email")
	}
}

func TestFetch_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	testURL, _ := url.Parse("https://example.com")

	result, err := svc.Fetch(context.Background(), testURL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
	}
}

func TestGenerateEPUB_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>Test content</p>",
	}

	reader, err := svc.GenerateEPUB(article)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reader == nil {
		t.Fatal("expected reader to not be nil")
	}

	_ = reader.Close()
}

type testReadCloser struct {
	data []byte
	pos  int
}

func (r *testReadCloser) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *testReadCloser) Close() error {
	return nil
}

type errorSenderMock struct{}

func (m *errorSenderMock) SendEmail(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
	return nil, errors.New("send error")
}

func TestGetUserProfile_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{getErr: errors.New("get error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.GetUserProfile(context.Background(), "account-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetUserEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserEmail(context.Background(), "account-1", "user@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetUserEmail_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{putErr: errors.New("put error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserEmail(context.Background(), "account-1", "user@example.com")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUserProfile_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.DeleteUserProfile(context.Background(), "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUserProfile_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{deleteErr: errors.New("delete error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.DeleteUserProfile(context.Background(), "account-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleBounce_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: "device@free.kindle.com",
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.HandleBounce(context.Background(), "device@free.kindle.com", "bounce error")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleBounce_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{putErr: errors.New("bounce error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.HandleBounce(context.Background(), testEmail, "bounce error")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIsEmailBouncing_True(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account: "account-1",
			BouncedEmails: map[string]model.BounceInfo{
				testEmail: {Timestamp: time.Now(), Error: "bounce"},
			},
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	bouncing, err := svc.IsEmailBouncing(context.Background(), "account-1", testEmail)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bouncing {
		t.Error("expected bouncing to be true")
	}
}

func TestIsEmailBouncing_False(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: "device@free.kindle.com",
			BouncedEmails: map[string]model.BounceInfo{
				"other@free.kindle.com": {},
			},
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	bouncing, err := svc.IsEmailBouncing(context.Background(), "account-1", "device@free.kindle.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bouncing {
		t.Error("expected bouncing to be false")
	}
}

func TestGetAccountIDByDeviceEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account: "account-1",
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	accountID, err := svc.GetAccountIDByDeviceEmail(context.Background(), testEmail)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "account-1" {
		t.Errorf("expected account ID 'account-1', got '%s'", accountID)
	}
}

func TestGetAccountIDByDeviceEmail_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{accountIDErr: errors.New("get error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.GetAccountIDByDeviceEmail(context.Background(), testEmail)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type emailSenderMock struct{}

func (m *emailSenderMock) SendEmail(_ context.Context, _ *email.Request) (*email.SendEmailResponse, error) {
	return &email.SendEmailResponse{MessageID: testMessageID}, nil
}

var _ repository.ArticlesRepository = (*testArticlesRepo)(nil)
var _ repository.UserProfileRepository = (*testUserProfileRepo)(nil)
var _ repository.SendsRepository = (*testSendsRepo)(nil)

func TestNewDependenciesFromConfig_WithMailjetAndAWS(t *testing.T) {
	cfg := &config.Config{
		EmailProvider:    consts.EmailBackendMailjet,
		MailjetAPIKey:    "test-api-key",
		MailjetAPISecret: "test-secret",
		SenderEmail:      "sender@example.com",
		AWSConfig:        &aws.Config{},
		ArticlesTable:    "articles-table",
		UserProfileTable: "profiles-table",
		SendsTable:       "sends-table",
		BrowserlessKey:   "browserless-key",
	}

	deps := NewDependenciesFromConfig(cfg)

	if deps.Fetcher == nil {
		t.Error("expected Fetcher to be not nil")
	}
	if deps.Extractor == nil {
		t.Error("expected Extractor to be not nil")
	}
	if deps.Cleaner == nil {
		t.Error("expected Cleaner to be not nil")
	}
	if deps.Publisher == nil {
		t.Error("expected Publisher to be not nil")
	}
	if deps.Sender == nil {
		t.Error("expected Sender to be not nil")
	}
	if deps.ArticlesRepo == nil {
		t.Error("expected ArticlesRepo to be not nil")
	}
	if deps.UserProfileRepo == nil {
		t.Error("expected UserProfileRepo to be not nil")
	}
	if deps.SendsRepo == nil {
		t.Error("expected SendsRepo to be not nil")
	}
}

func TestNewDependenciesFromConfig_NoMailjetNoAWS(t *testing.T) {
	cfg := &config.Config{
		EmailProvider:  "unknown",
		BrowserlessKey: "browserless-key",
	}

	deps := NewDependenciesFromConfig(cfg)

	if deps.Fetcher == nil {
		t.Error("expected Fetcher to be not nil")
	}
	if deps.Sender != nil {
		t.Error("expected Sender to be nil")
	}
	if deps.ArticlesRepo != nil {
		t.Error("expected ArticlesRepo to be nil")
	}
	if deps.UserProfileRepo != nil {
		t.Error("expected UserProfileRepo to be nil")
	}
	if deps.SendsRepo != nil {
		t.Error("expected SendsRepo to be nil")
	}
}

func TestNewDependenciesFromConfig_WithSQLite(t *testing.T) {
	cfg := &config.Config{
		StorageBackend: consts.StorageBackendSQLite,
		SQLitePath:     "/tmp/test.db",
		BrowserlessKey: "browserless-key",
	}

	deps := NewDependenciesFromConfig(cfg)

	if deps.Fetcher == nil {
		t.Error("expected Fetcher to be not nil")
	}
	if deps.Extractor == nil {
		t.Error("expected Extractor to be not nil")
	}
	if deps.Cleaner == nil {
		t.Error("expected Cleaner to be not nil")
	}
	if deps.Publisher == nil {
		t.Error("expected Publisher to be not nil")
	}
	if deps.Sender != nil {
		t.Error("expected Sender to be nil (no email provider set)")
	}
	if deps.ArticlesRepo == nil {
		t.Error("expected ArticlesRepo to be not nil")
	}
	if deps.UserProfileRepo == nil {
		t.Error("expected UserProfileRepo to be not nil")
	}
	if deps.SendsRepo == nil {
		t.Error("expected SendsRepo to be not nil")
	}
	if deps.Config == nil {
		t.Error("expected Config to be not nil")
	}
}

func TestNewFromConfig(t *testing.T) {
	cfg := &config.Config{
		EmailProvider:    consts.EmailBackendMailjet,
		MailjetAPIKey:    "test-api-key",
		MailjetAPISecret: "test-secret",
		SenderEmail:      "sender@example.com",
		AWSConfig:        &aws.Config{},
		ArticlesTable:    "articles-table",
		UserProfileTable: "profiles-table",
		SendsTable:       "sends-table",
		BrowserlessKey:   "browserless-key",
	}

	svc := NewFromConfig(cfg)

	if svc == nil {
		t.Fatal("expected service to be not nil")
	}
	if svc.fetcher == nil {
		t.Error("expected fetcher to be not nil")
	}
	if svc.extractor == nil {
		t.Error("expected extractor to be not nil")
	}
	if svc.cleaner == nil {
		t.Error("expected cleaner to be not nil")
	}
	if svc.publisher == nil {
		t.Error("expected publisher to be not nil")
	}
	if svc.articles == nil {
		t.Error("expected articles to be not nil")
	}
	if svc.profile == nil {
		t.Error("expected profile to be not nil")
	}
	if svc.sender == nil {
		t.Error("expected sender to be not nil")
	}
	if svc.sendsRepo == nil {
		t.Error("expected sendsRepo to be not nil")
	}
	if svc.cfg != cfg {
		t.Error("expected cfg to be set")
	}
}

func TestCreateSendRecordSetsTimestamp(t *testing.T) {
	ctx := context.Background()

	mockRepo := &testSendsRepoTracking{}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         &content.Fetcher{},
		Extractor:       &content.DOMExtractor{},
		Cleaner:         &content.TrafilaturaCleaner{},
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    &testArticlesRepo{},
		UserProfileRepo: &testUserProfileRepo{},
		SendsRepo:       mockRepo,
		Config:          cfg,
	})

	err := svc.createSendRecord(ctx, "account-123", "article-456", "Test Article", "device@test.com")
	require.NoError(t, err)
	require.Len(t, mockRepo.createdSends, 1, "should have created one send record")

	created := mockRepo.createdSends[0]
	assert.Equal(t, "account-123", created.Account)
	assert.Equal(t, "article-456", created.ArticleID)
	assert.Equal(t, "Test Article", created.Title)
	assert.Equal(t, "device@test.com", created.DestEmail)
	assert.Equal(t, "test@example.com", created.SenderEmail)
	assert.Equal(t, "mailjet", created.Provider)

	assert.NotEqual(t, time.Time{}, created.SentAt, "SentAt should not be zero value")
}

type testSendsRepoTracking struct {
	createdSends []*model.Send
}

func (r *testSendsRepoTracking) CreateSendRecord(_ context.Context, send *model.Send) error {
	r.createdSends = append(r.createdSends, send)
	return nil
}

func (r *testSendsRepoTracking) UpdateSendRecord(_ context.Context, _ *model.Send) error {
	return nil
}

func (r *testSendsRepoTracking) GetSendsByArticleID(_ context.Context, _ string) ([]*model.Send, error) {
	return nil, nil
}

func (r *testSendsRepoTracking) GetSendsByAccountDateRange(
	_ context.Context,
	_ string,
	_,
	_ time.Time,
) ([]*model.Send, error) {
	return nil, nil
}

func (r *testSendsRepoTracking) CountSendsByAccountDateRange(_ context.Context, _ string, _, _ time.Time) (int, error) {
	return 0, nil
}

func TestParseHTMLFromSource_HttpSuccess(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	testURL, _ := url.Parse("https://example.com")

	doc, err := svc.ParseHTMLFromSource(context.Background(), testURL)

	require.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestParseHTMLFromSource_FileSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.html")
	htmlContent := "<html><body><h1>Test</h1></body></html>"
	require.NoError(t, os.WriteFile(testFile, []byte(htmlContent), 0o600))

	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	testURL, _ := url.Parse("file://" + testFile)

	doc, err := svc.ParseHTMLFromSource(context.Background(), testURL)

	require.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestParseHTMLFromSource_FileNotExists(t *testing.T) {
	testURL, _ := url.Parse("file:///nonexistent/file.html")

	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.ParseHTMLFromSource(context.Background(), testURL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat file")
}

func TestParseHTMLFromSource_FileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	testURL, _ := url.Parse("file://" + tmpDir)

	_, err := svc.ParseHTMLFromSource(context.Background(), testURL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")
}

func TestFetch_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	fetcher := content.NewFetcher("")
	svc := New(&Dependencies{
		Fetcher:         fetcher,
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	invalidURL, _ := url.Parse("invalid://not-a-valid-url")

	_, err := svc.Fetch(context.Background(), invalidURL)

	require.Error(t, err)
}

func TestParseHTML_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	fetched := &content.FetchedContent{
		HTML: &testReadCloser{data: []byte("<html><body><h1>Test</h1></body></html>")},
		URL:  &url.URL{Scheme: "https", Host: "example.com"},
		Type: content.FetcherTypeGo,
	}

	doc, err := svc.ParseHTML(context.Background(), fetched)

	require.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestParseHTML_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	fetched := &content.FetchedContent{
		HTML: &testReadCloser{data: []byte("invalid html")},
		URL:  &url.URL{Scheme: "https", Host: "example.com"},
		Type: content.FetcherTypeGo,
	}

	_, err := svc.ParseHTML(context.Background(), fetched)

	require.NoError(t, err)
}

func TestClean_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	doc := &html.Node{
		Type: html.ElementNode,
		Data: "html",
		FirstChild: &html.Node{
			Type: html.ElementNode,
			Data: "body",
			FirstChild: &html.Node{
				Type: html.ElementNode,
				Data: "p",
				FirstChild: &html.Node{
					Type: html.TextNode,
					Data: "Test content",
				},
			},
		},
	}

	testURL, _ := url.Parse("https://example.com")

	article, err := svc.Clean(context.Background(), doc, testURL)

	require.NoError(t, err)
	assert.NotNil(t, article)
}

func TestClean_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	cleaner := &errorCleanerMock{}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         cleaner,
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	doc := &html.Node{
		Type: html.ElementNode,
		Data: "html",
	}

	testURL, _ := url.Parse("https://example.com")

	_, err := svc.Clean(context.Background(), doc, testURL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clean content")
}

func TestGetArticle_Success(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	article, err := svc.GetArticle(context.Background(), "account-1", "article-1")

	require.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, "article-1", article.ID)
}

func TestGetArticle_Error(t *testing.T) {
	repo := &testArticlesRepo{getErr: errors.New("get error")}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.GetArticle(context.Background(), "account-1", "article-1")

	require.Error(t, err)
}

func TestGetArticlesMetadata_Success(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article", CreatedAt: time.Now()},
		},
	}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	result, err := svc.GetArticlesMetadata(context.Background(), "account-1", 1, 10, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Articles, 1)
}

func TestGetArticlesMetadata_Error(t *testing.T) {
	repo := &testArticlesRepo{metadataErr: errors.New("metadata error")}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.GetArticlesMetadata(context.Background(), "account-1", 1, 10, nil)

	require.Error(t, err)
}

func TestDeleteArticle_Success(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	result, err := svc.DeleteArticle(context.Background(), "account-1", "article-1")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Deleted)
}

func TestDeleteArticle_Error(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
		deleteErr: errors.New("delete error"),
	}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.DeleteArticle(context.Background(), "account-1", "article-1")

	require.Error(t, err)
}

func TestSendArticleByID_EmailSendFailure(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: testEmail,
			AutoSend:    true,
		},
	}
	sendsRepo := &testSendsRepo{}

	sender := &errorSenderMock{}
	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          sender,
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          cfg,
	})

	_, err := svc.SendArticleByID(context.Background(), "account-1", "article-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "send error")
}

func TestSendArticleByID_CreateSendRecordError(t *testing.T) {
	repo := &testArticlesRepo{
		articles: []*model.Article{
			{Account: "account-1", ID: "article-1", Title: "Test Article"},
		},
	}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: testEmail,
			AutoSend:    true,
		},
	}
	sendsRepo := &testSendsRepo{createErr: errors.New("create error")}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          cfg,
	})

	_, err := svc.SendArticleByID(context.Background(), "account-1", "article-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create send record")
}

func TestUpdateSendRecordOnFailure_WithSendsRepo(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          cfg,
	})

	err := svc.updateSendRecordOnFailure(context.Background(), "account-1", "article-1", errors.New("send failed"))

	require.NoError(t, err)
}

func TestUpdateSendRecordOnFailure_WithoutSendsRepo(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       nil,
		Config:          cfg,
	})

	err := svc.updateSendRecordOnFailure(context.Background(), "account-1", "article-1", errors.New("send failed"))

	require.NoError(t, err)
}

func TestUpdateSendRecordOnSuccess_WithSendsRepo(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
		Config:          cfg,
	})

	err := svc.updateSendRecordOnSuccess(context.Background(), "account-1", "article-1", "msg-123")

	require.NoError(t, err)
}

func TestUpdateSendRecordOnSuccess_WithoutSendsRepo(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}

	cfg := &config.Config{
		SenderEmail:   "test@example.com",
		EmailProvider: consts.EmailBackendMailjet,
	}
	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewDOMExtractor(),
		Cleaner:         content.NewTrafilaturaCleaner(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       nil,
		Config:          cfg,
	})

	err := svc.updateSendRecordOnSuccess(context.Background(), "account-1", "article-1", "msg-123")

	require.NoError(t, err)
}

type errorCleanerMock struct{}

func (m *errorCleanerMock) Clean(_ context.Context, _ *html.Node, _ *url.URL) (*model.Article, error) {
	return nil, errors.New("clean error")
}

func TestReadEPUB(t *testing.T) {
	// Create a temporary test EPUB file
	testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")

	// Check if the test file exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("test EPUB file not found")
	}

	cfg := &config.Config{
		Mode: consts.ModeCLI,
	}
	svc := New(&Dependencies{
		Fetcher:   content.NewFetcher(""),
		Extractor: content.NewDOMExtractor(),
		Cleaner:   content.NewTrafilaturaCleaner(),
		Publisher: epub.NewPublisher(epub.WithMemoryStorage()),
		Reader:    epub.NewReader(),
		Config:    cfg,
	})

	ctx := context.Background()

	t.Run("successfully reads EPUB from file:// URL", func(t *testing.T) {
		absPath, err := filepath.Abs(testDataPath)
		require.NoError(t, err)
		u, err := url.Parse("file://" + absPath)
		require.NoError(t, err)

		epubReader, title, err := svc.ReadEPUB(ctx, u)
		require.NoError(t, err)
		require.NotEmpty(t, title, "title should not be empty")
		require.NotNil(t, epubReader, "reader should not be nil")

		// Verify that we can read from the reader
		data, err := io.ReadAll(epubReader)
		require.NoError(t, err)
		require.Greater(t, len(data), 0, "EPUB data should not be empty")

		err = epubReader.Close()
		require.NoError(t, err)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		u, err := url.Parse("file:///nonexistent/file.epub")
		require.NoError(t, err)

		_, _, err = svc.ReadEPUB(ctx, u)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to stat file")
	})

	t.Run("returns error for non-EPUB file", func(t *testing.T) {
		// Create a temporary non-EPUB file
		tmpFile, err := os.CreateTemp("", "test-*.txt")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		_, err = tmpFile.WriteString("This is not an EPUB file")
		require.NoError(t, err)
		_ = tmpFile.Close()

		u, err := url.Parse("file://" + tmpFile.Name())
		require.NoError(t, err)

		_, _, err = svc.ReadEPUB(ctx, u)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to open epub as zip")
	})

	t.Run("returns error for unsupported scheme", func(t *testing.T) {
		u, err := url.Parse("ftp://example.com/file.epub")
		require.NoError(t, err)

		_, _, err = svc.ReadEPUB(ctx, u)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported url scheme")
	})
}
