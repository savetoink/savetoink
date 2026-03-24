// Package repository provides article persistence interfaces.
package repository

import (
	"context"
	"time"

	"github.com/shaftoe/savetoink/backend/lib/internal/types"
	"github.com/shaftoe/savetoink/backend/lib/model"
)

// ArticlesRepository defines the interface for article persistence.
type ArticlesRepository interface {
	Store(ctx context.Context, article *model.Article) error
	GetByAccountAndID(ctx context.Context, account, id string) (*model.Article, error)
	GetMetadataByAccount(
		ctx context.Context,
		account string,
		page, pageSize int,
		filter *types.ArticleFilter,
	) ([]*model.Article, any, int, error)
	DeleteByAccountAndID(ctx context.Context, account, id string) error
	UpdateFavorite(ctx context.Context, account, id string, favorite bool) error
}

// UserProfileRepository defines the interface for user profile persistence.
type UserProfileRepository interface {
	GetUserProfile(ctx context.Context, account string) (*model.UserProfile, error)
	GetAccountIDByDeviceEmail(ctx context.Context, deviceEmail string) (string, error)
	PutUserProfile(ctx context.Context, profile *model.UserProfile) error
	DeleteUserProfile(ctx context.Context, account string) error
	DeleteUserDeviceEmail(ctx context.Context, account string) error
}

// SendsRepository defines the interface for send persistence.
type SendsRepository interface {
	CreateSendRecord(ctx context.Context, send *model.Send) error
	UpdateSendRecord(ctx context.Context, send *model.Send) error
	GetSendsByArticleID(ctx context.Context, articleID string) ([]*model.Send, error)
	GetSendsByAccountDateRange(ctx context.Context, account string, startDate, endDate time.Time) ([]*model.Send, error)
	CountSendsByAccountDateRange(ctx context.Context, account string, startDate, endDate time.Time) (int, error)
}
