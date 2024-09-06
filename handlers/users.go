package handlers

import (
	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
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

type PatchUserRequestBody struct {
	UserId   *string `json:"user_id"`
	Username *string `json:"username"`
	PfpUrl   *string `json:"pfp_url"`
}

func (a PatchUserRequestBody) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.UserId, validation.Length(3, 35), is.Alphanumeric),
		validation.Field(&a.Username, validation.Length(2, 35)),
		validation.Field(&a.PfpUrl, is.URL),
	)
}

// TODO
// PATCH - Updates user information
//
// relies on protected middleware
func PatchMe(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		logrus.Fatal("PatchMe handler must always be accompanied by the AuthRequired middleware")
		return
	}

	var patchRequestBody PatchUserRequestBody
	if err := ctx.BindJSON(&patchRequestBody); err != nil {
		result.ErrBodyBind().SendJSON(ctx)
		return
	}

	validateErr := patchRequestBody.Validate()
	if validateErr != nil {
		result.ErrValidate(validateErr).SendJSON(ctx)
		return
	}

	newUser, err := db.UpdateUserInfo(user.(db.User), patchRequestBody.UserId, patchRequestBody.Username, patchRequestBody.PfpUrl)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(200, newUser).SendJSON(ctx)
}
