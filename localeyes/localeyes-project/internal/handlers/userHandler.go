package handlers

import (
	"context"
	_ "database/sql"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"localeyes/config"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
	"localeyes/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	service   interfaces.UserServiceInterface
	validator *validator.Validate
}

func NewUserHandler(service interfaces.UserServiceInterface, validator *validator.Validate) *UserHandler {
	return &UserHandler{
		service,
		validator,
	}
}

func (handler *UserHandler) SendOtp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var userEmail models.UserEmail
	err := json.NewDecoder(r.Body).Decode(&userEmail)
	if err != nil {
		response := utils.NewBadRequestError("Invalid JSON body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(userEmail)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.service.SendOtp(r.Context(), userEmail.Email)
	if err != nil {
		if errors.Is(err, utils.NoUser) {
			response := utils.NewBadRequestError("No User")
			response.ToJson(w, http.StatusBadRequest)
			return
		}
		response := utils.NewInternalServerError("Internal server error" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resetUser models.ResetPasswordUser
	err := json.NewDecoder(r.Body).Decode(&resetUser)
	if err != nil {
		response := utils.NewBadRequestError("Invalid JSON body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(resetUser)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.service.PasswordReset(r.Context(), resetUser)
	if err != nil {
		if errors.Is(err, utils.WrongOTP) {
			response := utils.NewInternalServerError(err.Error())
			response.ToJson(w, http.StatusBadRequest)
			return
		}
		response := utils.NewInternalServerError("Internal server error")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// On Hold
	w.WriteHeader(http.StatusNotImplemented)
	response := &models.Response{
		Message: "This service is not available yet",
		Code:    http.StatusNotImplemented,
	}
	if response.Code == http.StatusNotImplemented {
		snsClient, topicArn, err := config.InitSNS()
		if err != nil {
			log.Printf("Failed to initialize SNS client: %v", err)
		} else {
			messageBody, _ := json.Marshal(map[string]string{
				"error_code":  "5500",
				"status_code": strconv.Itoa(response.Code),
				"message":     response.Message,
				"timestamp":   time.Now().Format(time.RFC3339),
			})

			message := &sns.PublishInput{
				Message:  aws.String(string(messageBody)),
				TopicArn: aws.String(topicArn),
				MessageAttributes: map[string]types.MessageAttributeValue{
					"ErrorCode": {
						DataType:    aws.String("String"),
						StringValue: aws.String("5500"),
					},
					"StatusCode": {
						DataType:    aws.String("Number"),
						StringValue: aws.String(strconv.Itoa(response.Code)),
					},
				},
			}

			_, err = snsClient.Publish(context.Background(), message)
			if err != nil {
				log.Printf("Failed to send message to SNS: %v", err)
			} else {
				log.Println("Error message successfully sent to SNS.")
			}
		}
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	return
}

func (handler *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var client models.Client
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Json")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	// validate userInput
	err = handler.validator.Struct(client)
	if err != nil {
		response := utils.NewBadRequestError(err.Error())
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	var livingSinceInYears = client.LivingSince.Days/365.0 + client.LivingSince.Months/12.0 + client.LivingSince.Years

	err = handler.service.Signup(r.Context(), client.Username, client.Password, client.Email, livingSinceInYears)
	if err != nil {
		if errors.Is(err, utils.UserExistsName) {
			response := utils.NewBadRequestError("Username already registered")
			response.ToJson(w, http.StatusBadRequest)
			return
		} else if errors.Is(err, utils.UserExistsEmail) {
			response := utils.NewBadRequestError("Email already registered")
			response.ToJson(w, http.StatusBadRequest)
			return
		}
		response := utils.NewInternalServerError("Error signing up" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("User signed up successfully")
	response := &models.Response{
		Message: "User created successfully",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusCreated)
	return
}

func (handler *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var client models.ClientLogin
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}

	// validate userInput
	err = handler.validator.Struct(client)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}

	user, err := handler.service.Login(r.Context(), client.Username, client.Password)
	if err != nil {
		response := utils.NewUnauthorizedError(err.Error())
		response.ToJson(w, http.StatusUnauthorized)
		return
	}
	generatedToken, err := utils.GenerateTokenFunc(user.Username, user.UId)
	if err != nil {
		response := utils.NewInternalServerError("Error generating token ")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("User logged in successfully")
	response := models.Response{
		Data:    generatedToken,
		Code:    http.StatusOK,
		Message: "User logged in successfully",
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) DeActivate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.Context().Value("Id").(string)
	err := handler.service.DeActivate(r.Context(), id)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("User deactivated successfully")
	response := &models.Response{
		Message: "User Deactivated successfully",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.Context().Value("Id").(string)
	user, err := handler.service.FetchProfile(r.Context(), id)
	if err != nil {
		response := utils.NewUnauthorizedError("Invalid token")
		response.ToJson(w, http.StatusUnauthorized)
		return
	}
	responseUser := &models.ResponseUser{
		UId:          user.UId,
		Email:        user.Email,
		Username:     user.Username,
		City:         user.City,
		LivingSince:  user.DwellingAge,
		Tag:          user.Tag,
		ActiveStatus: user.IsActive,
	}
	response := models.Response{
		Data:    responseUser,
		Code:    http.StatusOK,
		Message: "User viewed successfully",
	}
	utils.Logger.Info("User viewed successfully")
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) ViewNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.Context().Value("Id").(string)

	notifications, err := handler.service.GetNotifications(r.Context(), id)
	if err != nil {
		response := utils.NewInternalServerError("Error getting notifications")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	var dtoNotifications []*models.Notification
	for _, v := range notifications {
		notification := &models.Notification{
			PostId:    v.PostId,
			UId:       v.UId,
			Title:     v.Title,
			Type:      v.Type,
			Content:   v.Content,
			Likes:     v.Likes,
			CreatedAt: v.CreatedAt,
		}
		dtoNotifications = append(dtoNotifications, notification)
	}
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
		Data:    dtoNotifications,
	}
	utils.Logger.Info("User viewed successfully")
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["user_id"]
	user, err := handler.service.FetchProfile(r.Context(), id)
	if err != nil {
		response := utils.NewInternalServerError("Error getting user")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	responseUser := models.ResponseUser{
		UId:         user.UId,
		Email:       user.Email,
		Username:    user.Username,
		City:        user.City,
		LivingSince: user.DwellingAge,
		Tag:         user.Tag,
	}
	response := models.Response{
		Data:    responseUser,
		Code:    http.StatusOK,
		Message: "Success",
	}
	w.WriteHeader(http.StatusOK)
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["user_id"]
	var newUser models.UpdateClient
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		response := utils.NewBadRequestError("Invalid JSON body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	// validate userInput
	err = handler.validator.Struct(newUser)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.service.UpdateUser(r.Context(), id, &newUser)
	if err != nil {
		response := utils.NewInternalServerError("Error updating user")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

//Post related handlers

func (handler *UserHandler) DisplayPosts(w http.ResponseWriter, r *http.Request) {
	var filterPointer, searchPointer *string
	var limitPointer, offsetPointer *int
	queryParams := r.URL.Query()
	filter := queryParams.Get("filter")
	search := queryParams.Get("search")
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil {
		limitPointer = nil
	} else {
		limitPointer = &limit
	}
	offset, err := strconv.Atoi(queryParams.Get("offset"))
	if err != nil {
		offsetPointer = nil
	} else {
		offsetPointer = &offset
	}
	if !utils.IsValidFilter(filter) {
		filterPointer = nil
	} else {
		filterPointer = &filter
	}
	if search == "" {
		searchPointer = nil
	} else {
		searchPointer = &search
	}
	posts, err := handler.service.GiveAllPosts(r.Context(), limitPointer, offsetPointer, searchPointer, filterPointer)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	var responseData []models.ResponsePost
	for _, post := range posts {
		responseData = append(responseData, models.ResponsePost{
			PostId:    post.PostId,
			UId:       post.UId,
			Title:     post.Title,
			Type:      post.Type,
			Content:   post.Content,
			Likes:     post.Likes,
			CreatedAt: post.CreatedAt,
		})
	}
	response := models.Response{
		Data:    responseData,
		Code:    http.StatusOK,
		Message: "Success",
	}
	utils.Logger.Info("Successfully displayed posts")
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) DisplayUserPosts(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("Id").(string)
	posts, err := handler.service.GiveUserPosts(r.Context(), id)
	if err != nil {
		response := utils.NewInternalServerError("Error displaying posts")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	var responseData []models.ResponsePost
	for _, post := range posts {
		responseData = append(responseData, models.ResponsePost{
			PostId:    post.PostId,
			UId:       post.UId,
			Title:     post.Title,
			Type:      post.Type,
			Content:   post.Content,
			Likes:     post.Likes,
			CreatedAt: post.CreatedAt,
		})
	}
	response := models.Response{
		Data:    responseData,
		Code:    http.StatusOK,
		Message: "Success",
	}
	utils.Logger.Info("Successfully displayed user posts")
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var requestPost models.RequestPost
	id := r.Context().Value("Id").(string)
	err := json.NewDecoder(r.Body).Decode(&requestPost)
	if err != nil {
		response := utils.NewBadRequestError("Missing Request body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	// validate userInput
	err = handler.validator.Struct(requestPost)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.service.CreatePost(r.Context(), id, requestPost.Title, requestPost.Content, config.Filter(requestPost.Type))
	if err != nil {
		response := utils.NewInternalServerError("Error creating post")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully created post")
	response := &models.Response{
		Message: "Post created successfully",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusCreated)
	return
}

func (handler *UserHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	id := r.Context().Value("Id").(string)
	var post models.UpdatePost
	post.PostId = postId
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response := utils.NewBadRequestError("Invalid/Missing Request body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		response := utils.NewBadRequestError(err.Error())
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	post.UId = id
	err = handler.service.UpdatePost(r.Context(), &post)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully updated post")
	response := &models.Response{
		Message: "Post updated successfully",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
}

func (handler *UserHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	id := r.Context().Value("Id").(string)
	var post models.DeletePost
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response := utils.NewUnauthorizedError("Invalid Json")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		response := utils.NewBadRequestError(err.Error())
		response.ToJson(w, http.StatusBadRequest)
		return
	}

	err = handler.service.DeleteUserPost(r.Context(), id, postId, &post)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully deleted post")
	response := &models.Response{
		Message: "Post deleted successfully",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	userId := r.Context().Value("Id").(string)
	var post models.LikePost
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response := utils.NewUnauthorizedError("Invalid Json")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		response := utils.NewBadRequestError(err.Error())
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	status, err := handler.service.Like(r.Context(), userId, postId, &post)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully liked post")
	response := &models.Response{
		Message: "Post liked successfully",
		Code:    http.StatusOK,
		Data:    status,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) GetLikeStatus(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	userId := r.Context().Value("Id").(string)

	status, err := handler.service.GetLikeStatus(r.Context(), userId, postId)
	if err != nil {
		response := utils.NewInternalServerError(err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully viewed the like status")
	response := &models.Response{
		Message: "Successfully viewed the like status",
		Code:    http.StatusOK,
		Data:    status,
	}
	response.ToJson(w, http.StatusOK)
	return
}

// question related handlers

func (handler *UserHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var question models.RequestQuestion
	err := json.NewDecoder(r.Body).Decode(&question)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Request Body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(question)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	postId := mux.Vars(r)["post_id"]
	userId := r.Context().Value("Id").(string)
	question.UserId = userId
	question.PostId = postId
	err = handler.service.AddQuestion(r.Context(), &question)
	if err != nil {
		response := utils.NewInternalServerError("error while asking question")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Question created")
	response := &models.Response{
		Code:    http.StatusOK,
		Message: "Question Created",
	}
	response.ToJson(w, http.StatusCreated)
	return
}

func (handler *UserHandler) GetAllQuestions(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	questions, err := handler.service.GetQuestionByPId(r.Context(), postId)
	if err != nil {
		response := utils.NewInternalServerError("Error getting questions")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully retrieved all questions")
	response := &models.Response{
		Message: "Successfully retrieved all questions",
		Code:    http.StatusOK,
		Data:    questions,
	}
	response.ToJson(w, http.StatusOK)
}

func (handler *UserHandler) AddAnswer(w http.ResponseWriter, r *http.Request) {
	quesId := mux.Vars(r)["ques_id"]
	var request models.RequestAnswer
	err := json.NewDecoder(r.Body).Decode(&request)
	userId := r.Context().Value("Id").(string)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Request Body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	request.QId = quesId
	err = handler.service.AddAnswer(r.Context(), &request)
	if err != nil {
		response := utils.NewInternalServerError("error while adding answer")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Answer added")
	response := &models.Response{
		Code:    http.StatusOK,
		Message: "Answer Added",
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	quesId := mux.Vars(r)["ques_id"]
	postId := mux.Vars(r)["post_id"]
	userId := r.Context().Value("Id").(string)

	err := handler.service.DeleteQuestion(r.Context(), postId, quesId, userId)
	if err != nil {
		response := utils.NewInternalServerError("error while deleting question")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Question deleted")
	response := &models.Response{
		Code:    http.StatusOK,
		Message: "Question Deleted",
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	quesId := mux.Vars(r)["ques_id"]
	answerId := mux.Vars(r)["answer_id"]
	userId := r.Context().Value("Id").(string)
	err := handler.service.DeleteAnswer(r.Context(), quesId, answerId, userId)
	if err != nil {
		response := utils.NewInternalServerError("error while deleting answer")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Answer deleted")
	response := &models.Response{
		Code:    http.StatusOK,
		Message: "Answer Deleted",
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *UserHandler) GetAllAnswers(w http.ResponseWriter, r *http.Request) {
	quesId := mux.Vars(r)["ques_id"]
	answers, err := handler.service.GetAllAnswers(r.Context(), quesId)
	if err != nil {
		response := utils.NewInternalServerError("error while getting answers")
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	utils.Logger.Info("Successfully retrieved all answers")
	response := &models.Response{
		Message: "Successfully retrieved all answers",
		Code:    http.StatusOK,
		Data:    answers,
	}
	response.ToJson(w, http.StatusOK)
	return
}
