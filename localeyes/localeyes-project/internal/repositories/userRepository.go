package repositories

import (
	"context"
	"fmt"
	"localeyes/internal/models"
	"localeyes/utils"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserRepository struct {
	Db        *dynamodb.Client
	TableName string
}

func NewNoSQLUserRepository(db *dynamodb.Client) *UserRepository {
	return &UserRepository{
		db,
		os.Getenv("TABLE_NAME"),
	}
}

func (repo *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	users := make([]map[string]types.AttributeValue, 0, 3)
	userSKEmail := &models.UserSKEmail{
		PK:          "users",
		UId:         user.UId,
		Username:    user.Username,
		Password:    user.Password,
		Email:       "email:" + user.Email,
		Tag:         user.Tag,
		City:        user.City,
		IsActive:    user.IsActive,
		DwellingAge: user.DwellingAge,
	}
	userSKUsername := &models.UserSKUsername{
		PK:          "users",
		UId:         user.UId,
		Username:    "username:" + user.Username,
		Password:    user.Password,
		Email:       user.Email,
		Tag:         user.Tag,
		City:        user.City,
		IsActive:    user.IsActive,
		DwellingAge: user.DwellingAge,
	}
	userPKId := &models.User{
		UId:         "user:" + user.UId,
		Username:    user.Username,
		Password:    user.Password,
		Email:       user.Email,
		Tag:         user.Tag,
		City:        user.City,
		DwellingAge: user.DwellingAge,
	}
	userSKEmailAv, err := attributevalue.MarshalMap(userSKEmail)
	//userSKEmailAv["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

	userSKUsernameAv, err := attributevalue.MarshalMap(userSKUsername)
	//userSKUsernameAv["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

	av, err := attributevalue.MarshalMap(userPKId)
	if user.IsActive {
		av["sk"] = &types.AttributeValueMemberS{Value: "true"}
	} else {
		av["sk"] = &types.AttributeValueMemberS{Value: "false"}
	}
	//av["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

	users = append(users, userSKEmailAv, userSKUsernameAv, av)
	writeRequests := make([]types.WriteRequest, len(users))

	for i, user := range users {
		writeRequests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: user,
			},
		}
	}

	_, err = repo.Db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			repo.TableName: writeRequests,
		},
	})

	return err
}

func (repo *UserRepository) FetchUserByEmail(ctx context.Context, email string) (*models.UserSKEmail, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("email:%s", email)},
		},
	}

	result, err := repo.Db.GetItem(ctx, input)
	if err != nil {
		return &models.UserSKEmail{}, err
	}

	if result.Item == nil {
		return &models.UserSKEmail{}, utils.NoUser
	}

	var dbUser models.UserSKEmail
	if err := attributevalue.UnmarshalMap(result.Item, &dbUser); err != nil {
		return &models.UserSKEmail{}, err
	}
	dtoEmail := strings.Split(dbUser.Email, ":")
	dbUser.Email = dtoEmail[1]
	return &dbUser, nil
}

func (repo *UserRepository) FetchUserByUsername(ctx context.Context, username string) (*models.UserSKUsername, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("username:%s", username)},
		},
	}

	result, err := repo.Db.GetItem(ctx, input)
	if err != nil {
		return &models.UserSKUsername{}, err
	}

	if result.Item == nil {
		return &models.UserSKUsername{}, utils.NoUser
	}

	var dbUser models.UserSKUsername
	if err := attributevalue.UnmarshalMap(result.Item, &dbUser); err != nil {
		return &models.UserSKUsername{}, err
	}
	dtoUsername := strings.Split(dbUser.Username, ":")
	dbUser.Username = dtoUsername[1]
	return &dbUser, nil
}

func (repo *UserRepository) FetchUserById(ctx context.Context, uid string, isUserActive bool) (*models.User, error) {
	var input *dynamodb.GetItemInput
	input = &dynamodb.GetItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", uid)},
			"sk": &types.AttributeValueMemberS{Value: "true"},
		},
	}

	if !isUserActive {
		key := map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", uid)},
			"sk": &types.AttributeValueMemberS{Value: "false"},
		}
		input.Key = key
	}

	result, err := repo.Db.GetItem(ctx, input)
	if err != nil {
		return &models.User{}, err
	}

	if result.Item == nil {
		return &models.User{}, utils.NoUser
	}

	var dbUser models.UserWithStringStatus
	if err := attributevalue.UnmarshalMap(result.Item, &dbUser); err != nil {
		return &models.User{}, err
	}
	var user = &models.User{
		Username:    dbUser.Username,
		UId:         dbUser.UId,
		City:        dbUser.City,
		DwellingAge: dbUser.DwellingAge,
		Password:    dbUser.Password,
		Email:       dbUser.Email,
		Tag:         dbUser.Tag,
	}
	dtoId := strings.Split(dbUser.UId, ":")
	user.UId = dtoId[1]
	if dbUser.IsActive == "true" {
		user.IsActive = true
	} else {
		user.IsActive = false
	}
	return user, nil
}

