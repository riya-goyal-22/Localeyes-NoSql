package interfaces

import (
	"context"
	"localeyes/internal/models"
)

type QuestionRepoInterface interface {
	Create(ctx context.Context, question *models.Question) error
	DeleteByQId(ctx context.Context, qId string, pId string, uId string) error
	GetAllQuestionsByPId(ctx context.Context, pId string) ([]*models.Question, error)
}
