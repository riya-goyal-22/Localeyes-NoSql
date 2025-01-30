package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"os"
)

func GetDBClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("DYNAMO_REGION")),
	)
	if err != nil {
		panic("Failed to load AWS configuration: " + err.Error())
	}
	return dynamodb.NewFromConfig(cfg)
}