func (repo *UserRepository) UpdateUserById(ctx context.Context, user *models.User) error {
	input1 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", user.UId)},
			"sk": &types.AttributeValueMemberS{Value: "true"},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city":         &types.AttributeValueMemberS{Value: user.City},
			":password":     &types.AttributeValueMemberS{Value: user.Password},
			":dwelling_age": &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	}
	input2 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("email:%s", user.Email)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city":         &types.AttributeValueMemberS{Value: user.City},
			":password":     &types.AttributeValueMemberS{Value: user.Password},
			":dwelling_age": &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	}
	input3 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("username:%s", user.Username)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city":         &types.AttributeValueMemberS{Value: user.City},
			":password":     &types.AttributeValueMemberS{Value: user.Password},
			":dwelling_age": &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var upErr error

	update := func(input *dynamodb.UpdateItemInput) {
		defer wg.Done()
		_, updateErr := repo.Db.UpdateItem(ctx, input)
		if updateErr != nil {
			mu.Lock()
			if upErr == nil {
				upErr = updateErr
			}
			mu.Unlock()
		}
	}
	wg.Add(3)
	go update(input1)
	go update(input2)
	go update(input3)
	wg.Wait()
	return upErr
}

func (repo *UserRepository) ToggleUserActiveStatus(ctx context.Context, user *models.User) error {
	var activeStatus string
	if user.IsActive {
		activeStatus = "true"
	} else {
		activeStatus = "false"
	}
	var input1 *types.DeleteRequest

	if ctx.Value("Role").(string) == "admin" {
		input1 = &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", user.UId)},
				"sk": &types.AttributeValueMemberS{Value: "false"},
			},
		}
	} else {
		input1 = &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", user.UId)},
				"sk": &types.AttributeValueMemberS{Value: "true"},
			},
		}
	}
	userPKId := &models.User{
		UId:         "user:" + user.UId,
		Username:    user.Username,
		Password:    user.Password,
		Email:       user.Email,
		Tag:         user.Tag,
		City:        user.City,
		DwellingAge: user.DwellingAge,
	}
	userAv, err := attributevalue.MarshalMap(userPKId)
	if err != nil {
		return err
	}
	userAv["sk"] = &types.AttributeValueMemberS{Value: activeStatus}
	input2 := &types.PutRequest{
		Item: userAv,
	}
	writeRequests := []types.WriteRequest{
		{
			DeleteRequest: input1,
		},
		{
			PutRequest: input2,
		},
	}
	_, err = repo.Db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			repo.TableName: writeRequests,
		},
	})
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var upErr error

	update := func(input *dynamodb.UpdateItemInput) {
		defer wg.Done()
		_, updateErr := repo.Db.UpdateItem(ctx, input)
		if updateErr != nil {
			mu.Lock()
			if upErr == nil {
				upErr = updateErr
			}
			mu.Unlock()
		}
	}
	input3 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("email:%s", user.Email)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city":         &types.AttributeValueMemberS{Value: user.City},
			":password":     &types.AttributeValueMemberS{Value: user.Password},
			":dwelling_age": &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)},
			":active":       &types.AttributeValueMemberBOOL{Value: user.IsActive},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age, is_active = :active")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	}
	input4 := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "users"},
			"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("username:%s", user.Username)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city":         &types.AttributeValueMemberS{Value: user.City},
			":password":     &types.AttributeValueMemberS{Value: user.Password},
			":dwelling_age": &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)},
			":active":       &types.AttributeValueMemberBOOL{Value: user.IsActive},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age, is_active = :active")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	}
	wg.Add(2)
	go update(input3)
	go update(input4)
	wg.Wait()
	return upErr
}

