package handlers

import (
	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/email"
	"github.com/adrieljansen/go-serverus/env"
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
	UserId    string `json:"user_id"`
	Username  string `json:"username"`
	PfpUrl    string `json:"pfp_url"`
	Biography string `json:"biography"`
}

func (a PatchUserRequestBody) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.UserId, validation.Length(3, 35), is.Alphanumeric),
		validation.Field(&a.Username, validation.Length(2, 35)),
		validation.Field(&a.PfpUrl, is.URL),
		validation.Field(&a.Biography, validation.Length(0, 200)),
	)
}

// PATCH - Updates user information
//
// relies on protected middleware
func PatchMe(ctx *gin.Context) {
	usr, exists := ctx.Get("user")
	if !exists {
		logrus.Fatal("PatchMe handler must always be accompanied by the AuthRequired middleware")
		return
	}
	var user *db.User = usr.(*db.User)

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

	newUser, err := db.UpdateUserInfo(*user, patchRequestBody.UserId, patchRequestBody.Username, patchRequestBody.Biography, patchRequestBody.PfpUrl)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(200, newUser).SendJSON(ctx)
}

type PatchRequestResetPass struct {
	NewPassword string `json:"new_password"`
}

func (a PatchRequestResetPass) Validate() error {
	return validation.ValidateStruct(
		validation.Field(&a.NewPassword, validation.Required, validation.Length(5, 30)),
	)
}

// PATCH - Updates user information (for password)
//
// relies on protected middleware
func PatchMeCredentials(ctx *gin.Context) {
	usr, exists := ctx.Get("user")
	if !exists {
		logrus.Fatal("PatchMeCredentials handler must always be accompanied by the AuthRequired middleware")
		return
	}

	var user *db.User = usr.(*db.User)

	var patchRequestBody PatchRequestResetPass
	if err := ctx.BindJSON(&patchRequestBody); err != nil {
		result.ErrBodyBind().SendJSON(ctx)
		return
	}

	validateErr := patchRequestBody.Validate()
	if validateErr != nil {
		result.ErrValidate(validateErr).SendJSON(ctx)
		return
	}

	err := email.GenerateOtpAndAct(patchRequestBody.NewPassword, func(otpCode string) {
		err := email.SendEmailResetPassword(user.Email, otpCode, user.UserId)
		if err != nil {
			return
		}

		email.PendingResetPassCache.Store(user.Email, email.PendingResetPass{
			OTPCode:     otpCode,
			NewPassword: patchRequestBody.NewPassword,
			User:        user,
		}, int64(env.EmailResetPassTTL))
	})

	if err != nil {
		err.SendJSON(ctx)
		return
	}
	result.Ok(204, "reset password email sent").SendJSON(ctx)
}

// POST - OTP Code to reset password
//
// needs the following queries
//
//   - otp
//
// content:
//
//	`user`: User
//	`jwt`: string
//
// relies on protected middleware
func VerifyResetPass(ctx *gin.Context) {
	otpCode := ctx.Query("otp")
	if otpCode == "" {
		result.Err(400, nil, "EMPTY_OTP_QUERY", "cannot have empty otp in url query").SendJSON(ctx)
		return
	}

	usr, exists := ctx.Get("user")
	if !exists {
		logrus.Fatal("VerifyResetPass handler must be used with protected middleware")
		return
	}

	var user *db.User = usr.(*db.User)

	req, exists := email.PendingResetPassCache.Get(user.Email)

	if !exists || otpCode != req.OTPCode {
		result.Err(400, nil, "OTP_NOT_EXIST", "the otp for resetting password has already expired or does not exist").SendJSON(ctx)
		return
	}

	newUser, err := db.UpdateUserPassword(*user, req.NewPassword)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	jwt, err := db.GenerateJwtToken(newUser)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	email.PendingResetPassCache.Delete(user.Email)

	mp := make(map[string]any)
	mp["user"] = user
	mp["jwt"] = jwt
	result.Ok(200, mp).SendJSON(ctx)
}
