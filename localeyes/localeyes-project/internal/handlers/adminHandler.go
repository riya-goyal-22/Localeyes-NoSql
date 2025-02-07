package handlers

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
	"localeyes/utils"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	service   interfaces.AdminServiceInterface
	validator *validator.Validate
}

func NewAdminHandler(service interfaces.AdminServiceInterface, validator *validator.Validate) *AdminHandler {
	return &AdminHandler{
		service,
		validator,
	}
}

func (handler *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	params := &models.GetUsersParams{}
	queryParams := r.URL.Query()
	params.Search = queryParams.Get("search")
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	params.Limit = int32(limit)
	offset, err := strconv.Atoi(queryParams.Get("offset"))
	params.Offset = int32(offset)

	users, err := handler.service.GetAllUsers(r.Context(), *params)
	if err != nil {
		response := utils.NewInternalServerError("Error fetching all users" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := &models.Response{
		Message: "Successfully got all users",
		Data:    users,
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var user models.DeleteUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		response := utils.NewBadRequestError("Invalid JSON body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(user)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	user.UId = userId
	err = handler.service.DeleteUser(r.Context(), &user)
	if err != nil {
		response := utils.NewInternalServerError("Error deleting user" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Message: "Successfully deleted user",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *AdminHandler) ReActivateUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	err := handler.service.ReactivateUser(r.Context(), userId)
	if err != nil {
		response := utils.NewInternalServerError("Error reactivating user" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Message: "Successfully re-activated user",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	userId := mux.Vars(r)["user_id"]
	var post models.DeletePost
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		response := utils.NewBadRequestError("Invalid JSON body")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		response := utils.NewBadRequestError("Invalid Input")
		response.ToJson(w, http.StatusBadRequest)
		return
	}
	err = handler.service.DeletePost(r.Context(), userId, postId, &post)
	if err != nil {
		response := utils.NewInternalServerError("Error deleting post" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Message: "Successfully deleted post",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := mux.Vars(r)["ques_id"]
	postId := mux.Vars(r)["post_id"]
	userId := mux.Vars(r)["user_id"]
	err := handler.service.DeleteQuestion(r.Context(), postId, questionId, userId)
	if err != nil {
		response := utils.NewInternalServerError("Error deleting question" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Message: "Successfully deleted question",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}

func (handler *AdminHandler) DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	questionId := mux.Vars(r)["ques_id"]
	ansId := mux.Vars(r)["answer_id"]
	userId := mux.Vars(r)["user_id"]
	err := handler.service.DeleteAnswer(r.Context(), ansId, questionId, userId)
	if err != nil {
		response := utils.NewInternalServerError("Error deleting reply" + err.Error())
		response.ToJson(w, http.StatusInternalServerError)
		return
	}
	response := models.Response{
		Message: "Successfully deleted answer",
		Code:    http.StatusOK,
	}
	response.ToJson(w, http.StatusOK)
	return
}
