package repositories

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"localeyes/config"
	"localeyes/internal/models"
	"os"
	"strings"
	"sync"
	"time"
)

type PostRepository struct {
	Db        *dynamodb.Client
	TableName string
	IndexName string
}

func NewPostRepository(db *dynamodb.Client) *PostRepository {
	return &PostRepository{
		db,
		os.Getenv("TABLE_NAME"),
		os.Getenv("INDEX_NAME"),
	}
}

func (repo *PostRepository) Create(ctx context.Context, post *models.Post) error {
	postPKId := &models.Post{
		PostId:    "post:" + post.PostId,
		Title:     post.Title,
		Content:   post.Content,
		Type:      post.Type,
		CreatedAt: post.CreatedAt,
		UId:       "user:" + post.UId,
		Likes:     post.Likes,
	}
	postSKFilter := &models.PostSKFilter{
		Title:     post.Title,
		Content:   post.Content,
		SK:        fmt.Sprintf("post:%s:%s:%s", post.Type, post.CreatedAt.Format("RFC3339"), post.PostId),
		CreatedAt: post.CreatedAt,
		UId:       post.UId,
		Likes:     post.Likes,
		PK:        "posts",
	}
	notification := &models.Notification{
		PK:        "notifications",
		PostId:    "post:" + post.PostId,
		Title:     post.Title,
		Content:   post.Content,
		Type:      post.Type,
		CreatedAt: post.CreatedAt,
		UId:       post.UId,
		Likes:     post.Likes,
		TTl:       time.Now().Add(10 * time.Minute).Unix(),
	}
	posts := make([]map[string]types.AttributeValue, 0)

	postPKIdAv, err := attributevalue.MarshalMap(postPKId)
	postSKFilterAv, err := attributevalue.MarshalMap(postSKFilter)
	notificationAv, err := attributevalue.MarshalMap(notification)
	postPKIdAv["created_at"] = &types.AttributeValueMemberS{
		Value: post.CreatedAt.Format("RFC3339"),
	}
	notificationAv["created_at"] = &types.AttributeValueMemberS{
		Value: post.CreatedAt.Format("RFC3339"),
	}
	posts = append(posts, postPKIdAv, postSKFilterAv, notificationAv)
	writeRequests := make([]types.WriteRequest, len(posts))
	for i, post := range posts {
		writeRequests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: post,
			},
		}
	}
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			repo.TableName: writeRequests,
		},
	}
	_, err = repo.Db.BatchWriteItem(ctx, input)
	return err
}

func (repo *PostRepository) GetAllPostsWithFilter(ctx context.Context, limit, offset *int, search, filter *string) ([]*models.Post, error) {
	var allItems []map[string]types.AttributeValue
	if search != nil && *search != "" {
		for {
			queryInput := &dynamodb.QueryInput{
				TableName:              aws.String(repo.TableName),
				FilterExpression:       aws.String("contains(title,:search)"),
				KeyConditionExpression: aws.String("pk = :pk and begins_with(sk, :sk)"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk":     &types.AttributeValueMemberS{Value: "posts"},
					":search": &types.AttributeValueMemberS{Value: *search},
					":sk":     &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s", *filter)},
				},
				ScanIndexForward: aws.Bool(false),
			}
			result, err := repo.Db.Query(ctx, queryInput)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, result.Items...)
			if limit != nil && *limit > 0 && offset != nil {
				if result.LastEvaluatedKey != nil && *limit+*offset >= len(allItems) {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			} else {
				if result.LastEvaluatedKey != nil {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			}
		}
	} else {
		for {
			queryInput := &dynamodb.QueryInput{
				TableName:              aws.String(repo.TableName),
				KeyConditionExpression: aws.String("pk = :pk and begins_with(sk, :sk)"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk": &types.AttributeValueMemberS{Value: "posts"},
					":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s", *filter)},
				},
				ScanIndexForward: aws.Bool(false),
			}
			result, err := repo.Db.Query(ctx, queryInput)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, result.Items...)
			if limit != nil && *limit > 0 && offset != nil {
				if result.LastEvaluatedKey != nil && *limit+*offset >= len(allItems) {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			} else {
				if result.LastEvaluatedKey != nil {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			}
		}
	}
	posts := make([]*models.Post, 0)
	for _, item := range allItems {
		var postWithSK models.PostSKFilter
		err := attributevalue.UnmarshalMap(item, &postWithSK)
		if err != nil {
			return nil, err
		}
		sk := strings.Split(postWithSK.SK, ":")
		post := &models.Post{
			Title:     postWithSK.Title,
			Content:   postWithSK.Content,
			Likes:     postWithSK.Likes,
			CreatedAt: postWithSK.CreatedAt,
			UId:       postWithSK.UId,
			PostId:    sk[3],
			Type:      config.Filter(sk[1]),
		}
		posts = append(posts, post)
	}
	if limit != nil && *limit > 0 && offset != nil && *limit+*offset <= len(posts) {
		return posts[*offset : *limit+1], nil
	}
	return posts, nil
}

