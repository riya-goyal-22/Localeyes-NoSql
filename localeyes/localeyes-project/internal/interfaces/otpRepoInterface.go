package interfaces

import "context"

type OTPRepoInterface interface {
	GenerateOTP() (string, error)
	SaveOTP(ctx context.Context, email string, otp string) error
	ValidateOTP(ctx context.Context, email string, otp string) bool
}
