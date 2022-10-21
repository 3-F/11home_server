package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/3-F/feishu-sdk-golang/core/model/vo"
	"github.com/gin-gonic/gin"
)

type ActionFiled struct {
	Value map[string]string `json:"value"`
	Tag   string            `json:"tag"`
}

type ActionRequest struct {
	OpenId        string      `json:"open_id"`
	UserId        string      `json:"user_id"`
	OpenMessageId string      `json:"open_message_id"`
	TenantKey     string      `json:"tenant_key"`
	Token         string      `json:"token"`
	Action        ActionFiled `json:"action"`
}

func main() {
	r := gin.Default()
	fiOpenId := "ou_b7dc8d6831dba04bec363734926bf0ea"
	answerTable := make(map[string]map[string]string)
	AnswerData, _ := ioutil.ReadFile("./440Answer.json")
	json.Unmarshal(AnswerData, &answerTable)

	answer := make(map[string]map[string]struct{})
	credit := make(map[string]uint)

	r.POST("/feishu", func(ctx *gin.Context) {
		jsonData, _ := ioutil.ReadAll(ctx.Request.Body)
		var rawData map[string]interface{}
		json.Unmarshal(jsonData, &rawData)
		if v, ok := rawData["type"]; ok {
			if v.(string) == "url_verification" {
				c := rawData["challenge"].(string)
				ctx.JSON(http.StatusOK, gin.H{
					"challenge": c,
				})
				return
			}
		}

		var actionReq ActionRequest
		json.Unmarshal(jsonData, &actionReq)
		fiWin := false
		gikkiWin := false
		respAnser := ""
		for k, v := range actionReq.Action.Value {
			if _, ok := answer[k][string(v[0])]; ok {
				continue
			} else if a, ok := answerTable[k][string(v[0])]; ok {
				if string(v[1:]) == a {
					respAnser = a
					if actionReq.OpenId == fiOpenId {
						credit["fi"] += 1
						fiWin = true
					} else {
						credit["gakki"] += 1
						gikkiWin = true
					}
					answer[k] = make(map[string]struct{})
					answer[k][string(v[0])] = struct{}{}
				}

			}
		}
		var respMsg string
		if fiWin {
			respMsg = "Fi Win! Point +1"
		} else if gikkiWin {
			respMsg = "Gakke Win! Point +1"
		} else {
			return
		}
		ctx.JSON(http.StatusOK, &vo.Card{
			Config: &vo.CardConfig{
				WideScreenMode: true,
			},
			Header: &vo.CardHeader{
				Template: "purple",
				Title: &vo.CardHeaderTitle{
					Tag:     "plain_text",
					Content: respMsg,
				},
			},
			Elements: []interface{}{
				vo.CardElementContentModule{
					Tag: "div",
					Text: &vo.CardElementText{
						Tag:     "lark_md",
						Content: "The answer is: " + respAnser,
					},
				},
			},
		})
	})
	r.Run(":4488")
}
