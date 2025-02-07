package models

type User struct {
	UId         string  `json:"id" dynamodbav:"pk"`
	Email       string  `json:"email" dynamodbav:"email"`
	Username    string  `json:"username" dynamodbav:"username"`
	Password    string  `json:"password" dynamodbav:"password"`
	City        string  `json:"city" dynamodbav:"city"`
	DwellingAge float64 `json:"dwelling_age" dynamodbav:"dwelling_age"`
	IsActive    bool    `json:"is_active" dynamodbav:"sk"`
	Tag         string  `json:"tag" dynamodbav:"tag"`
}

type UserWithStringStatus struct {
	UId         string  `json:"id" dynamodbav:"pk"`
	Email       string  `json:"email" dynamodbav:"email"`
	Username    string  `json:"username" dynamodbav:"username"`
	Password    string  `json:"password" dynamodbav:"password"`
	City        string  `json:"city" dynamodbav:"city"`
	DwellingAge float64 `json:"dwelling_age" dynamodbav:"dwelling_age"`
	IsActive    string  `json:"is_active" dynamodbav:"sk"`
	Tag         string  `json:"tag" dynamodbav:"tag"`
}

type UserSKEmail struct {
	PK          string  `json:"pk" dynamodbav:"pk"`
	UId         string  `json:"id" dynamodbav:"uid"`
	Email       string  `json:"email" dynamodbav:"sk"`
	Username    string  `json:"username" dynamodbav:"username"`
	Password    string  `json:"password" dynamodbav:"password"`
	City        string  `json:"city" dynamodbav:"city"`
	DwellingAge float64 `json:"dwelling_age" dynamodbav:"dwelling_age"`
	IsActive    bool    `json:"is_active" dynamodbav:"is_active"`
	Tag         string  `json:"tag" dynamodbav:"tag"`
}

type UserSKUsername struct {
	PK          string  `json:"pk" dynamodbav:"pk"`
	UId         string  `json:"id" dynamodbav:"uid"`
	Email       string  `json:"email" dynamodbav:"email"`
	Username    string  `json:"username" dynamodbav:"sk"`
	Password    string  `json:"password" dynamodbav:"password"`
	City        string  `json:"city" dynamodbav:"city"`
	DwellingAge float64 `json:"dwelling_age" dynamodbav:"dwelling_age"`
	IsActive    bool    `json:"is_active" dynamodbav:"is_active"`
	Tag         string  `json:"tag" dynamodbav:"tag"`
}

type UserEmail struct {
	Email string `json:"email" validate:"required"`
}

type ResetPasswordUser struct {
	Email       string `json:"email" validate:"required"`
	OTP         string `json:"otp" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type GetUsersParams struct {
	Limit  int32
	Offset int32
	Search string
}
