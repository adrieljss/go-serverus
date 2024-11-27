package handlers

import (
	"strings"

	"github.com/adrieljss/go-serverus/db"
	"github.com/adrieljss/go-serverus/email"
	"github.com/adrieljss/go-serverus/env"
	"github.com/adrieljss/go-serverus/result"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

// [POST] Registers a user
//
// Registers a user, but sending them an email first to confirm email.
// Needs the following fields in the JSON body:
//
//   - user_id
//   - username
//   - email
//   - password
//
// Success Content:
//
//	`user`: User
//	`jwt`: string
func Register(ctx *gin.Context) {
	var registerRequestBody db.RequiredUser
	registerRequestBody.Email = strings.ToLower(registerRequestBody.Email)

	if err := ctx.BindJSON(&registerRequestBody); err != nil {
		result.ErrBodyBind().SendJSON(ctx)
		return
	}

	validateErr := registerRequestBody.Validate()
	if validateErr != nil {
		result.ErrValidate(validateErr).SendJSON(ctx)
		return
	}

	metadata := make(map[string]string)
	dupeExists := false
	// check dupe
	{
		has, err := db.UserExistsUserId(ctx.Request.Context(), registerRequestBody.UserId)
		if err != nil {
			err.SendJSON(ctx)
			return
		}
		if has {
			metadata["user_id"] = "there exists a user with the same user_id"
			dupeExists = true
		}
	}

	{
		has, err := db.UserExistsEmail(ctx.Request.Context(), registerRequestBody.Email)
		if err != nil {
			err.SendJSON(ctx)
			return
		}
		if has {
			metadata["email"] = "there exists a user with the same email"
			dupeExists = true
		}
	}

	if dupeExists {
		result.ErrWithMetadata(400, nil, "USER_EXISTS_SAME_CREDENTIALS", "another user already exists with the same credentials", metadata).SendJSON(ctx)
		return
	}

	err := email.GenerateOtpAndAct(registerRequestBody.Password, func(otpCode string) {
		err := email.SendEmailVerification(registerRequestBody.Email, otpCode, registerRequestBody.UserId)
		if err != nil {
			// err.SendJSON(ctx)
			return
		}

		email.PendingConfirmationEmailRegisterCache.Store(registerRequestBody.Email, email.PendingConfirmationEmail{
			VerificationCode: otpCode,
			UserToCreate:     &registerRequestBody,
		}, int64(env.EmailConfirmationTTL))
	})

	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(204, "verification email sent").SendJSON(ctx)
}

// [POST] Verifies email using OTP code.
//
// No JSON body needed.
// Needs the following url queries:
//
//   - otp
//   - email
//
// Success content:
//
//	`user`: User
//	`jwt`: string
func VerifyEmail(ctx *gin.Context) {
	emailString := ctx.Query("email")
	otpCode := ctx.Query("otp")
	if emailString == "" {
		result.Err(400, nil, "EMPTY_EMAIL_QUERY", "cannot have empty email in url query").SendJSON(ctx)
		return
	}

	if otpCode == "" {
		result.Err(400, nil, "EMPTY_OTP_QUERY", "cannot have empty otp in url query").SendJSON(ctx)
		return
	}

	req, exists := email.PendingConfirmationEmailRegisterCache.Get(emailString)
	if !exists || req.VerificationCode != otpCode {
		result.Err(404, nil, "OTP_NOT_EXIST", "the otp for given email does not exist or has expired").SendJSON(ctx)
		return
	}

	res, errRes := db.CreateUserAndGenerateJWT(ctx.Request.Context(), req.UserToCreate.UserId, req.UserToCreate.Username, req.UserToCreate.Email, req.UserToCreate.Password)
	if errRes != nil {
		errRes.SendJSON(ctx)
		return
	}

	email.PendingConfirmationEmailRegisterCache.Delete(emailString)

	result.Ok(200, res).SendJSON(ctx)
}

type LoginRequestBody struct {
	EmailOrUserId string `json:"email_or_userid"`
	Password      string `json:"password"`
}

func (a LoginRequestBody) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.EmailOrUserId, validation.Required, validation.Length(3, 50)),
		validation.Field(&a.Password, validation.Required, validation.Length(5, 30)),
	)
}

// [POST] Logins a User.
//
// Needs the following fields in the JSON body:
//   - email_or_userid
//   - password
//
// content:
//
//	`user`: User
//	`jwt`: string
func Login(ctx *gin.Context) {
	var loginRequestBody LoginRequestBody
	if err := ctx.BindJSON(&loginRequestBody); err != nil {
		result.ErrBodyBind().SendJSON(ctx)
		return
	}

	validateErr := loginRequestBody.Validate()
	if validateErr != nil {
		result.ErrValidate(validateErr).SendJSON(ctx)
		return
	}

	res, err := db.LoginAndGenerateJWT(ctx.Request.Context(), loginRequestBody.EmailOrUserId, loginRequestBody.Password)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(200, res).SendJSON(ctx)
}
