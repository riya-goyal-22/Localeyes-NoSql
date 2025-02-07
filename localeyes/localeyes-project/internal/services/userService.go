package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/gomail.v2"
	"localeyes/config"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
	"localeyes/utils"
	"math"
	"os"
	"strconv"
	"time"
)

type UserService struct {
	UserRepo interfaces.UserRepository
	PostRepo interfaces.PostRepository
	QuesRepo interfaces.QuestionRepoInterface
	AnsRepo  interfaces.AnswerRepoInterface
	OTPRepo  interfaces.OTPRepoInterface
}

func NewUserService(
	userRepo interfaces.UserRepository,
	postRepo interfaces.PostRepository,
	quesRepo interfaces.QuestionRepoInterface,
	ansRepo interfaces.AnswerRepoInterface,
	otpRepo interfaces.OTPRepoInterface,
) *UserService {
	return &UserService{
		UserRepo: userRepo,
		PostRepo: postRepo,
		QuesRepo: quesRepo,
		AnsRepo:  ansRepo,
		OTPRepo:  otpRepo,
	}
}

func (s *UserService) Signup(ctx context.Context, username, password, email string, dwellingAge float64) error {
	if !s.validateEmail(ctx, email) {
		return utils.UserExistsEmail
	}
	if !s.validateUsername(ctx, username) {
		return utils.UserExistsName
	}
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	hashedPassword := hashPassword(password)
	tag := utils.SetTag(dwellingAge)
	user := &models.User{
		UId:         uid.String(),
		Username:    username,
		Password:    hashedPassword,
		City:        "delhi",
		IsActive:    true,
		DwellingAge: math.Round(dwellingAge*100) / 100,
		Tag:         tag,
		Email:       email,
	}
	err = s.UserRepo.CreateUser(ctx, user)
	return err
}

func (s *UserService) Login(ctx context.Context, username, password string) (*models.User, error) {
	hashedPassword := hashPassword(password)
	dbUser, err := s.UserRepo.FetchUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New(err.Error())
	} else if errors.Is(err, utils.NoUser) {
		return nil, utils.InvalidAccountCredentials
	} else if dbUser.IsActive == false {
		return nil, utils.InactiveUser
	} else if dbUser.Password != hashedPassword {
		return nil, utils.InvalidAccountCredentials
	}
	user := &models.User{
		Username:    dbUser.Username,
		UId:         dbUser.UId,
		City:        dbUser.City,
		DwellingAge: dbUser.DwellingAge,
		Password:    dbUser.Password,
		Email:       dbUser.Email,
		Tag:         dbUser.Tag,
		IsActive:    dbUser.IsActive,
	}
	return user, nil
}

func (s *UserService) FetchProfile(ctx context.Context, uid string) (*models.User, error) {
	user, err := s.UserRepo.FetchUserById(ctx, uid, true)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeActivate(ctx context.Context, uid string) error {
	user, err := s.UserRepo.FetchUserById(ctx, uid, true)
	user.IsActive = false
	err = s.UserRepo.ToggleUserActiveStatus(ctx, user)

	return err
}

func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *UserService) GetNotifications(ctx context.Context, uid string) ([]*models.Notification, error) {
	notifications, err := s.UserRepo.FetchNotifications(ctx, uid)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *UserService) validateUsername(ctx context.Context, username string) bool {
	if username == "admin" || username == "Admin" {
		return false
	}
	user, err := s.UserRepo.FetchUserByUsername(ctx, username)
	if user == nil || err != nil {
		return true
	}
	return false
}

func (s *UserService) validateEmail(ctx context.Context, email string) bool {
	if email == "localeyes22@gmail.com" {
		return false
	}
	_, err := s.UserRepo.FetchUserByEmail(ctx, email)
	if err != nil {
		return true
	}
	return false
}

func (s *UserService) UpdateUser(ctx context.Context, uId string, requestUser *models.UpdateClient) error {
	var dwellingAge = (requestUser.LivingSince.Days / 365.0) + (requestUser.LivingSince.Years) + (requestUser.LivingSince.Months / 12.0)
	var hashedPassword = hashPassword(requestUser.Password)
	var tag = utils.SetTag(dwellingAge)
	user, err := s.UserRepo.FetchUserById(ctx, uId, true)
	if err != nil {
		return err
	}
	user.Tag = tag
	user.DwellingAge = dwellingAge
	user.Password = hashedPassword
	user.City = requestUser.City

	err = s.UserRepo.UpdateUserById(ctx, user)

	return err
}

//Post related functionality

func (s *UserService) CreatePost(ctx context.Context, userId string, title, content string, postType config.Filter) error {
	post := &models.Post{
		UId:       userId,
		PostId:    utils.GenerateRandomId(),
		Title:     title,
		Content:   content,
		Type:      postType,
		CreatedAt: time.Now(),
		Likes:     0,
	}
	err := s.PostRepo.Create(ctx, post)
	return err
}

