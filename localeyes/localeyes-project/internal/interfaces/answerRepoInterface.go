package interfaces

import (
	"context"
	"localeyes/internal/models"
)

type AnswerRepoInterface interface {
	AddAnswer(ctx context.Context, answer *models.Reply) error
	DeleteAnswer(ctx context.Context, qId string, rId string, uId string) error
	GetAllAnswersByQId(ctx context.Context, qId string) ([]*models.Reply, error)
}
