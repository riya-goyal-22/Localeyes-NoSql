package interfaces

import (
	"context"
	"localeyes/config"
	"localeyes/internal/models"
)

type UserServiceInterface interface {
	Signup(ctx context.Context, username string, password string, email string, dwellingAge float64) error
	Login(ctx context.Context, username string, password string) (*models.User, error)
	FetchProfile(ctx context.Context, uid string) (*models.User, error)
	DeActivate(ctx context.Context, uid string) error
	GetNotifications(ctx context.Context, uid string) ([]*models.Notification, error)
	UpdateUser(ctx context.Context, uId string, requestUser *models.UpdateClient) error
	CreatePost(ctx context.Context, userId string, title string, content string, postType config.Filter) error
	UpdatePost(ctx context.Context, post *models.UpdatePost) error
	GiveAllPosts(ctx context.Context, limit *int, offset *int, search *string, filter *string) ([]*models.Post, error)
	GiveUserPosts(ctx context.Context, uId string) ([]*models.Post, error)
	DeleteUserPost(ctx context.Context, uId string, pId string, post *models.DeleteOrLikePost) error
	Like(ctx context.Context, uId string, pId string, post *models.DeleteOrLikePost) (config.LikeStatus, error)
	GetLikeStatus(ctx context.Context, uId string, pId string) (config.LikeStatus, error)
	AddQuestion(ctx context.Context, ques *models.RequestQuestion) error
	DeleteQuestion(ctx context.Context, pId string, qId string, uId string) error
	GetQuestionByPId(ctx context.Context, pId string) ([]*models.Question, error)
	AddAnswer(ctx context.Context, ans *models.RequestAnswer) error
	DeleteAnswer(ctx context.Context, qId string, rId string, uId string) error
	GetAllAnswers(ctx context.Context, qId string) ([]*models.Reply, error)
}
