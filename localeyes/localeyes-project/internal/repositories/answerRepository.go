package repositories

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"localeyes/internal/models"
	"os"
	"strings"
)

type AnswerRepository struct {
	Db        *dynamodb.Client
	TableName string
	IndexName string
}

func NewAnswerRepository(db *dynamodb.Client) *AnswerRepository {
	return &AnswerRepository{
		db,
		os.Getenv("TABLE_NAME"),
		os.Getenv("INDEX_NAME"),
	}
}

func (repo *AnswerRepository) AddAnswer(ctx context.Context, answer *models.Reply) error {
	answerNew := &models.Reply{
		RId:    "reply:" + answer.RId,
		QId:    "question:" + answer.QId,
		Answer: answer.Answer,
		UserId: answer.UserId,
	}
	answerAv, err := attributevalue.MarshalMap(answerNew)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      answerAv,
		TableName: aws.String(repo.TableName),
	}
	_, err = repo.Db.PutItem(ctx, input)
	return err
}

func (repo *AnswerRepository) DeleteAnswer(ctx context.Context, qId, rId, uId string) error {
	_, err := repo.Db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           aws.String(repo.TableName),
		ConditionExpression: aws.String("r_user_id = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: uId},
		},
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "question:" + qId},
			"sk": &types.AttributeValueMemberS{Value: "reply:" + rId},
		},
	})
	return err
}

func (repo *AnswerRepository) GetAllAnswersByQId(ctx context.Context, qId string) ([]*models.Reply, error) {
	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "question:" + qId},
			":sk": &types.AttributeValueMemberS{Value: "reply:"},
		},
	})
	if err != nil {
		return nil, err
	}
	var replies []*models.Reply
	for _, v := range result.Items {
		var reply models.Reply
		err = attributevalue.UnmarshalMap(v, &reply)
		reply.QId = strings.Split(reply.QId, ":")[1]
		reply.RId = strings.Split(reply.RId, ":")[1]
		replies = append(replies, &reply)
	}
	return replies, nil
}
