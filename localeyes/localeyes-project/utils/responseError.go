package utils

import (
	"localeyes/internal/models"
)

var DBError = 5500
var AuthError = 3300
var InvalidRequest = 4400

func NewNotFoundError(message string) *models.Response {
	return &models.Response{
		Message: message,
		Code:    InvalidRequest,
	}
}

func NewInternalServerError(message string) *models.Response {
	return &models.Response{
		Message: message,
		Code:    DBError,
	}
}

func NewBadRequestError(message string) *models.Response {
	return &models.Response{
		Message: message,
		Code:    InvalidRequest,
	}
}

func NewUnauthorizedError(message string) *models.Response {
	return &models.Response{
		Message: message,
		Code:    AuthError,
	}
}
