@echo off
chcp 65001 > nul
if %1 == backend (
    echo バックエンドビルド開始
    cd /d %~dp0
    cd backend
    set GOOS=linux
    set GOARCH=amd64
    set CGO_ENABLED=0
    go build -o bootstrap .
    echo バックエンドビルド完了
)

if %1 == frontend (
    echo フロントエンドビルド開始
    cd /d %~dp0
    cd frontend
    call npm run build
    echo フロントエンドビルド完了
)

if %1 == cdk (
    echo AWSにデプロイ開始
    cd /d %~dp0
    cd cdk
    call cdk deploy
    echo AWSにデプロイ完了
)

cd /d %~dp0
pause