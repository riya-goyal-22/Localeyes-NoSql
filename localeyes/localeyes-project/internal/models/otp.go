package models

type OTP struct {
	Email string `json:"pk" dynamodbav:"pk"`
	Otp   string `json:"otp" dynamodbav:"sk"`
	TTl   int64  `json:"ttl" dynamodbav:"ttl"`
}
