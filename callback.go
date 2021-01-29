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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fastwego/feishu"

	"github.com/gin-gonic/gin"
)

func Callback(c *gin.Context) {
	// 加解密处理器
	dingCrypto := feishu.NewCrypto(FeishuConfig["EncryptKey"])

	// Post Body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return
	}

	log.Printf(string(body))

	msgJson := struct {
		Encrypt string `json:"encrypt"`
	}{}
	err = json.Unmarshal(body, &msgJson)
	if err != nil {
		return
	}

	decryptMsg, err := dingCrypto.GetDecryptMsg(msgJson.Encrypt)
	if err != nil {
		return
	}

	eventJson := struct {
		Type  string `json:"type"`
		Event struct {
			Type   string `json:"type"`
			OpenId string `json:"open_id"`
			Text   string `json:"text"`
		} `json:"event"`
	}{}

	err = json.Unmarshal(decryptMsg, &eventJson)
	if err != nil {
		return
	}

	switch eventJson.Type {
	case "url_verification":
		// 响应 challenge
		_, _ = c.Writer.Write(decryptMsg)
		log.Println(string(decryptMsg))
		return
	}

	switch eventJson.Event.Type {
	case "message":
		// 响应 消息
		_, _ = c.Writer.Write(decryptMsg)
		log.Println(string(decryptMsg))

		replyTextMsg := struct {
			OpenId  string `json:"open_id"`
			MsgType string `json:"msg_type"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		}{
			OpenId:  eventJson.Event.OpenId,
			MsgType: "text",
			Content: struct {
				Text string `json:"text"`
			}{Text: eventJson.Event.Text},
		}

		data, err := json.Marshal(replyTextMsg)
		if err != nil {
			fmt.Println(err)
			return
		}
		tenantAccessToken, err := Atm.GetAccessToken()
		if err != nil {
			log.Println(err)
			return
		}

		request, err := http.NewRequest(http.MethodPost, feishu.ServerUrl+"/open-apis/message/v4/send/", bytes.NewReader(data))
		resp, err := FeishuClient.Do(request, tenantAccessToken)
		fmt.Println(string(resp), err)
	}
}
