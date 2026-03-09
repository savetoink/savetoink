package service

import (
	"context"
	"errors"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	svcemail "github.com/shaftoe/savetoink/backend/lib/email"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"github.com/shaftoe/savetoink/backend/lib/repository"
	repoimpl "github.com/shaftoe/savetoink/backend/lib/repository/dynamodb"
	"github.com/shaftoe/savetoink/backend/lib/service/content"
	"github.com/shaftoe/savetoink/backend/lib/service/epub"
)

const testAccountID = "account-1"

type testArticlesRepo struct {
	articles     []*model.Article
	storeErr     error
	getErr       error
	metadataErr  error
	deleteErr    error
	updateFavErr error
	dbErr        error
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
) (articles []*model.Article, lastKey map[string]awstypes.AttributeValue, total int, err error) {
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

func (r *testSendsRepo) CreateSendRecord(_ context.Context, _, _, _, _ string) error {
	return r.createErr
}

func (r *testSendsRepo) UpdateSendRecord(_ context.Context, _, _, _, _, _ string) error {
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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	article, err := svc.CreateArticle(context.Background(), "https://example.com", "account-1")

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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.CreateArticle(context.Background(), "https://example.com", "account-1")

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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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

func TestCountSendsByAccountDateRange_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{countErr: errors.New("count error")}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
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

func TestGetDBError_NoError(t *testing.T) {
	repo := &testArticlesRepo{dbErr: nil}
	profileRepo := &testUserProfileRepo{}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.GetDBError()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUserDeviceEmail_Success(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{
		profile: &model.UserProfile{
			Account:     "account-1",
			DeviceEmail: "device@example.com",
			AutoSend:    true,
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	email, autoSend, err := svc.GetUserDeviceEmail(context.Background(), "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "device@example.com" {
		t.Errorf("expected email 'device@example.com', got '%s'", email)
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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmail(context.Background(), "account-1", "device@example.com")

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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmail(context.Background(), "account-1", "device@example.com")

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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.SetUserDeviceEmailWithAutoSend(context.Background(), "account-1", "device@example.com", true)

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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
			DeviceEmail: "device@example.com",
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	profile, err := svc.GetUserProfile(context.Background(), "account-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Account != testAccountID {
		t.Errorf("expected account '%s', got '%s'", testAccountID, profile.Account)
	}
}

func TestGetUserProfile_Error(t *testing.T) {
	repo := &testArticlesRepo{}
	profileRepo := &testUserProfileRepo{getErr: errors.New("get error")}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	err := svc.HandleBounce(context.Background(), "device@example.com", "bounce error")

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
				"device@example.com": {Timestamp: time.Now(), Error: "bounce"},
			},
		},
	}
	sendsRepo := &testSendsRepo{}

	svc := New(&Dependencies{
		Fetcher:         content.NewFetcher(""),
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	bouncing, err := svc.IsEmailBouncing(context.Background(), "account-1", "device@example.com")

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
		Extractor:       content.NewExtractor(),
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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	accountID, err := svc.GetAccountIDByDeviceEmail(context.Background(), "device@example.com")

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
		Extractor:       content.NewExtractor(),
		Publisher:       epub.NewPublisher(),
		Sender:          &emailSenderMock{},
		ArticlesRepo:    repo,
		UserProfileRepo: profileRepo,
		SendsRepo:       sendsRepo,
	})

	_, err := svc.GetAccountIDByDeviceEmail(context.Background(), "device@example.com")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type emailSenderMock struct{}

func (m *emailSenderMock) SendEmail(_ context.Context, _ *svcemail.Request) (*svcemail.SendEmailResponse, error) {
	return &svcemail.SendEmailResponse{MessageID: "msg-123"}, nil
}

var _ repository.ArticlesRepository = (*testArticlesRepo)(nil)
var _ repository.UserProfileRepository = (*testUserProfileRepo)(nil)
var _ repository.SendsRepository = (*testSendsRepo)(nil)
