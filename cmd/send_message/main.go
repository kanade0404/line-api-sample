package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type AccessTokenResponse struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
	Exp   int64  `json:"expires_in"`
	Id    string `json:"key_id"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	privateKeyFile, err := os.Open("private.key")
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(privateKeyFile)
	privateKeyContent, err := io.ReadAll(privateKeyFile)
	if err != nil {
		return err
	}
	privateKey, err := jwk.ParseKey(privateKeyContent)
	if err != nil {
		return err
	}
	aud := []string{"https://api.line.me/"}
	// JWTを構成する
	tok, err := jwt.NewBuilder().
		Subject(os.Getenv("CHANNEL_ID")).
		Issuer(os.Getenv("CHANNEL_ID")).
		Audience(aud).
		Expiration(time.Now().Add(30 * time.Minute)). // JWTの有効期間で最大30分を入れる
		Build()
	if err != nil {
		return err
	}

	if err := tok.Set("token_exp", 60*60*24*30); err != nil {
		return err
	} // token_expプロパティ、チャネルアクセストークンの有効期間を指定

	// JWTを発行する
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privateKey))
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Add("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	form.Add("client_assertion", string(signed))

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest(http.MethodPost, "https://api.line.me/oauth2/v2.1/token", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(res.Body)

	var accessTokenResponse AccessTokenResponse
	if err = json.NewDecoder(res.Body).Decode(&accessTokenResponse); err != nil {
		return err
	}
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), accessTokenResponse.Token)
	if err != nil {
		return err
	}
	if _, err = bot.PushMessage(os.Getenv("USER_ID"), linebot.NewTextMessage(`Hello, World🎉`)).Do(); err != nil {
		return err
	}
	return nil
}