func (repo *PostRepository) GetAllPosts(ctx context.Context, limit, offset *int, search *string) ([]*models.Post, error) {
	var allItems []map[string]types.AttributeValue
	if search != nil && *search != "" {
		for {
			queryInput := &dynamodb.QueryInput{
				TableName:              aws.String(repo.TableName),
				IndexName:              aws.String(repo.IndexName),
				FilterExpression:       aws.String("contains(title,:search)"),
				KeyConditionExpression: aws.String("pk = :pk"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk":     &types.AttributeValueMemberS{Value: "posts"},
					":search": &types.AttributeValueMemberS{Value: *search},
				},
				ScanIndexForward: aws.Bool(false),
			}
			result, err := repo.Db.Query(ctx, queryInput)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, result.Items...)
			if limit != nil && *limit > 0 && offset != nil {
				if result.LastEvaluatedKey != nil && *limit+*offset >= len(allItems) {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			} else {
				if result.LastEvaluatedKey != nil {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			}
		}
	} else {
		for {
			queryInput := &dynamodb.QueryInput{
				TableName:              aws.String(repo.TableName),
				IndexName:              aws.String(repo.IndexName),
				KeyConditionExpression: aws.String("pk = :pk"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk": &types.AttributeValueMemberS{Value: "posts"},
				},
				ScanIndexForward: aws.Bool(false),
			}
			result, err := repo.Db.Query(ctx, queryInput)
			if err != nil {
				return nil, err
			}
			allItems = append(allItems, result.Items...)
			if limit != nil && *limit > 0 && offset != nil {
				if result.LastEvaluatedKey != nil && *limit+*offset >= len(allItems) {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			} else {
				if result.LastEvaluatedKey != nil {
					queryInput.ExclusiveStartKey = result.LastEvaluatedKey
				} else {
					break
				}
			}
		}
	}
	posts := make([]*models.Post, 0)
	for _, item := range allItems {
		var postWithSK models.PostSKFilter
		err := attributevalue.UnmarshalMap(item, &postWithSK)
		if err != nil {
			return nil, err
		}
		sk := strings.Split(postWithSK.SK, ":")
		post := &models.Post{
			Title:     postWithSK.Title,
			Content:   postWithSK.Content,
			Likes:     postWithSK.Likes,
			CreatedAt: postWithSK.CreatedAt,
			UId:       postWithSK.UId,
			PostId:    sk[3],
			Type:      config.Filter(sk[1]),
		}
		posts = append(posts, post)
	}
	if limit != nil && *limit > 0 && offset != nil && *limit+*offset <= len(posts) {
		return posts[*offset : *limit+1], nil
	}
	return posts, nil
}

func (repo *PostRepository) DeletePost(ctx context.Context, filter, time, uId, pId string) error {
	input1 := &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "posts"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s:%s:%s", filter, time, pId)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: uId},
		},
		ConditionExpression: aws.String("user_id =:userId"),
		ReturnValues:        types.ReturnValueAllOld,
	}
	input2 := &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
			"sk": &types.AttributeValueMemberS{Value: "post:" + pId},
		},
	}
	result, deleteErr := repo.Db.DeleteItem(ctx, input1)
	if deleteErr != nil {
		return deleteErr
	}
	if result.Attributes != nil {
		var writeRequests []types.WriteRequest
		var pks []string
		var queryOutput *dynamodb.QueryOutput
		var err error
		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(repo.TableName),
			KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "post:" + pId},
				":sk": &types.AttributeValueMemberS{Value: "question:"},
			},
		}
		for {
			queryOutput, err = repo.Db.Query(ctx, queryInput)
			if err != nil {
				return err
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
				pks = append(pks, item["sk"].(*types.AttributeValueMemberS).Value)
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
		for _, pk := range pks {
			var writeRequests []types.WriteRequest
			var queryOutput *dynamodb.QueryOutput
			var err error
			queryInput := &dynamodb.QueryInput{
				TableName:              aws.String(repo.TableName),
				KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk": &types.AttributeValueMemberS{Value: pk},
					":sk": &types.AttributeValueMemberS{Value: "reply:"},
				},
			}
			for {
				queryOutput, err = repo.Db.Query(ctx, queryInput)
				if err != nil {
					return err
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
	}
	_, err := repo.Db.DeleteItem(ctx, input2)
	return err
}

func (repo *PostRepository) GetPostsByUId(ctx context.Context, uId string) ([]*models.Post, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk) "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "user:" + uId},
			":sk": &types.AttributeValueMemberS{Value: "post:"},
		},
		ScanIndexForward: aws.Bool(false),
	}
	result, err := repo.Db.Query(ctx, input)
	if err != nil {
		return nil, err
	}
	posts := make([]*models.Post, 0)
	for _, item := range result.Items {
		var postDB models.Post
		err := attributevalue.UnmarshalMap(item, &postDB)
		if err != nil {
			return nil, err
		}
		dtoPostId := strings.Split(postDB.PostId, ":")
		dtoUserId := strings.Split(postDB.UId, ":")
		post := &models.Post{
			Title:     postDB.Title,
			Content:   postDB.Content,
			Likes:     postDB.Likes,
			CreatedAt: postDB.CreatedAt,
			UId:       dtoUserId[1],
			PostId:    dtoPostId[1],
			Type:      postDB.Type,
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (repo *PostRepository) UpdatePost(ctx context.Context, uId string, post *models.Post) error {
	input1 := &dynamodb.UpdateItemInput{
		TableName:           aws.String(repo.TableName),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk) AND user_id = :userId"),
		Key: map[string]types.AttributeValue{
			"pk":     &types.AttributeValueMemberS{Value: "posts"},
			"sk":     &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s:%s:%s", post.Type, post.CreatedAt.Format("RFC3339"), post.PostId)},
			"userId": &types.AttributeValueMemberS{Value: uId},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":title":   &types.AttributeValueMemberS{Value: post.Title},
			":content": &types.AttributeValueMemberS{Value: post.Content},
		},
		UpdateExpression: aws.String("SET title =:title, content =:content"),
	}
	input2 := &dynamodb.UpdateItemInput{
		TableName:           aws.String(repo.TableName),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
			"sk": &types.AttributeValueMemberS{Value: "post:" + post.PostId},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":title":   &types.AttributeValueMemberS{Value: post.Title},
			":content": &types.AttributeValueMemberS{Value: post.Content},
		},
		UpdateExpression: aws.String("SET title =:title, content =:content"),
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var err error

	update := func(input *dynamodb.UpdateItemInput) {
		defer wg.Done()
		_, updateErr := repo.Db.UpdateItem(ctx, input)
		if updateErr != nil {
			mu.Lock()
			if err == nil {
				err = updateErr
			}
			mu.Unlock()
		}
	}
	wg.Add(2)
	go update(input1)
	go update(input2)
	wg.Wait()
	return err
}

func (repo *PostRepository) ToggleLike(ctx context.Context, uId, filter, pId string, time time.Time) (config.LikeStatus, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var err error
	update := func(input *dynamodb.UpdateItemInput) {
		defer wg.Done()
		_, updateErr := repo.Db.UpdateItem(ctx, input)
		if updateErr != nil {
			mu.Lock()
			if err == nil {
				err = updateErr
			}
			mu.Unlock()
		}
	}
	hasLiked, err := repo.HasUserLikedAPost(ctx, uId, pId)
	if err != nil {
		return "0", err
	}
	if hasLiked {
		input1 := &dynamodb.UpdateItemInput{
			TableName: aws.String(repo.TableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
				"sk": &types.AttributeValueMemberS{Value: "post:" + pId},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":likes": &types.AttributeValueMemberN{Value: "1"},
			},
			UpdateExpression: aws.String("SET likes = likes - :likes"),
		}
		input2 := &dynamodb.UpdateItemInput{
			TableName: aws.String(repo.TableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "posts"},
				"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s:%s:%s", filter, time.Format("RFC3339"), pId)},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":likes": &types.AttributeValueMemberN{Value: "1"},
			},
			UpdateExpression: aws.String("SET likes = likes - :likes"),
		}
		wg.Add(3)
		go update(input1)
		go update(input2)
		err = repo.deleteLikeEntry(ctx, uId, pId)
		wg.Wait()
		return config.NotLiked, err
	}
	input1 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
			"sk": &types.AttributeValueMemberS{Value: "post:" + pId},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":likes": &types.AttributeValueMemberN{Value: "1"},
		},
		UpdateExpression: aws.String("SET likes = likes + :likes"),
	}
	input2 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "posts"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("post:%s:%s:%s", filter, time.Format("RFC3339"), pId)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":likes": &types.AttributeValueMemberN{Value: "1"},
		},
		UpdateExpression: aws.String("SET likes = likes + :likes"),
	}
	wg.Add(3)
	go update(input1)
	go update(input2)
	err = repo.enterLikeEntry(ctx, uId, pId)
	wg.Wait()
	return config.Liked, err
}

func (repo *PostRepository) deleteLikeEntry(ctx context.Context, uId, pId string) error {
	_, err := repo.Db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "like:" + pId},
			"sk": &types.AttributeValueMemberS{Value: "user:" + uId},
		},
	})
	return err
}

func (repo *PostRepository) enterLikeEntry(ctx context.Context, uId, pId string) error {
	_, err := repo.Db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(repo.TableName),
		Item: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "like:" + pId},
			"sk": &types.AttributeValueMemberS{Value: "user:" + uId},
		},
	})
	return err
}

func (repo *PostRepository) HasUserLikedAPost(ctx context.Context, uId, pId string) (bool, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "like:" + pId},
			":sk": &types.AttributeValueMemberS{Value: "user:" + uId},
		},
	}
	result, err := repo.Db.Query(ctx, input)
	if err != nil {
		return false, err
	}
	return len(result.Items) > 0, err
}
