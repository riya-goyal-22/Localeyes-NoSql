package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"localeyes/config"
	"localeyes/internal/handlers"
	"localeyes/internal/middlewares"
	"localeyes/internal/repositories"
	"localeyes/internal/services"
	"localeyes/utils"
	"net/http"
	"net/http/httptest"
	"strings"
)

var client *dynamodb.Client
var customValidator *validator.Validate

func init() {
	client = config.GetDBClient()
	customValidator = validator.New()
	_ = customValidator.RegisterValidation("isValidFilter", utils.ValidateFilter)
	_ = customValidator.RegisterValidation("isValidPassword", utils.ValidatePassword)
	_ = customValidator.RegisterValidation("isValidTime", utils.ValidateTime)
}

func createRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middlewares.AuthenticationMiddleware)
	userService := services.NewUserService(
		repositories.NewNoSQLUserRepository(client),
		repositories.NewPostRepository(client),
		repositories.NewQuestionRepository(client),
		repositories.NewAnswerRepository(client),
	)
	userHandler := handlers.NewUserHandler(userService, customValidator)

	// Define routes
	router.HandleFunc("/signup", userHandler.SignUp).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")

	router.HandleFunc("/user/profile", userHandler.ViewProfile).Methods("GET")
	router.HandleFunc("/user/deactivate", userHandler.DeActivate).Methods("POST")
	router.HandleFunc("/user/notifications", userHandler.ViewNotifications).Methods("GET")
	router.HandleFunc("/user/{user_id}", userHandler.GetUserById).Methods("GET")
	router.HandleFunc("/user/{user_id}", userHandler.UpdateUserById).Methods("PUT")
	router.HandleFunc("/post", userHandler.CreatePost).Methods("POST")
	router.HandleFunc("/posts/all", userHandler.DisplayPosts).Methods("GET")
	router.HandleFunc("/post/{post_id}/like", userHandler.LikePost).Methods("POST")
	router.HandleFunc("/user/post/{post_id}", userHandler.UpdatePost).Methods("PUT")
	router.HandleFunc("/user/post/{post_id}", userHandler.DeletePost).Methods("DELETE")
	router.HandleFunc("/user/posts/all", userHandler.DisplayUserPosts).Methods("GET")
	router.HandleFunc("user/post/{post_id}", userHandler.GetLikeStatus).Methods("GET")
	router.HandleFunc("/post/{post_id}/questions/all", userHandler.GetAllQuestions).Methods("GET")

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.AdminAuthMiddleware)

	return router
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Create the mux router
	router := createRouter()

	// Create the HTTP request from the API Gateway event
	req, err := http.NewRequestWithContext(ctx, request.HTTPMethod, request.Path, strings.NewReader(request.Body))
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// Add headers from API Gateway event to the request
	for key, value := range request.Headers {
		req.Header.Add(key, value)
	}

	// Create an in-memory response recorder
	rr := httptest.NewRecorder()

	// Serve the HTTP request using the mux router
	router.ServeHTTP(rr, req)

	// Return the API Gateway response
	return events.APIGatewayProxyResponse{
		StatusCode: rr.Code,
		Body:       rr.Body.String(),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS, PUT, DELETE",
		},
	}, nil
}

func main() {
	// Start the Lambda function
	lambda.Start(lambdaHandler)
}
