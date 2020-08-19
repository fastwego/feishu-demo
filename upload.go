package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/fastwego/feishu/apis/message"
	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	params := url.Values{}
	params.Add("image_type", "message")
	resp, err := message.ImagePut(App, "hi.jpg", params)
	fmt.Println(string(resp), err)

	if err != nil {
		return
	}
	image := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(resp, &image)
	if err != nil {
		return
	}

	params.Add("image_key", image.Data.ImageKey)
	resp, err = message.ImageGet(App, params)
	if err != nil {
		fmt.Println(err)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(resp)

	html := `<!DOCTYPE html>
<html>
  <head>
    <title>Display Image</title>
  </head>
  <body>
    <img style='display:block;' id='base64image'                 
       src='data:image/jpeg;base64, ` + encoded + `' />
  </body>
</html>`

	fmt.Println(html)

	_, _ = c.Writer.WriteString(html)

}
