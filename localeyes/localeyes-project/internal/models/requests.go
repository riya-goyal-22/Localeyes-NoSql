package models

import (
	"localeyes/config"
	"time"
)

type LivingSince struct {
	Days   float64 `json:"days"`
	Months float64 `json:"months"`
	Years  float64 `json:"years"`
}

type Client struct {
	Username    string      `json:"username" validate:"required"`
	Password    string      `json:"password" validate:"required,isValidPassword"`
	City        string      `json:"city" validate:"required"`
	LivingSince LivingSince `json:"living_since" validate:"required"`
	Email       string      `json:"email" validate:"required,email"`
}

type UpdateClient struct {
	Password    string      `json:"password" validate:"required,isValidPassword"`
	City        string      `json:"city" validate:"required"`
	LivingSince LivingSince `json:"living_since" validate:"required"`
}

type ClientLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RequestPost struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
	Type    string `json:"type" validate:"required,isValidFilter"`
}

type UpdatePost struct {
	PostId    string        `json:"post_id"`
	UId       string        `json:"user_id"`
	Title     string        `json:"title" validate:"required"`
	Type      config.Filter `json:"type" validate:"required,isValidFilter"`
	Content   string        `json:"content" validate:"required"`
	CreatedAt time.Time     `json:"created_at" validate:"required,isValidTime"`
}

type DeleteOrLikePost struct {
	CreatedAt time.Time     `json:"created_at" validate:"required,isValidTime"`
	Type      config.Filter `json:"type" validate:"required,isValidFilter"`
}

type RequestQuestion struct {
	PostId string `json:"post_id" validate:"required"`
	UserId string `json:"q_user_id"`
	Text   string `json:"text" validate:"required"`
}

type RequestAnswer struct {
	QId    string `json:"q_id" validate:"required"`
	Answer string `json:"answer" validate:"required"`
	UserId string `json:"r_user_id"`
}
