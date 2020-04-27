package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	ddba "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/orivej/OddHareGameBot/bot/chatstate"
	"github.com/orivej/e"
)

type DDBChatStateMap struct {
	*dynamodb.DynamoDB
	Table string
}
type DDBChatStateMapItemKey struct{ ID int64 }
type DDBChatStateMapItem struct {
	ID     int64
	TTL    ddba.UnixTime // Time at which DDB can erase this item.
	Locked ddba.UnixTime // Time until which this item is locked by the handler, or 0.
	Data   []byte
}

func NewDDBChatStateMap(table string) DDBChatStateMap {
	opts := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess := session.Must(session.NewSessionWithOptions(opts))
	ddb := dynamodb.New(sess)
	return DDBChatStateMap{DynamoDB: ddb, Table: table}
}

func (ddb DDBChatStateMap) Get(chatID int64) (*chatstate.ChatState, func()) {
	result, err := ddb.GetItem(&dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(ddb.Table),
		Key:            MarshalKey(chatID),
	})
	e.Exit(err)
	item := DDBChatStateMapItem{}
	err = ddba.UnmarshalMap(result.Item, &item)
	av, err := ddba.MarshalMap(item)
	e.Exit(err)
	input := &dynamodb.PutItemInput{
		TableName: aws.String(ddb.Table),
		Item:      av,
	}
	_, err = ddb.PutItem(input)
	e.Exit(err)
	return nil, func() {}
}

func MarshalKey(chatID int64) map[string]*dynamodb.AttributeValue {
	return MarshalMap(DDBChatStateMapItemKey{chatID})
}

func MarshalMap(in interface{}) map[string]*dynamodb.AttributeValue {
	attr, err := ddba.MarshalMap(in)
	e.Exit(err)
	return attr
}
