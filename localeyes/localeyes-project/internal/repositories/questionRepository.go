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

type QuestionRepository struct {
	Db        *dynamodb.Client
	TableName string
	IndexName string
}

func NewQuestionRepository(db *dynamodb.Client) *QuestionRepository {
	return &QuestionRepository{
		db,
		os.Getenv("TABLE_NAME"),
		os.Getenv("INDEX_NAME"),
	}
}

func (repo *QuestionRepository) Create(ctx context.Context, question *models.Question) error {
	questionNew := &models.Question{
		QId:    "question:" + question.QId,
		PostId: "post:" + question.PostId,
		Text:   question.Text,
		UserId: question.UserId,
	}
	questionNewAv, err := attributevalue.MarshalMap(questionNew)
	if err != nil {
		return err
	}
	_, err = repo.Db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(repo.TableName),
		Item:      questionNewAv,
	})
	return err
}

func (repo *QuestionRepository) DeleteByQId(ctx context.Context, qId, pId, uId string) error {
	result, err := repo.Db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.TableName),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: uId},
		},
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "post:" + pId},
			"sk": &types.AttributeValueMemberS{Value: "question:" + qId},
		},
		ConditionExpression: aws.String("q_user_id = :userId"),
		ReturnValues:        types.ReturnValueAllOld,
	})
	if err != nil {
		return err
	}
	if result.Attributes != nil {
		var writeRequests []types.WriteRequest
		var queryOutput *dynamodb.QueryOutput
		var err error
		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(repo.TableName),
			KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "question:" + qId},
				":sk": &types.AttributeValueMemberS{Value: "reply:"},
			},
		}
		for {
			queryOutput, err = repo.Db.Query(ctx, queryInput)
			if err != nil {
				return err
			}
			if len(queryOutput.Items) == 0 {
				break
			}
			for _, item := range queryOutput.Items {
				writeRequests = append(writeRequests, types.WriteRequest{
					DeleteRequest: &types.DeleteRequest{
						Key: map[string]types.AttributeValue{
							"pk": &types.AttributeValueMemberS{Value: item["pk"].(*types.AttributeValueMemberS).Value},
							"sk": &types.AttributeValueMemberS{Value: item["sk"].(*types.AttributeValueMemberS).Value},
						},
					},
				})
			}
			_, err := repo.Db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					repo.TableName: writeRequests,
				},
			})
			if err != nil {
				return err
			}
			if queryOutput.LastEvaluatedKey == nil {
				break
			}
			queryInput.ExclusiveStartKey = queryOutput.LastEvaluatedKey
			writeRequests = nil
		}
	}
	return err
}

func (repo *QuestionRepository) GetAllQuestionsByPId(ctx context.Context, pId string) ([]*models.Question, error) {
	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "post:" + pId},
			":sk": &types.AttributeValueMemberS{Value: "question:"},
		},
	})
	if err != nil {
		return nil, err
	}
	var questions []*models.Question
	for _, v := range result.Items {
		var question models.Question
		err := attributevalue.UnmarshalMap(v, &question)
		if err != nil {
			return nil, err
		}
		dtoQId := strings.Split(question.QId, ":")[1]
		dtoPId := strings.Split(question.PostId, ":")[1]
		question.PostId = dtoPId
		question.QId = dtoQId
		questions = append(questions, &question)
	}
	return questions, nil
}
