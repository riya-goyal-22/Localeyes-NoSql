package models

import (
	"localeyes/config"
	"time"
)

type Notification struct {
	PK        string        `json:"pk" dynamodbav:"pk"`
	PostId    string        `json:"post_id" dynamodbav:"sk"`
	UId       string        `json:"user_id" dynamodbav:"user_id"`
	Title     string        `json:"title" dynamodbav:"title"`
	Type      config.Filter `json:"type" dynamodbav:"type"`
	Content   string        `json:"content" dynamodbav:"content"`
	Likes     int           `json:"likes" dynamodbav:"likes"`
	CreatedAt time.Time     `json:"created_at" dynamodbav:"created_at"`
	TTl       int64         `json:"ttl" dynamodbav:"ttl"`
}
