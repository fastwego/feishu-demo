package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fastwego/feishu/apis/capabilities/approval"

	"github.com/fastwego/feishu/apis/capabilities/calendar"
	"github.com/fastwego/feishu/apis/capabilities/meeting_room"

	"github.com/fastwego/feishu"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var App *feishu.App
var PublicApp *feishu.PublicApp

func init() {
	// 加载配置文件
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	App = feishu.NewApp(feishu.AppConfig{
		AppId:             viper.GetString("APPID"),
		AppSecret:         viper.GetString("SECRET"),
		VerificationToken: viper.GetString("TOKEN"),
		EncryptKey:        viper.GetString("AESKey"),
	})

	PublicApp = feishu.NewPublicApp(feishu.AppConfig{
		AppId:             viper.GetString("APPID"),
		AppSecret:         viper.GetString("SECRET"),
		VerificationToken: viper.GetString("TOKEN"),
		EncryptKey:        viper.GetString("AESKey"),
	}, "helloworld")

}

func main() {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("/api/feishu/callback", Callback)

	router.GET("/api/feishu/demo", func(c *gin.Context) {
		params := url.Values{}
		list, err := meeting_room.BuildingList(PublicApp.App, params)
		fmt.Println(string(list), err)

		params = url.Values{}
		params.Add("calendarId", "10086")
		resp, err := calendar.GetCalendarById(App, params)
		fmt.Println(string(resp), err)

		resp, err = calendar.DeleteCalendarById(App, params)
		fmt.Println(string(resp), err)
	})

	router.GET("/api/feishu/upload", Upload)

	router.GET("/api/feishu/upload2", func(c *gin.Context) {
		params := url.Values{}
		params.Add("type", "image")
		resp, err := approval.Upload(App, "hi.jpg", params)
		fmt.Println(string(resp), err)

	})

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
