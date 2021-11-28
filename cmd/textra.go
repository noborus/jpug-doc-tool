package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// みんなの自動翻訳 https://mt-auto-minhon-mlt.ucri.jgn-x.jp/
// APIを使用した翻訳

const APIURL = "https://mt-auto-minhon-mlt.ucri.jgn-x.jp/"

type MTClient struct {
	client       *http.Client
	token        *oauth2.Token
	ClientID     string
	ClientSecret string
	Name         string
	APIName      string
	APIParam     string
}

type MTResult struct {
	Resultset struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Request struct {
			URL   string `json:"url"`
			Text  string `json:"text"`
			Split int    `json:"split"`
			Data  string `json:"data"`
		} `json:"request"`
		Result struct {
			Text        string      `json:"text"`
			Information interface{} `json:"information"`
		} `json:"result"`
	} `json:"resultset"`
}

func apiClient(c apiConfig) MTClient {
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     Config.ClientID,
		ClientSecret: Config.ClientSecret,
		TokenURL:     APIURL + "oauth2/token.php",
	}

	client := conf.Client(ctx)
	token, err := conf.Token(ctx)
	if err != nil {
		log.Fatal(err)
	}
	api := MTClient{}
	api.client = client
	api.token = token
	api.ClientID = Config.ClientID
	api.ClientSecret = Config.ClientSecret
	api.Name = Config.Name
	api.APIName = Config.APIAutoTranslate
	api.APIParam = Config.APIAutoTranslateType
	return api
}

func (a MTClient) textraTranslate(enstr string) string {
	values := url.Values{
		"access_token": []string{a.token.AccessToken},
		"key":          []string{a.ClientID},
		"api_name":     []string{a.APIName},
		"api_param":    []string{a.APIParam},
		"name":         []string{a.Name},
		"type":         []string{"json"},
		"text":         []string{enstr},
	}

	resp, err := a.client.PostForm(APIURL+"api/", values)
	if err != nil {
		log.Fatal(err)
	}
	s, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	data := new(MTResult)
	if err := json.Unmarshal(s, data); err != nil {
		fmt.Println(string(s))
		log.Fatal(err)
	}
	return data.Resultset.Result.Text
}
