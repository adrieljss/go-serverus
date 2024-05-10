package handlers

import (
	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GET - get a user by their user_id
//
// need to have parameter accessible by
//
//	ctx.Param("user_id")
func GetByUserId(ctx *gin.Context) {
	user_id := ctx.Param("user_id")
	user, err := db.FetchUserByUserId(user_id)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	user.ClearPrivateInfo()
	result.Ok(200, user).SendJSON(ctx)
}

// GET - Gets a user by their auth token
//
// relies on protected middleware
func GetMe(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	// ctx.Get("user") must exists at all times,
	// if this doesnt exist, panic
	if !exists {
		logrus.Fatal("GetMe handler must always be accompanied by the AuthRequired middleware")
		return
	}

	result.Ok(200, user).SendJSON(ctx)
}
