package repositories

import (
	"context"
	"errors"
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
		PK:       "users",
		UId:      user.UId,
		Username: user.Username,
		Password: user.Password,
		Email:    "email:" + user.Email,
		Tag:      user.Tag,
		City:     user.City,
		IsActive: user.IsActive,
	}
	userSKUsername := &models.UserSKUsername{
		PK:       "users",
		UId:      user.UId,
		Username: "username:" + user.Username,
		Password: user.Password,
		Email:    user.Email,
		Tag:      user.Tag,
		City:     user.City,
		IsActive: user.IsActive,
	}
	userPKId := &models.User{
		UId:      "user:" + user.UId,
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
		Tag:      user.Tag,
		City:     user.City,
	}
	userSKEmailAv, err := attributevalue.MarshalMap(userSKEmail)
	userSKEmailAv["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

	userSKUsernameAv, err := attributevalue.MarshalMap(userSKUsername)
	userSKUsernameAv["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

	av, err := attributevalue.MarshalMap(userPKId)
	if user.IsActive {
		av["sk"] = &types.AttributeValueMemberS{Value: "true"}
	} else {
		av["sk"] = &types.AttributeValueMemberS{Value: "false"}
	}
	av["dwelling_age"] = &types.AttributeValueMemberN{Value: strconv.FormatFloat(user.DwellingAge, 'f', -1, 64)}

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
		return &models.UserSKEmail{}, errors.New("user not found")
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
		return &models.UserSKUsername{}, errors.New("user not found")
	}

	var dbUser models.UserSKUsername
	if err := attributevalue.UnmarshalMap(result.Item, &dbUser); err != nil {
		return &models.UserSKUsername{}, err
	}
	dtoUsername := strings.Split(dbUser.Username, ":")
	dbUser.Username = dtoUsername[1]
	return &dbUser, nil
}

func (repo *UserRepository) FetchUserById(ctx context.Context, uid string) (*models.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(repo.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", uid)},
			"sk": &types.AttributeValueMemberS{Value: "true"},
		},
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
	var activeStatus string
	if user.IsActive {
		activeStatus = "true"
	} else {
		activeStatus = "false"
	}
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
			":is_active":    &types.AttributeValueMemberS{Value: activeStatus},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age, is_active =:is_active")),
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
			":is_active":    &types.AttributeValueMemberS{Value: activeStatus},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age, is_active =:is_active")),
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
			":is_active":    &types.AttributeValueMemberS{Value: activeStatus},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET city =:city, password =:password , dwelling_age =:dwelling_age, is_active =:is_active")),
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
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
	wg.Add(3)
	go update(input1)
	go update(input2)
	go update(input3)
	wg.Wait()
	return err
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

func (repo *UserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	result, err := repo.Db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "users"},
			":sk": &types.AttributeValueMemberS{Value: "username:"},
		},
	})
	if err != nil {
		return nil, err
	}
	var users []*models.User
	for _, user := range result.Items {
		var userModel models.UserSKUsername
		err := attributevalue.UnmarshalMap(user, userModel)
		if err != nil {
			return nil, err
		}
		userNew := &models.User{
			Username:    strings.Split(userModel.Username, ":")[1],
			UId:         userModel.UId,
			Email:       userModel.Email,
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
