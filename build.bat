@echo off
chcp 65001
rem バックエンドビルド
if "%1" == "backend" goto build_backend
if "%1" == "all"     goto build_backend
goto end_backend

:build_backend
echo START build_backend
cd /d %~dp0
cd backend
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o bootstrap .
echo END build_backend
:end_backend

rem フロントエンドビルド
if "%1" == "frontend" goto build_frontend
if "%1" == "all"      goto build_frontend
goto end_frontend

:build_frontend
echo START build_frontend
cd /d %~dp0
cd frontend
call npm run build
echo END build_frontend
:end_frontend

rem マイグレーションビルド
if "%1" == "migration" goto build_migration
if "%1" == "all"     goto build_migration
goto end_backend

:build_migration
echo START build_migration
cd /d %~dp0
cd migration
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o bootstrap .
echo END build_migration
:build_migration

rem AWSへデプロイ
if "%1" == "cdk" goto cdk_deploy
if "%1" == "all" goto cdk_deploy
goto end_deploy

:cdk_deploy
echo START cdk_deploy
cd /d %~dp0
cd cdk
call cdk deploy
echo END cdk_deploy
:end_deploy

cd /d %~dp0
pause