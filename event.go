package main

import (
	"fmt"

	"github.com/fastwego/feishu/types/event_types"
	"github.com/gin-gonic/gin"
)

func Callback(c *gin.Context) {

	event, err := App.Server.ParseEvent(c.Request)
	fmt.Println(event, err)

	switch event.(type) {

	case event_types.EventAppTicket:
		err := PublicApp.ReceiveAppTicketHandler(event.(event_types.EventAppTicket).Event.AppTicket)
		fmt.Println(err)
	}
}
