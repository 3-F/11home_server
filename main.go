package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/3-F/feishu-sdk-golang/core/model/vo"
	"github.com/gin-gonic/gin"
)

type ActionValue struct {
	Answer map[string]string `json:"answer"`
	Data   string            `json:"data"`
}

type ActionFiled struct {
	Value ActionValue `json:"value"`
	Tag   string      `json:"tag"`
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
	answerTableData, _ := ioutil.ReadFile("./440Answer.json")
	json.Unmarshal(answerTableData, &answerTable)

	const answerPath = "./answer.json"
	answerData, _ := ioutil.ReadFile(answerPath)
	answer := make(map[string]map[string]struct{})
	json.Unmarshal(answerData, &answer)

	const creditPath = "./credit.json"
	creditData, _ := ioutil.ReadFile(creditPath)
	credit := make(map[string]uint)
	json.Unmarshal(creditData, &credit)

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
			} else if v.(string) == "points" {
				ctx.JSON(http.StatusOK, credit)
				return
			}
		}

		var actionReq ActionRequest
		json.Unmarshal(jsonData, &actionReq)
		fiWin := false
		gikkiWin := false
		respAnswer := ""
		respData := actionReq.Action.Value.Data
		for k, v := range actionReq.Action.Value.Answer {
			if a, ok := answerTable[k][string(v[0])]; ok {
				if string(v[1:]) == a {
					respAnswer = a
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
		// // write back
		answerRawData, _ := json.Marshal(answer)
		os.WriteFile(answerPath, answerRawData, 0644)
		creditRawData, _ := json.Marshal(credit)
		os.WriteFile(creditPath, creditRawData, 0644)

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
						Content: respData,
					},
				},
				vo.CardElementBrModule{
					Tag: "hr",
				},
				vo.CardElementContentModule{
					Tag: "div",
					Text: &vo.CardElementText{
						Tag:     "lark_md",
						Content: "The answer is: " + respAnswer,
					},
				},
			},
		})
	})

	r.Run(":4488")
}
