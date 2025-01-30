package models

import (
	"localeyes/config"
	"time"
)

type Post struct {
	PostId    string        `json:"post_id" dynamodbav:"sk"`
	UId       string        `json:"user_id" dynamodbav:"pk"`
	Title     string        `json:"title" dynamodbav:"title"`
	Type      config.Filter `json:"type" dynamodbav:"type"`
	Content   string        `json:"content" dynamodbav:"content"`
	Likes     int           `json:"likes" dynamodbav:"likes"`
	CreatedAt time.Time     `json:"created_at" dynamodbav:"created_at"`
}

type PostSKFilter struct {
	PK        string    `json:"pk" dynamodbav:"pk"`
	SK        string    `json:"sk" dynamodbav:"sk"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UId       string    `json:"user_id" dynamodbav:"user_id"`
	Title     string    `json:"title" dynamodbav:"title"`
	Content   string    `json:"content" dynamodbav:"content"`
	Likes     int       `json:"likes" dynamodbav:"likes"`
}
