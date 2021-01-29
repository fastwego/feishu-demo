// Copyright 2021 FastWeGo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fastwego/feishu"

	"github.com/faabiosr/cachego/file"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var FeishuClient *feishu.Client
var FeishuConfig map[string]string

var AppAtm *feishu.DefaultAccessTokenManager
var TenantAtm *feishu.DefaultAccessTokenManager

func init() {

	// 加载配置文件
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	FeishuConfig = map[string]string{
		"AppId":     viper.GetString("AppId"),
		"AppSecret": viper.GetString("AppSecret"),

		"VerificationToken": viper.GetString("VerificationToken"),
		"EncryptKey":        viper.GetString("EncryptKey"),

		"RedirectUri": viper.GetString("RedirectUri"),
	}
	// 内部应用 app_access_token 管理器
	AppAtm = &feishu.DefaultAccessTokenManager{
		Id:    "login-app:app_access_token:" + FeishuConfig["AppId"],
		Cache: file.New(os.TempDir()),
		GetRefreshRequestFunc: func() *http.Request {
			payload := `{
				"app_id":"` + FeishuConfig["AppId"] + `",
				"app_secret":"` + FeishuConfig["AppSecret"] + `"
			}`
			req, _ := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/auth/v3/app_access_token/internal/", strings.NewReader(payload))

			return req
		},
	}

	// 内部应用 tenant_access_token 管理器
	TenantAtm = &feishu.DefaultAccessTokenManager{
		Id:    "login-app:tenant_access_token:" + FeishuConfig["AppId"],
		Cache: file.New(os.TempDir()),
		GetRefreshRequestFunc: func() *http.Request {
			payload := `{
				"app_id":"` + FeishuConfig["AppId"] + `",
				"app_secret":"` + FeishuConfig["AppSecret"] + `"
			}`
			req, _ := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/auth/v3/tenant_access_token/internal/", strings.NewReader(payload))

			return req
		},
	}

	FeishuClient = feishu.NewClient()

}

func main() {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Session
	gob.Register(User{})
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("gosession", store))

	router.GET("/", Index)
	router.GET("/oauth", Oauth)

	svr := &http.Server{
		Addr:    viper.GetString("LISTEN"),
		Handler: router,
	}

	go func() {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	timeout := time.Duration(5) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

type User struct {
	AccessToken      string `json:"access_token"`
	AvatarURL        string `json:"avatar_url"`
	AvatarThumb      string `json:"avatar_thumb"`
	AvatarMiddle     string `json:"avatar_middle"`
	AvatarBig        string `json:"avatar_big"`
	ExpiresIn        int    `json:"expires_in"`
	Name             string `json:"name"`
	EnName           string `json:"en_name"`
	OpenID           string `json:"open_id"`
	UnionID          string `json:"union_id"`
	Email            string `json:"email"`
	UserID           string `json:"user_id"`
	Mobile           string `json:"mobile"`
	TenantKey        string `json:"tenant_key"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`

	Message string `json:"message"`
}

func Index(c *gin.Context) {

	session := sessions.Default(c)
	user := session.Get("user")

	loginUser, ok := user.(User)
	if !ok { // 未登录
		params := url.Values{}
		params.Add("redirect_uri", FeishuConfig["RedirectUri"])
		params.Add("app_id", FeishuConfig["AppId"])
		params.Add("state", "STATE")
		uri := feishu.ServerUrl + "/open-apis/authen/v1/index?" + params.Encode()

		log.Println("redirect ", uri)
		c.Redirect(302, uri)
		return
	}

	join := c.Query("join")
	if len(join) > 0 {
		// 发送 报名信息
		replyTextMsg := struct {
			OpenId  string `json:"open_id"`
			MsgType string `json:"msg_type"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		}{
			OpenId:  loginUser.OpenID,
			MsgType: "text",
			Content: struct {
				Text string `json:"text"`
			}{Text: "报名成功"},
		}

		data, err := json.Marshal(replyTextMsg)
		if err != nil {
			fmt.Println(err)
			return
		}
		tenantAccessToken, err := TenantAtm.GetAccessToken()
		if err != nil {
			log.Println(err)
			return
		}

		request, err := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/message/v4/send/", bytes.NewReader(data))
		resp, err := FeishuClient.Do(request, tenantAccessToken)
		fmt.Println(string(resp), err)

		loginUser.Message = "报名成功~"
	}

	t1, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	t1.Execute(c.Writer, loginUser)
}

func Oauth(c *gin.Context) {

	code := c.Query("code")
	fmt.Println("code ", code)

	if len(code) == 0 {
		return
	}
	appAccessToken, err := AppAtm.GetAccessToken()
	if err != nil {
		log.Println(err)
		return
	}

	// 获取用户身份
	params := url.Values{}
	params.Add("code", code)
	payload := `{
    "app_access_token":"` + appAccessToken + `",
    "grant_type":"authorization_code",
    "code":"` + code + `"
}`
	req, _ := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/authen/v1/access_token", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	userInfo, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	defer response.Body.Close()

	log.Println(string(userInfo), err)
	if err != nil {
		return
	}

	UserInfoResp := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data User   `json:"data"`
	}{}

	err = json.Unmarshal(userInfo, &UserInfoResp)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 记录 Session
	gob.Register(User{})
	session := sessions.Default(c)
	session.Set("user", UserInfoResp.Data)
	fmt.Println(UserInfoResp)
	err = session.Save()

	if err != nil {
		fmt.Println(err)
		return
	}

	// 登录成功
	c.Redirect(302, "/")
}
