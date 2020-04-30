package ddb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	ddba "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/klauspost/compress/zstd"
	"github.com/orivej/e"
	"github.com/orivej/enlapin/bot/chatstate"
)

type DDBChatStateMap struct {
	*dynamodb.DynamoDB
	Table string
}
type DDBChatStateMapItem struct {
	ID      int64
	Expired time.Time `dynamodbav:",unixtime"` // Time at which DDB can erase this item.
	Locked  time.Time `dynamodbav:",unixtime"` // Time until which this item is locked by the handler, or 0.
	CS      []byte
}

func NewDDBChatStateMap(table string) DDBChatStateMap {
	opts := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess := session.Must(session.NewSessionWithOptions(opts))
	ddb := dynamodb.New(sess)
	return DDBChatStateMap{DynamoDB: ddb, Table: table}
}

func (ddb DDBChatStateMap) Get(chatID int64) (*chatstate.ChatState, func()) {
	var result *dynamodb.UpdateItemOutput
	var err error
	for {
		now := time.Now()
		params := struct {
			Now     time.Time `dynamodbav:":Now,unixtime"`
			Expired time.Time `dynamodbav:":Expired,unixtime"`
			Locked  time.Time `dynamodbav:":Locked,unixtime"`
		}{now, now.Add(chatstate.Lifetime), now.Add(chatstate.Locktime)}
		cexpr := "attribute_not_exists(ID) OR Locked < :Now"
		uexpr := "set Expired = :Expired, Locked = :Locked"
		result, err = ddb.UpdateItem(&dynamodb.UpdateItemInput{
			TableName:                 &ddb.Table,
			Key:                       MarshalKey(chatID),
			ConditionExpression:       &cexpr,
			UpdateExpression:          &uexpr,
			ExpressionAttributeValues: MarshalMap(&params),
			ReturnValues:              aws.String(dynamodb.ReturnValueAllNew),
		})
		if err == nil {
			break
		}
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			fmt.Println(result)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		e.Print(err)
		return nil, nil
	}

	item := &DDBChatStateMapItem{}
	err = ddba.UnmarshalMap(result.Attributes, item)
	e.Exit(err)
	cs := &chatstate.ChatState{}
	if len(item.CS) > 0 {
		r, err := zstd.NewReader(bytes.NewBuffer(item.CS))
		e.Exit(err)
		err = json.NewDecoder(r).Decode(cs)
		e.Exit(err)
		r.Close()
	}
	return cs, func() { ddb.Unlock(chatID, cs) }
}

func (ddb DDBChatStateMap) Unlock(chatID int64, cs *chatstate.ChatState) {
	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf)
	e.Exit(err)
	err = json.NewEncoder(w).Encode(cs)
	e.Exit(err)
	err = w.Close()
	e.Exit(err)
	now := time.Now()
	params := struct {
		Now     time.Time `dynamodbav:":Now,unixtime"`
		Expired time.Time `dynamodbav:":Expired,unixtime"`
		Locked  int       `dynamodbav:":Locked"`
		CS      []byte    `dynamodbav:":CS"`
	}{now, now.Add(chatstate.Lifetime), 0, buf.Bytes()}
	cexpr := "attribute_not_exists(ID) OR Locked >= :Now"
	uexpr := "set Expired = :Expired, Locked = :Locked, CS = :CS"
	_, err = ddb.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 &ddb.Table,
		Key:                       MarshalKey(chatID),
		ConditionExpression:       &cexpr,
		UpdateExpression:          &uexpr,
		ExpressionAttributeValues: MarshalMap(&params),
	})
	e.Print(err)
}

func MarshalKey(chatID int64) map[string]*dynamodb.AttributeValue {
	return MarshalMap(struct{ ID int64 }{chatID})
}

func MarshalMap(in interface{}) map[string]*dynamodb.AttributeValue {
	attr, err := ddba.MarshalMap(in)
	e.Exit(err)
	return attr
}
