package main

import (
	"encoding/json"
	"fmt"

	"github.com/fastwego/feishu/apis/message"
	"github.com/fastwego/feishu/types/event_types"
	"github.com/gin-gonic/gin"
)

func Callback(c *gin.Context) {

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

}
