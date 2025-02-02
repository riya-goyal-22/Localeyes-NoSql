package interfaces

import (
	"context"
	"localeyes/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FetchUserByEmail(ctx context.Context, email string) (*models.UserSKEmail, error)
	FetchUserByUsername(ctx context.Context, username string) (*models.UserSKUsername, error)
	FetchUserById(ctx context.Context, uid string) (*models.User, error)
	UpdateUserById(ctx context.Context, user *models.User) error
	ToggleUserActiveStatus(ctx context.Context, user *models.User) error
	FetchNotifications(ctx context.Context, uId string) ([]*models.Notification, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	DeleteUser(ctx context.Context, uId, username, email string) error
}
