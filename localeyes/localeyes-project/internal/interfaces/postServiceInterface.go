package interfaces

import (
	"context"
	"localeyes/config"
	"localeyes/internal/models"
)

type PostService interface {
	CreatePost(ctx context.Context, userId string, title string, content string, postType config.Filter) error
	UpdateMyPost(postId string, userId string, title string, content string) error
	GiveAllPosts(ctx context.Context, limit *int, offset *int, search *string, filter *string) ([]*models.Post, error)
}
