package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	uuid "github.com/satori/go.uuid"
)

const tableName = "sns2imTelegram"

// Item for DynamoDB Item data type
type Item struct {
	ID        string `json:"_id"`
	ChatID    string `json:"chatId"`
	CreatedAt string `json:"createdAt"`
}

// DynamoDBHandler ...
type DynamoDBHandler struct {
	svc *dynamodb.DynamoDB
}

// NewDdbHandler ...
func NewDdbHandler() (*DynamoDBHandler, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1")},
	)
	if err != nil {
		log.Println("failed to initializa the DynamoDB Handler")
		return &DynamoDBHandler{}, err
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	return &DynamoDBHandler{
		svc,
	}, err
}

// GetItem ...
func (db *DynamoDBHandler) GetItem(chatID int32) (Item, error) {
	strChatID := strconv.FormatInt(int64(chatID), 10)
	result, err := db.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"chatId": {
				S: aws.String(strChatID),
			},
		},
	})

	if err != nil {
		log.Println(err.Error())
		return Item{}, err
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	log.Println("got item")
	jsonstr, _ := json.MarshalIndent(item, "", "  ")
	log.Println(string(jsonstr))
	return item, err
}

// IsValidChatID ...
func (db *DynamoDBHandler) IsValidChatID(chatID int32, UUID string) bool {
	item, err := db.GetItem(chatID)
	if err != nil {
		log.Printf("got error: %v", err)
		return false
	}
	if item.ID == UUID {
		log.Printf("item.ID = UUID")
		return true
	}
	log.Printf("item.ID != UUID")
	return false
}

// CreateItem ...
func (db *DynamoDBHandler) CreateItem(chatID int32) (string, string, error) {
	u2 := uuid.NewV4()
	item := Item{
		ID:        u2.String(),
		ChatID:    strconv.FormatInt(int64(chatID), 10),
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	av, err := dynamodbattribute.MarshalMap(item)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = db.svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return "", "", err
	}

	fmt.Printf("CreateItem with chatID:%v completed", chatID)
	return item.ChatID, item.ID, err
}
