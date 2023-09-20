# LINE Messaging API with Golang

## Usage

### 1. 管理画面から設定する

[LINE Developersのドキュメント](https://developers.line.biz/ja/docs/messaging-api/getting-started/)より手順に沿って設定をする。

秘密鍵はブラウザで作るのが楽です。
private.keyというファイル名で保存してください。

### 2. Set environment variables

```shell
cp .envrc.sample .envrc
direnv allow
```

### 3. Run

```shell
go run cmd/send_message/main.go
```
