package services

import (
	"context"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
)

type AdminService struct {
	UserRepo interfaces.UserRepository
	PostRepo interfaces.PostRepository
	QuesRepo interfaces.QuestionRepoInterface
	AnsRepo  interfaces.AnswerRepoInterface
}

func NewAdminService(userRepo interfaces.UserRepository, postRepo interfaces.PostRepository, quesRepo interfaces.QuestionRepoInterface, ansRepo interfaces.AnswerRepoInterface) *AdminService {
	return &AdminService{
		UserRepo: userRepo,
		PostRepo: postRepo,
		QuesRepo: quesRepo,
		AnsRepo:  ansRepo,
	}
}

func (s *AdminService) GetAllUsers(ctx context.Context) ([]*models.ResponseUser, error) {
	users, err := s.UserRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	var userResults []*models.ResponseUser
	for _, user := range users {
		userResult := &models.ResponseUser{
			UId:          user.UId,
			Username:     user.Username,
			Email:        user.Email,
			City:         user.City,
			LivingSince:  user.DwellingAge,
			Tag:          user.Tag,
			ActiveStatus: user.IsActive,
		}
		userResults = append(userResults, userResult)
	}
	return userResults, nil
}

func (s *AdminService) ReactivateUser(ctx context.Context, uId string) error {
	user, err := s.UserRepo.FetchUserById(ctx, uId)
	if err != nil {
		return err
	}
	user.IsActive = true
	err = s.UserRepo.ToggleUserActiveStatus(ctx, user)
	return err
}

func (s *AdminService) DeleteUser(ctx context.Context, user *models.DeleteUser) error {
	err := s.UserRepo.DeleteUser(ctx, user.UId, user.Username, user.Email)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminService) DeletePost(ctx context.Context, uId, pId string, post *models.DeleteOrLikePost) error {
	err := s.PostRepo.DeletePost(ctx, post.Type, post.CreatedAt, uId, pId)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminService) DeleteQuestion(ctx context.Context, pId, qId, uId string) error {
	err := s.QuesRepo.DeleteByQId(ctx, qId, pId, uId)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminService) DeleteAnswer(ctx context.Context, rId, qId, uId string) error {
	err := s.AnsRepo.DeleteAnswer(ctx, qId, rId, uId)
	if err != nil {
		return err
	}
	return nil
}
