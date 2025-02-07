package models

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"localeyes/config"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

type ResponseUser struct {
	UId          string  `json:"id"`
	Username     string  `json:"username"`
	City         string  `json:"city"`
	LivingSince  float64 `json:"living_since"`
	Tag          string  `json:"tag"`
	ActiveStatus bool    `json:"active_status"`
	Email        string  `json:"email"`
}

type ResponseQuestion struct {
	QId       string   `json:"question_id"`
	PostId    string   `json:"post_id"`
	UserId    string   `json:"q_user_id"`
	Text      string   `json:"text"`
	Replies   []string `json:"replies"`
	CreatedAt string   `json:"created_at"`
}

type ResponsePost struct {
	PostId    string        `json:"post_id"`
	UId       string        `json:"user_id"`
	Title     string        `json:"title"`
	Type      config.Filter `json:"type"`
	Content   string        `json:"content"`
	Likes     int           `json:"likes"`
	CreatedAt time.Time     `json:"created_at"`
}

func (res *Response) ToJson(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if res.Code == 5500 {
		snsClient, topicArn, err := config.InitSNS()
		if err != nil {
			log.Printf("Failed to initialize SNS client: %v", err)
		} else {
			messageBody, _ := json.Marshal(map[string]string{
				"error_code":  strconv.Itoa(res.Code),
				"status_code": strconv.Itoa(statusCode),
				"message":     res.Message,
				"timestamp":   time.Now().Format(time.RFC3339),
			})

			message := &sns.PublishInput{
				Message:  aws.String(string(messageBody)),
				TopicArn: aws.String(topicArn),
				MessageAttributes: map[string]types.MessageAttributeValue{
					"ErrorCode": {
						DataType:    aws.String("String"),
						StringValue: aws.String(strconv.Itoa(res.Code)),
					},
					"StatusCode": {
						DataType:    aws.String("Number"),
						StringValue: aws.String(strconv.Itoa(statusCode)),
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

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Printf("Failed to encode error response: %v", err)
		return
	}
}
