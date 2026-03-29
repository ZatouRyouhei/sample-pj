# AWS CDKを利用したサンプルプログラム

AWS CDK + Go + React で構築したサーバレスWebアプリケーション

## 構成

- **Backend**: Go Lambda + API Gateway + DynamoDB
- **Frontend**: React + CloudFront + S3
- **IaC**: AWS CDK (Go)

### 初回設定
cdk/cdk.jsonのdeveloperを自身の名前に変更してください。

### 初回デプロイ

```bash
# CDK Bootstrap（初回のみ）
cd cdk
cdk bootstrap

# 全体をビルド・デプロイ
build backend
build frontend
build cdk
```