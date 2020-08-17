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

	"github.com/fastwego/feishu/apis/capabilities/calendar"

	"github.com/fastwego/feishu/apis/capabilities/meeting"

	"github.com/fastwego/feishu/types"

	"github.com/fastwego/feishu"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

var App *feishu.App

func init() {
	// 加载配置文件
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	App = feishu.NewInternalApp(feishu.AppConfig{
		AppId:     viper.GetString("APPID"),
		AppSecret: viper.GetString("SECRET"),
		//VerificationToken: viper.GetString("TOKEN"),
		//EncryptKey:        viper.GetString("AESKey"),
	})
}

func main() {

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.POST("/api/feishu/callback", func(c *gin.Context) {

		event, err := App.Server.ParseEvent(c.Request)
		fmt.Println(event, err)

		switch event.(type) {
		case types.EventChallenge: // url 校验
			App.Server.Challenge(c.Writer, event.(types.EventChallenge))
		case types.EventAppTicket: // app ticket
			err := App.AppTicket.ReceiveAppTicketHandler(App, event.(types.EventAppTicket).Event.AppTicket)
			fmt.Println(err)
		case types.EventMessageText:
			fmt.Println(event.(types.EventMessageText))
		}

	})

	router.GET("/api/feishu/demo", func(c *gin.Context) {
		params := url.Values{}
		list, err := meeting.BuildingList(App, params)
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
