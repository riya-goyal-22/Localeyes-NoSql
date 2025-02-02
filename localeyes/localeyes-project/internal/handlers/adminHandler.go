package handlers

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"localeyes/internal/interfaces"
	"localeyes/internal/models"
	"localeyes/utils"
	"net/http"
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
	users, err := handler.service.GetAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error fetching all users" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := &models.Response{
		Message: "Successfully got all users",
		Data:    users,
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var user models.DeleteUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid JSON body")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	err = handler.validator.Struct(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	user.UId = userId
	err = handler.service.DeleteUser(r.Context(), &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deleting user" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := models.Response{
		Message: "Successfully deleted user",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *AdminHandler) ReActivateUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	err := handler.service.ReactivateUser(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error reactivating user" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := models.Response{
		Message: "Successfully re-activated user",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *AdminHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	userId := mux.Vars(r)["user_id"]
	var post models.DeleteOrLikePost
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid JSON body")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	err = handler.validator.Struct(post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := utils.NewBadRequestError("Invalid Input")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	err = handler.service.DeletePost(r.Context(), userId, postId, &post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deleting post" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := models.Response{
		Message: "Successfully deleted post",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := mux.Vars(r)["ques_id"]
	postId := mux.Vars(r)["post_id"]
	userId := mux.Vars(r)["user_id"]
	err := handler.service.DeleteQuestion(r.Context(), postId, questionId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deleting question" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := models.Response{
		Message: "Successfully deleted question",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}

func (handler *AdminHandler) DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	questionId := mux.Vars(r)["ques_id"]
	ansId := mux.Vars(r)["answer_id"]
	userId := mux.Vars(r)["user_id"]
	err := handler.service.DeleteAnswer(r.Context(), ansId, questionId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := utils.NewInternalServerError("Error deleting reply" + err.Error())
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			utils.Logger.Error("ERROR: Error encoding response")
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	response := models.Response{
		Message: "Successfully deleted answer",
		Code:    http.StatusOK,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.Logger.Error("ERROR: Error encoding response")
	}
}
