package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"walk-server/utility"
)

// GraduateRegisterData 定义接收校友报名用的数据的类型
type GraduateRegisterData struct {
	Name    string `json:"name" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Gender  int    `json:"gender" binding:"required"`
	Contact struct {
		Qq     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
	Healthcodeimgurl string `json:"healthcodeimgurl" binding:"required"`
}

func StudentRegister(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": "200", "msg": "register"})
}

func GraduateRegister(context *gin.Context) {
	var postData GraduateRegisterData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}
}
