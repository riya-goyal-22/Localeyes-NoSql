package interfaces

import (
	"context"
	"localeyes/internal/models"
)

type AdminServiceInterface interface {
	GetAllUsers(ctx context.Context, params models.GetUsersParams) ([]*models.ResponseUser, error)
	ReactivateUser(ctx context.Context, uId string) error
	DeleteUser(ctx context.Context, user *models.DeleteUser) error
	DeletePost(ctx context.Context, uId string, pId string, post *models.DeletePost) error
	DeleteQuestion(ctx context.Context, pId string, qId string, uId string) error
	DeleteAnswer(ctx context.Context, rId string, qId string, uId string) error
}
