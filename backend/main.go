package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aurora-dsql-connectors/go/pgx/dsql"
)

type AwsCdkTable struct {
	PK   string `dynamodbav:"PK"`
	SK   string `dynamodbav:"SK"`
	Name string `dynamodbav:"Name"`
}

// func hello() (string, error) {
// 	return "Hello AWS-CDK lambda!", nil
// }

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := request.Path
	switch path {
	case "/api/test":
		tableItem, err := GetItem("12345", "01")
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "ERROR:" + path,
			}, errors.New("値が取得できませんでした。")
		}
		body, err := json.Marshal(tableItem)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "ERROR: JSON変換失敗",
			}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}, nil
	case "/api/test2":
		ctx := context.Background()
		// コネクションプール作成
		pool, err := dsql.NewPool(ctx, dsql.Config{
			Host: os.Getenv("CLUSTER_ENDPOINT"),
		})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "ERROR: DB接続失敗",
			}, err
		}
		defer pool.Close()

		// SQL実行
		var name string
		err = pool.QueryRow(ctx, "select first_name from petclinic.owners where id = '4fd413f3-74e1-42d7-90d5-d9aaea9975c9'").Scan(&name)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "ERROR: SQL実行失敗",
			}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       name,
		}, nil
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "ERROR:" + path,
		}, errors.New("URLが不正です。" + path)
	}
}

// DynamoDBから値を取得する
func GetItem(pk string, sk string) (*AwsCdkTable, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)
	paramPk, err := attributevalue.Marshal(pk)
	if err != nil {
		return nil, err
	}
	paramSk, err := attributevalue.Marshal(sk)
	if err != nil {
		return nil, err
	}
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"PK": paramPk,
			"SK": paramSk,
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")), // テーブル名は環境変数から取得
	}
	response, err := client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	awsCdkTable := new(AwsCdkTable)
	err = attributevalue.UnmarshalMap(response.Item, awsCdkTable)
	if err != nil {
		return nil, err
	}
	return awsCdkTable, nil
}

func main() {
	// lambda.Start(hello)
	lambda.Start(handler)
}
