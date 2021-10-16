/*
 * Copyright (c) 2021 IInfo.
 */

package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Login(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": "200", "msg": "login"})
}

func Register(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": "200", "msg": "register"})
}
