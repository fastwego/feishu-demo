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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/fastwego/feishu"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {

	path := "hi.jpg"

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", path)
	if err != nil {
		return
	}
	_, err = io.Copy(part, file)
	_ = writer.WriteField("image_type", "message")
	err = writer.Close()
	if err != nil {
		return
	}
	request, _ := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/image/v4/put/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	tenantAccessToken, err := Atm.GetAccessToken()
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := FeishuClient.Do(request, tenantAccessToken)

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

	params := url.Values{}
	params.Add("image_key", image.Data.ImageKey)
	request, err = http.NewRequest(http.MethodGet, feishu.ServerUrl+"/open-apis/image/v4/get?"+params.Encode(), nil)

	resp, err = FeishuClient.Do(request, tenantAccessToken)

	fmt.Println(string(resp), err)

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
