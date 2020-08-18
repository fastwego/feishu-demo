package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fastwego/feishu/apis/message"

	"github.com/fastwego/feishu/apis/capabilities/calendar"
	"github.com/fastwego/feishu/types/event_types"

	"github.com/fastwego/feishu/apis/capabilities/meeting"

	"github.com/fastwego/feishu"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
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

	router.POST("/api/feishu/callback", func(c *gin.Context) {

		event, err := App.Server.ParseEvent(c.Request)
		fmt.Println(event, err)

		switch event.(type) {
		case event_types.EventChallenge: // url 校验
			App.Server.Challenge(c.Writer, event.(event_types.EventChallenge))
		case event_types.EventAppTicket:
			err := PublicApp.ReceiveAppTicketHandler(event.(event_types.EventAppTicket).Event.AppTicket)
			fmt.Println(err)
		case event_types.EventMessageText:
			userMsg := event.(event_types.EventMessageText)
			fmt.Println(userMsg)

			replyTextMsg := struct {
				OpenId  string `json:"open_id"`
				MsgType string `json:"msg_type"`
				Content struct {
					Text string `json:"text"`
				} `json:"content"`
			}{
				OpenId:  userMsg.Event.OpenID,
				MsgType: "text",
				Content: struct {
					Text string `json:"text"`
				}{Text: userMsg.Event.Text},
			}

			data, err := json.Marshal(replyTextMsg)
			if err != nil {
				fmt.Println(err)
				return
			}

			resp, err := message.Send(App, data)
			fmt.Println(string(resp), err)
		}

	})

	router.GET("/api/feishu/demo", func(c *gin.Context) {
		params := url.Values{}
		list, err := meeting.BuildingList(PublicApp.App, params)
		fmt.Println(string(list), err)

		params = url.Values{}
		params.Add("calendarId", "10086")
		resp, err := calendar.GetCalendarById(App, params)
		fmt.Println(string(resp), err)

		resp, err = calendar.DeleteCalendarById(App, params)
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