func (repo *UserRepository) FetchNotifications(ctx context.Context, uId string) ([]*models.Notification, error) {
	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		FilterExpression:       aws.String("user_id <> :userId"),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "notifications"},
			":userId": &types.AttributeValueMemberS{Value: uId},
		},
	})
	if err != nil {
		return nil, err
	}
	var notifications []*models.Notification
	for _, item := range result.Items {
		var notification models.Notification
		currentTime := time.Now().Unix()
		err := attributevalue.UnmarshalMap(item, &notification)
		if err != nil {
			return nil, err
		}
		if currentTime <= notification.TTl {
			notifications = append(notifications, &notification)
		}
	}
	return notifications, nil
}

//func (repo *UserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
//	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
//		TableName:              aws.String(repo.TableName),
//		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
//		ExpressionAttributeValues: map[string]types.AttributeValue{
//			":pk": &types.AttributeValueMemberS{Value: "users"},
//			":sk": &types.AttributeValueMemberS{Value: "username:"},
//		},
//	})
//	if err != nil {
//		return nil, err
//	}
//	var users []*models.User
//	for _, user := range result.Items {
//		var userModel models.UserSKUsername
//		err := attributevalue.UnmarshalMap(user, &userModel)
//		if err != nil {
//			return nil, err
//		}
//		userNew := &models.User{
//			Username:    strings.Split(userModel.Username, ":")[1],
//			UId:         userModel.UId,
//			Email:       userModel.Email,
//			DwellingAge: userModel.DwellingAge,
//			Password:    userModel.Password,
//			City:        userModel.City,
//			IsActive:    userModel.IsActive,
//			Tag:         userModel.Tag,
//		}
//		users = append(users, userNew)
//	}
//	return users, nil
//}

func (repo *UserRepository) GetAllUsers(ctx context.Context, params models.GetUsersParams) ([]*models.User, error) {
	if params.Limit == 0 {
		params.Limit = 10
	}

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "users"},
			":sk": &types.AttributeValueMemberS{Value: "email:"},
		},
		Limit: aws.Int32(params.Limit),
	}

	// Add search filter if search parameter is provided
	if params.Search != "" {
		queryInput.FilterExpression = aws.String("contains(username, :search)")
		queryInput.ExpressionAttributeValues[":search"] = &types.AttributeValueMemberS{Value: params.Search}
	}

	// Handle offset using ExclusiveStartKey
	if params.Offset > 0 {
		// First, we need to get the key at offset position
		offsetInput := *queryInput
		offsetInput.Limit = aws.Int32(params.Offset)

		offsetResult, err := repo.Db.Query(ctx, &offsetInput)
		if err != nil {
			return nil, err
		}

		if len(offsetResult.Items) > 0 {
			queryInput.ExclusiveStartKey = offsetResult.LastEvaluatedKey
		} else {
			return []*models.User{}, nil // Return empty if offset is beyond available data
		}
	}

	result, err := repo.Db.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	var users []*models.User
	for _, user := range result.Items {
		var userModel models.UserSKEmail
		err := attributevalue.UnmarshalMap(user, &userModel)
		if err != nil {
			return nil, err
		}

		userNew := &models.User{
			Username:    userModel.Username,
			UId:         userModel.UId,
			Email:       strings.Split(userModel.Email, ":")[1],
			DwellingAge: userModel.DwellingAge,
			Password:    userModel.Password,
			City:        userModel.City,
			IsActive:    userModel.IsActive,
			Tag:         userModel.Tag,
		}
		users = append(users, userNew)
	}
	return users, nil
}

func (repo *UserRepository) DeleteUser(ctx context.Context, uId, username, email string) error {
	input1 := types.WriteRequest{
		DeleteRequest: &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "users"},
				"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("username:%s", username)},
			},
		},
	}
	input2 := types.WriteRequest{
		DeleteRequest: &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "users"},
				"sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("email:%s", email)},
			},
		},
	}
	input3 := types.WriteRequest{
		DeleteRequest: &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
				"sk": &types.AttributeValueMemberS{Value: "true"},
			},
		},
	}
	input4 := types.WriteRequest{
		DeleteRequest: &types.DeleteRequest{
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "user:" + uId},
				"sk": &types.AttributeValueMemberS{Value: "false"},
			},
		},
	}
	writeRequests := []types.WriteRequest{input1, input2, input3, input4}
	_, err := repo.Db.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			repo.TableName: writeRequests,
		},
	})
	return err
}
