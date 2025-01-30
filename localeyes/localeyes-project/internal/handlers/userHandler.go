package handlers

import (
	_ "database/sql"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"localeyes/config"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
	"localeyes/utils"
	"net/http"
	"strconv"
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

func (handler *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var client models.Client
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	// validate userInput
	err = handler.validator.Struct(client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError(err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	var livingSinceInYears = client.LivingSince.Days/365.0 + client.LivingSince.Months/12.0 + client.LivingSince.Years

	err = handler.service.Signup(r.Context(), client.Username, client.Password, client.Email, livingSinceInYears)
	if err != nil {
		if errors.Is(err, utils.UserExistsName) {
			w.WriteHeader(http.StatusBadRequest)
			response := utils.NewBadRequestError("User already exists with the given username")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		} else if errors.Is(err, utils.UserExistsEmail) {
			w.WriteHeader(http.StatusBadRequest)
			response := utils.NewBadRequestError("User already exists with the given email")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error signing up" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	utils.Logger.Info("User signed up successfully")
	response := &models.Response{
		Message: "User created successfully",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var client models.ClientLogin
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}

	// validate userInput
	err = handler.validator.Struct(client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}

	user, err := handler.service.Login(r.Context(), client.Username, client.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError(err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	generatedToken, err := utils.GenerateTokenFunc(user.Username, user.UId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error generating token ")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("User logged in successfully")
	response := models.Response{
		Data:    generatedToken,
		Code:    http.StatusOK,
		Message: "User logged in successfully",
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *UserHandler) DeActivate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bearerToken := r.Header.Get("Authorization")
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Invalid token")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	id := claims["id"].(string)
	//intId := int(id)
	err = handler.service.DeActivate(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deactivating user")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("User deactivated successfully")
	response := &models.Response{
		Message: "User Deactivated successfully",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
	return
}

func (handler *UserHandler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bearerToken := r.Header.Get("Authorization")
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewUnauthorizedError("Error extracting claims")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	id := claims["id"].(string)
	user, err := handler.service.FetchProfile(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid token")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
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
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("User viewed successfully")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
	return
}

func (handler *UserHandler) ViewNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bearerToken := r.Header.Get("Authorization")
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid token")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	id := claims["id"].(string)

	notifications, err := handler.service.GetNotifications(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error getting notifications")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
		Data:    notifications,
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("User viewed successfully")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["user_id"]
	user, err := handler.service.FetchProfile(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error getting user")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
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
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *UserHandler) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["user_id"]
	var newUser models.UpdateClient
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid JSON body")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	// validate userInput
	err = handler.validator.Struct(newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	err = handler.service.UpdateUser(r.Context(), id, &newUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error updating user")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
	}
	w.WriteHeader(http.StatusOK)
	response := &models.Response{
		Message: "Success",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
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
	if filter == "" {
		filterPointer = nil
	} else {
		if !utils.IsValidFilter(filter) {
			filterPointer = nil
		} else {
			filterPointer = &filter
		}
	}
	if search == "" {
		searchPointer = nil
	} else {
		searchPointer = &search
	}
	posts, err := handler.service.GiveAllPosts(r.Context(), limitPointer, offsetPointer, searchPointer, filterPointer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError(err.Error())
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
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
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully displayed posts")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
	return
}

func (handler *UserHandler) DisplayUserPosts(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid token")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	id := claims["id"].(string)
	posts, err := handler.service.GiveUserPosts(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error displaying posts")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
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
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully displayed user posts")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
	return
}

func (handler *UserHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	var requestPost models.RequestPost
	err := json.NewDecoder(r.Body).Decode(&requestPost)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Missing Request body")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	// validate userInput
	err = handler.validator.Struct(requestPost)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	//if !utils.ValidateFilter(requestPost.Type) {
	//	w.WriteHeader(http.StatusBadRequest)
	//	response := utils.NewBadRequestError("Invalid Filter for post")
	//	err = json.NewEncoder(w).Encode(response)
	//	if err != nil {
	//		utils.Logger.Error("ERROR: Error encoding response")
	//	}
	//	return
	//}
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid Token")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	id := claims["id"].(string)
	err = handler.service.CreatePost(r.Context(), id, requestPost.Title, requestPost.Content, config.Filter(requestPost.Type))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error creating post")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	utils.Logger.Info("Successfully created post")
	response := &models.Response{
		Message: "Post created successfully",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}

func (handler *UserHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	postId := mux.Vars(r)["post_id"]

	var post models.UpdatePost
	post.PostId = postId
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid/Missing Request body")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError(err.Error())
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid token")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	id := claims["id"].(string)
	post.UId = id
	err = handler.service.UpdatePost(r.Context(), &post)
	if err != nil {
		if errors.Is(err, utils.NotYourPost) {
			w.WriteHeader(http.StatusNotFound)
			response := utils.NewNotFoundError(err.Error())
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("Error encoding response:" + err.Error())
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error updating post")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully updated post")
	response := &models.Response{
		Message: "Post updated successfully",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}

func (handler *UserHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	postId := mux.Vars(r)["post_id"]
	var post models.DeleteOrLikePost
	err := json.NewDecoder(r.Body).Decode(&post)
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		response := utils.NewUnauthorizedError("Invalid token")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	id := claims["id"].(string)
	err = handler.service.DeleteUserPost(r.Context(), id, postId, &post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deleting post")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully deleted post")
	response := &models.Response{
		Message: "Post deleted successfully",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}

func (handler *UserHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	bearerToken := r.Header.Get("Authorization")
	var post models.DeleteOrLikePost
	err := json.NewDecoder(r.Body).Decode(&post)
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewUnauthorizedError("Error extracting claims")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	userId := claims["id"].(string)
	status, err := handler.service.Like(r.Context(), userId, postId, &post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error liking post")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully liked post")
	response := &models.Response{
		Message: "Post liked successfully",
		Code:    http.StatusOK,
		Data:    status,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}

func (handler *UserHandler) GetLikeStatus(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	bearerToken := r.Header.Get("Authorization")
	claims, err := utils.ExtractClaimsFunc(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewUnauthorizedError("Error extracting claims")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	userId := claims["id"].(string)
	status, err := handler.service.GetLikeStatus(r.Context(), userId, postId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error disliking post")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully disliked post")
	response := &models.Response{
		Message: "Post disliked successfully",
		Code:    http.StatusOK,
		Data:    status,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}

// question related handlers

func (handler *UserHandler) GetAllQuestions(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	questions, err := handler.service.GetQuestionByPId(r.Context(), postId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error getting questions")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("Error encoding response:" + err.Error())
		}
	}
	w.WriteHeader(http.StatusOK)
	utils.Logger.Info("Successfully retrieved all questions")
	response := &models.Response{
		Message: "Successfully retrieved all questions",
		Code:    http.StatusOK,
		Data:    questions,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("Error encoding response:" + err.Error())
	}
}
