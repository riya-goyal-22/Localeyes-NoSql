package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"os"
)

func InitSNS() (*sns.Client, string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(os.Getenv("SNS_REGION")),
	)
	if err != nil {
		return nil, "", err
	}

	SNSClient := sns.NewFromConfig(cfg)
	TopicArn := os.Getenv("SNS_TOPIC_ARN")
	return SNSClient, TopicArn, nil
}
