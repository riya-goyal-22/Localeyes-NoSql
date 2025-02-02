package interfaces

import (
	"context"
	"localeyes/config"
	"localeyes/internal/models"
	"time"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetAllPostsWithFilter(ctx context.Context, limit *int, offset *int, search *string, filter *string) ([]*models.Post, error)
	DeletePost(ctx context.Context, filter config.Filter, createdAt time.Time, uId string, pId string) error
	GetPostsByUId(ctx context.Context, uId string) ([]*models.Post, error)
	UpdatePost(ctx context.Context, uId string, post *models.Post) error
	ToggleLike(ctx context.Context, uId string, filter string, pId string, createdAt time.Time) (config.LikeStatus, error)
	HasUserLikedAPost(ctx context.Context, uId string, pId string) (bool, error)
}