func (s *UserService) UpdatePost(ctx context.Context, post *models.UpdatePost) error {
	postNew := &models.Post{
		PostId:    post.PostId,
		Title:     post.Title,
		Content:   post.Content,
		Type:      post.Type,
		UId:       post.UId,
		CreatedAt: post.CreatedAt,
	}
	err := s.PostRepo.UpdatePost(ctx, post.UId, postNew)
	return err
}

func (s *UserService) GiveAllPosts(ctx context.Context, limit, offset *int, search, filter *string) ([]*models.Post, error) {
	posts, err := s.PostRepo.GetAllPostsWithFilter(ctx, limit, offset, search, filter)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *UserService) GiveUserPosts(ctx context.Context, uId string) ([]*models.Post, error) {
	posts, err := s.PostRepo.GetPostsByUId(ctx, uId)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *UserService) DeleteUserPost(ctx context.Context, uId, pId string, post *models.DeletePost) error {
	err := s.PostRepo.DeletePost(ctx, post.Type, post.CreatedAt, uId, pId)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) Like(ctx context.Context, uId, pId string, post *models.LikePost) (config.LikeStatus, error) {
	status, err := s.PostRepo.ToggleLike(ctx, post.UId, uId, string(post.Type), pId, post.CreatedAt)
	if err != nil {
		return "0", err
	}
	return status, nil
}

func (s *UserService) GetLikeStatus(ctx context.Context, uId, pId string) (config.LikeStatus, error) {
	status, err := s.PostRepo.HasUserLikedAPost(ctx, uId, pId)
	if err != nil {
		return "0", err
	}
	if status == true {
		return config.Liked, nil
	}
	return config.NotLiked, nil
}

//forget password

func (s *UserService) SendOtp(ctx context.Context, email string) error {
	_, err := s.UserRepo.FetchUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	otp, err := s.OTPRepo.GenerateOTP()
	if err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTPSenderEmail"))
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Go SMTP Test")
	message.SetBody("text/plain", "Hello,\r\nThis is your otp to reset password: "+otp)

	port, _ := strconv.Atoi(os.Getenv("SMTPPort"))
	// Create a dialer with SMTP server information
	dialer := gomail.NewDialer(
		os.Getenv("SMTPServer"),
		port,
		os.Getenv("SMTPSenderEmail"),
		os.Getenv("SMTPSenderPassword"),
	)
	if os.Getenv("SMTPServer") == "" || os.Getenv("SMTPPort") == "" || os.Getenv("SMTPSenderEmail") == "" || os.Getenv("SMTPSenderPassword") == "" {
		return fmt.Errorf("missing required environment variables for SMTP configuration")

	}

	// Send the email via the dialer
	if err := dialer.DialAndSend(message); err != nil {
		return err
	}
	err = s.OTPRepo.SaveOTP(ctx, email, otp)
	return err
}

func (s *UserService) PasswordReset(ctx context.Context, resetUser models.ResetPasswordUser) error {
	if s.OTPRepo.ValidateOTP(ctx, resetUser.Email, resetUser.OTP) {
		user, err := s.UserRepo.FetchUserByEmail(ctx, resetUser.Email)
		if err != nil {
			return err
		}
		hashedPassword := hashPassword(resetUser.NewPassword)
		var updatedUser = models.User{
			Username:    user.Username,
			Password:    hashedPassword,
			IsActive:    user.IsActive,
			UId:         user.UId,
			City:        user.City,
			DwellingAge: user.DwellingAge,
			Tag:         user.Tag,
			Email:       user.Email,
		}
		err = s.UserRepo.UpdateUserById(ctx, &updatedUser)
		return err
	}
	return utils.WrongOTP
}

//question related services

func (s *UserService) AddQuestion(ctx context.Context, ques *models.RequestQuestion) error {
	question := &models.Question{
		QId:    utils.GenerateRandomId(),
		PostId: ques.PostId,
		Text:   ques.Text,
		UserId: ques.UserId,
	}
	err := s.QuesRepo.Create(ctx, question)
	return err
}

func (s *UserService) DeleteQuestion(ctx context.Context, pId, qId, uId string) error {
	err := s.QuesRepo.DeleteByQId(ctx, qId, pId, uId)
	return err
}

func (s *UserService) GetQuestionByPId(ctx context.Context, pId string) ([]*models.Question, error) {
	questions, err := s.QuesRepo.GetAllQuestionsByPId(ctx, pId)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

//answer related services

func (s *UserService) AddAnswer(ctx context.Context, ans *models.RequestAnswer) error {
	answer := &models.Reply{
		RId:    utils.GenerateRandomId(),
		Answer: ans.Answer,
		UserId: ans.UserId,
		QId:    ans.QId,
	}
	err := s.AnsRepo.AddAnswer(ctx, answer)
	return err
}

func (s *UserService) DeleteAnswer(ctx context.Context, qId, rId, uId string) error {
	err := s.AnsRepo.DeleteAnswer(ctx, qId, rId, uId)
	return err
}

func (s *UserService) GetAllAnswers(ctx context.Context, qId string) ([]*models.Reply, error) {
	answers, err := s.AnsRepo.GetAllAnswersByQId(ctx, qId)
	if err != nil {
		return nil, err
	}
	return answers, nil
}
