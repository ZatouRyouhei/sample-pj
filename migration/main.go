package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aurora-dsql-connectors/go/pgx/dsql"
)

func handler(ctx context.Context, event cfn.Event) (string, map[string]any, error) {
	// Deleteイベントは何もしない
	if event.RequestType == cfn.RequestDelete {
		return "db-migration", nil, nil
	}

	// コネクションプール作成
	pool, err := dsql.NewPool(ctx, dsql.Config{
		Host: os.Getenv("CLUSTER_ENDPOINT"),
	})
	if err != nil {
		return "db-migration", nil, err
	}
	defer pool.Close()

	// テーブル作成
	migrations := []string{
		// スキーマ作成
		`CREATE SCHEMA IF NOT EXISTS myapp`,
		// スキーマ内にテーブル作成
		`CREATE TABLE IF NOT EXISTS myapp.users (
			id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS myapp.items (
			id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			title   TEXT NOT NULL
		)`,
	}

	for _, m := range migrations {
		_, err := pool.Exec(ctx, m)
		if err != nil {
			return "db-migration", nil, err
		}
	}

	return "db-migration", map[string]any{"Status": "OK"}, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}
