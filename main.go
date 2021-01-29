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
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/faabiosr/cachego/file"

	"github.com/fastwego/feishu"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var FeishuClient *feishu.Client
var FeishuConfig map[string]string

var Atm *feishu.DefaultAccessTokenManager

func init() {

	// 加载配置文件
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	FeishuConfig = map[string]string{
		"AppId":     viper.GetString("AppId"),
		"AppSecret": viper.GetString("AppSecret"),

		"VerificationToken": viper.GetString("VerificationToken"),
		"EncryptKey":        viper.GetString("EncryptKey"),
	}

	// 内部应用 tenant_access_token 管理器
	Atm = &feishu.DefaultAccessTokenManager{
		Id:    FeishuConfig["AppId"],
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

	router.POST("/api/feishu/callback", Callback)

	router.GET("/open-apis/meeting_room/building/list", func(c *gin.Context) {

		tenantAccessToken, err := Atm.GetAccessToken()
		if err != nil {
			log.Println(err)
			return
		}
		params := url.Values{}
		params.Add("page_size", "10")
		request, _ := http.NewRequest(http.MethodGet, feishu.ServerUrl+"/open-apis/meeting_room/building/list?"+params.Encode(), nil)
		resp, err := FeishuClient.Do(request, tenantAccessToken)
		log.Println(string(resp), err)
	})

	router.GET("/api/feishu/upload", Upload)

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
