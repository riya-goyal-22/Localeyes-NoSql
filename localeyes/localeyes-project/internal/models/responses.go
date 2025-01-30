package models

import (
	"localeyes/config"
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
