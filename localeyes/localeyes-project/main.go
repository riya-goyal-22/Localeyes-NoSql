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
		repositories.NewOtpRepository(client),
	)
	adminService := services.NewAdminService(
		repositories.NewNoSQLUserRepository(client),
		repositories.NewPostRepository(client),
		repositories.NewQuestionRepository(client),
		repositories.NewAnswerRepository(client),
	)
	userHandler := handlers.NewUserHandler(userService, customValidator)
	adminHandler := handlers.NewAdminHandler(adminService, customValidator)

	// Define routes
	router.HandleFunc("/signup", userHandler.SignUp).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")
	router.HandleFunc("/otp", userHandler.SendOtp).Methods("POST")
	router.HandleFunc("/password/reset", userHandler.ResetPassword).Methods("POST")
	router.HandleFunc("/sns", userHandler.ForgotPassword).Methods("POST")

	router.HandleFunc("/user/profile", userHandler.ViewProfile).Methods("GET")
	router.HandleFunc("/user/deactivate", userHandler.DeActivate).Methods("POST") //need to be checked
	router.HandleFunc("/user/notifications", userHandler.ViewNotifications).Methods("GET")
	router.HandleFunc("/user/{user_id}", userHandler.GetUserById).Methods("GET")
	router.HandleFunc("/user/{user_id}", userHandler.UpdateUserById).Methods("PUT")
	router.HandleFunc("/user/post", userHandler.CreatePost).Methods("POST")
	router.HandleFunc("/posts/all", userHandler.DisplayPosts).Methods("GET") // error
	router.HandleFunc("/post/{post_id}/like", userHandler.LikePost).Methods("POST")
	router.HandleFunc("/user/post/{post_id}", userHandler.UpdatePost).Methods("PUT")
	router.HandleFunc("/user/post/{post_id}", userHandler.DeletePost).Methods("DELETE")
	router.HandleFunc("/user/posts/all", userHandler.DisplayUserPosts).Methods("GET")
	router.HandleFunc("/user/post/{post_id}", userHandler.GetLikeStatus).Methods("GET")
	router.HandleFunc("/post/{post_id}/questions/all", userHandler.GetAllQuestions).Methods("GET")
	router.HandleFunc("/post/{post_id}/question", userHandler.CreateQuestion).Methods("POST")
	router.HandleFunc("/post/{post_id}/question/{ques_id}", userHandler.DeleteQuestion).Methods("DELETE")
	router.HandleFunc("/question/{ques_id}/answer", userHandler.AddAnswer).Methods("POST")
	router.HandleFunc("/question/{ques_id}/answer/{answer_id}", userHandler.DeleteAnswer).Methods("DELETE")
	router.HandleFunc("/question/{ques_id}/answers/all", userHandler.GetAllAnswers).Methods("GET")

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.AdminAuthMiddleware)
	adminRouter.HandleFunc("/user/{user_id}", adminHandler.DeleteUser).Methods("DELETE")
	adminRouter.HandleFunc("/user/{user_id}/reactivate", adminHandler.ReActivateUser).Methods("POST")
	adminRouter.HandleFunc("/users/all", adminHandler.GetAllUsers).Methods("GET")
	adminRouter.HandleFunc("/user/{user_id}/post/{post_id}", adminHandler.DeletePost).Methods("DELETE")
	adminRouter.HandleFunc("/post/{post_id}/user/{user_id}/question/{ques_id}", adminHandler.DeleteQuestion).Methods("DELETE")
	adminRouter.HandleFunc("/question/{ques_id}/user/{user_id}/answer/{answer_id}", adminHandler.DeleteAnswer).Methods("DELETE")

	return router
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Create the mux router
	router := createRouter()

	url := request.Path
	if len(request.QueryStringParameters) > 0 {
		params := make([]string, 0)
		for key, value := range request.QueryStringParameters {
			params = append(params, key+"="+value)
		}
		url += "?" + strings.Join(params, "&")
	}

	// Create the HTTP request from the API Gateway event
	req, err := http.NewRequestWithContext(ctx, request.HTTPMethod, url, strings.NewReader(request.Body))
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

	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Requested-With",
				"Access-Control-Allow-Methods":     "POST,GET,OPTIONS,PUT,DELETE",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil
	}
	// Return the API Gateway response
	return events.APIGatewayProxyResponse{
		StatusCode: rr.Code,
		Body:       rr.Body.String(),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Requested-With",
			"Access-Control-Allow-Methods":     "POST,GET,OPTIONS,PUT,DELETE",
			"Access-Control-Allow-Credentials": "true",
			"Content-Type":                     "application/json",
		},
	}, nil
}

func main() {
	// Start the Lambda function
	lambda.Start(lambdaHandler)
}
