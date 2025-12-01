package admin

import (
	"errors"
	"log"
	"strconv"
	"walk-server/constant"
	"walk-server/global"
	"walk-server/service/adminService"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type autoLoginForm struct {
	Code string `json:"code" binding:"required"`
}

type passwordLoginForm struct {
	Username string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type LoginForm struct {
	Username string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	ID           uint   `json:"admin_id"`
	WechatOpenID string `json:"-"`
	Name         string `json:"name"`
	Account      string `json:"account"`
	Password     string `json:"-"`
	Point        string `json:"point"`
	Route        uint8  `json:"route"`
}

func AuthByPassword(c *gin.Context) {
	var postForm passwordLoginForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	user, err := adminService.GetUserByAccount(postForm.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utility.ResponseError(c, "账号错误")
		return
	}
	if err != nil {
		log.Println(err)
		utility.ResponseError(c, "服务错误")
		return
	}

	if user.Password != postForm.Password {
		utility.ResponseError(c, "密码错误")
		return
	}

	if user.WechatOpenID == "" {
		session, err := global.MiniProgram.Auth.Session(global.Wctx, postForm.Code)
		if err != nil {
			utility.ResponseError(c, "OpenID错误")
			return
		}
		user.WechatOpenID = session.OpenID
		adminService.UpdateOpenID(user)
	}
	var jwtData utility.JwtData
	jwtData.OpenID = utility.AesEncrypt(strconv.Itoa(int(user.ID)), global.Config.GetString("server.AESSecret"))
	// 生成 JWT
	jwtToken, err := utility.GenerateStandardJwt(&jwtData)

	utility.ResponseSuccess(c, gin.H{
		"admin": LoginResp{
			ID:           user.ID,
			WechatOpenID: user.WechatOpenID,
			Name:         user.Name,
			Account:      user.Account,
			Point:        constant.GetPointName(user.Route, user.Point),
			Route:        user.Route,
		},
		"jwt": jwtToken,
	})
}

func WeChatLogin(c *gin.Context) {
	var postForm autoLoginForm
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	session, err := global.MiniProgram.Auth.Session(global.Wctx, postForm.Code)
	if err != nil {
		log.Println(err)
		utility.ResponseError(c, "服务错误")
		return
	}

	user := adminService.GetUserByWechatOpenID(session.OpenID)
	if user == nil {
		utility.ResponseError(c, "登陆错误")
		return
	}

	var jwtData utility.JwtData
	jwtData.OpenID = utility.AesEncrypt(strconv.Itoa(int(user.ID)), global.Config.GetString("server.AESSecret"))
	// 生成 JWT
	jwtToken, err := utility.GenerateStandardJwt(&jwtData)

	utility.ResponseSuccess(c, gin.H{
		"admin": LoginResp{
			ID:           user.ID,
			WechatOpenID: user.WechatOpenID,
			Name:         user.Name,
			Account:      user.Account,
			Point:        constant.GetPointName(user.Route, user.Point),
			Route:        user.Route,
		},
		"jwt": jwtToken,
	})
}

func AuthWithoutCode(c *gin.Context) {
	var postForm LoginForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	user, err := adminService.GetUserByAccount(postForm.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utility.ResponseError(c, "账号错误")
		return
	}
	if err != nil {
		log.Println(err)
		utility.ResponseError(c, "服务错误")
		return
	}

	if user.Password != postForm.Password {
		utility.ResponseError(c, "密码错误")
		return
	}

	var jwtData utility.JwtData
	jwtData.OpenID = utility.AesEncrypt(strconv.Itoa(int(user.ID)), global.Config.GetString("server.AESSecret"))
	// 生成 JWT
	jwtToken, err := utility.GenerateStandardJwt(&jwtData)
	utility.ResponseSuccess(c, gin.H{
		"admin": LoginResp{
			ID:           user.ID,
			WechatOpenID: user.WechatOpenID,
			Name:         user.Name,
			Account:      user.Account,
			Point:        constant.GetPointName(user.Route, user.Point),
			Route:        user.Route,
		},
		"jwt": jwtToken,
	})
}

type BlockWithSecretForm struct {
	Secret string `json:"secret" binding:"required"`
}

func BlockWithSecret(c *gin.Context) {
	var postForm BlockWithSecretForm
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}
	utility.ResponseSuccess(c, nil)
}
