package interfaces

type OTPRepoInterface interface {
	GenerateOTP() (string, error)
	SaveOTP(email string, otp string)
	ValidateOTP(email string, otp string) bool
}
