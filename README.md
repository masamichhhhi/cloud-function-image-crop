## image-crop-function
cloud functionで画像のクロップをする

### ローカル実行
```bash
go build function.go
go run cmd/main.go
```

### TODO
- gifがスローモーションになる
- デプロイしてみる


### デプロイ
```bash
gcloud functions deploy image-crop-function --entry-point CropImage --set-env-vars ENVIRONMENT=production
```