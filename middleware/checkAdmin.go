package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"strconv"
	"time"
	"walk-server/global"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/utility"
)

func CheckAdmin(context *gin.Context) {
	jwtToken := context.GetHeader("Authorization")
	if jwtToken == "" {
		utility.ResponseError(context, "缺少登录凭证")
		context.Abort()
		return
	} else {
		jwtToken = jwtToken[7:]
	}
	jwtData, err := utility.ParseToken(jwtToken)
	// jwt token 解析失败
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	userID := utility.AesDecrypt(jwtData.OpenID, global.Config.GetString("server.AESSecret"))
	user_id, err := strconv.Atoi(userID)
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	user, err := adminService.GetAdminByID(uint(user_id))
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	if user == nil {
		utility.ResponseError(context, "未登陆")
		return
	}

	var requestData map[string]interface{}
	var jsonData []byte

	// 尝试从请求体中获取数据
	rawData, err := context.GetRawData()
	if err == nil && len(rawData) > 0 {
		// 请求体有数据，解析为 JSON
		if err := json.Unmarshal(rawData, &requestData); err != nil {
			utility.ResponseError(context, "请求体解析失败")
			context.Abort()
			return
		}

		// 将 requestData 转换为 JSON
		jsonData, err = json.Marshal(requestData)
		if err != nil {
			utility.ResponseError(context, "数据序列化失败")
			context.Abort()
			return
		}

		// 重置请求体，以便后续处理中读取请求体
		context.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
	} else {
		// 请求体无数据，从 query 参数中获取
		queryData := context.Request.URL.Query()
		if len(queryData) > 0 {
			requestData = make(map[string]interface{})
			for key, values := range queryData {
				if len(values) > 0 {
					requestData[key] = values[0] // 只取第一个值
				}
			}

			// 将 query 参数转换为 JSON
			jsonData, err = json.Marshal(requestData)
			if err != nil {
				utility.ResponseError(context, "query 数据序列化失败")
				context.Abort()
				return
			}
		} else {
			// 如果请求体和 query 参数都没有数据
			jsonData = []byte("{}")
		}
	}
	err = model.InsertForm(model.Form{
		AdminID: user.ID,
		Route:   user.Route,
		Point:   user.Point,
		Data:    jsonData,
		Time:    time.Now(),
	})
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}
	context.Next()
}
