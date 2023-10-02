package adminService

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"
)

func GetUserByAccount(username string) (*model.Admin, error) {
	user := model.Admin{}
	result := global.DB.Where(
		&model.Admin{
			Account: username,
		},
	).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func GetUserByWechatOpenID(openid string) *model.Admin {
	user := model.Admin{}
	result := global.DB.Where(
		&model.Admin{
			WechatOpenID: openid,
		},
	).First(&user)
	if result.Error != nil {
		return nil
	}

	return &user
}

func GetAdminByID(id uint) (*model.Admin, error) {
	user := model.Admin{}
	result := global.DB.Where(
		&model.Admin{
			ID: id,
		},
	).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func GetAdminByJWT(context *gin.Context) (*model.Admin, error) {
	jwtData := utility.GetJwtData(context)
	// jwt token 解析失败
	userID := utility.AesDecrypt(jwtData.OpenID, global.Config.GetString("server.AESSecret"))
	user_id, err := strconv.Atoi(userID)
	user, err := GetAdminByID(uint(user_id))
	if err != nil {
		return nil, err
	}
	return user, err
}
