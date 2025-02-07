package repositories

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"localeyes/internal/models"
	"math/big"
	"os"
	"time"
)

type OtpRepository struct {
	Db        *dynamodb.Client
	TableName string
}

func NewOtpRepository(db *dynamodb.Client) *OtpRepository {
	return &OtpRepository{
		db,
		os.Getenv("TABLE_NAME"),
	}
}

func (repo *OtpRepository) GenerateOTP() (string, error) {
	otp := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num)
	}
	return otp, nil
}

func (repo *OtpRepository) SaveOTP(ctx context.Context, email, otp string) error {
	otpModal := &models.OTP{
		Email: "otp:email:" + email,
		Otp:   otp,
		TTl:   time.Now().Add(10 * time.Minute).Unix(),
	}
	otpModalAv, err := attributevalue.MarshalMap(otpModal)
	if err != nil {
		return err
	}
	_, err = repo.Db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(repo.TableName),
		Item:      otpModalAv,
	})
	return err
}

func (repo *OtpRepository) ValidateOTP(ctx context.Context, email, otp string) bool {
	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("otp:email:%s", email)},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s", otp)},
		},
	})
	return err == nil && len(result.Items) == 1
}
