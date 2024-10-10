package handlers

import (
	"strings"
	"time"

	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/email"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

// # POST - Register a user
//
// needs the following fields in a JSON body
//
// user_id, username, email, password
//
// content:
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
		has, err := db.HasUserWithSameId(registerRequestBody.UserId)
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
		has, err := db.UserExistsEmail(registerRequestBody.Email)
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
		}, time.Now().Add(time.Minute*40).Unix())
	})

	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(204, "verification email sent").SendJSON(ctx)
}

// POST - To verify email using given otp code
//
// needs the following queries
//
//   - otp
//   - email
//
// content:
//
//	`user`: User
//	`jwt`: string
//
// after this forward user to frontend
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

	res, errRes := db.CreateUserAndGenerateJWT(req.UserToCreate.UserId, req.UserToCreate.Username, req.UserToCreate.Email, req.UserToCreate.Password)
	if errRes != nil {
		errRes.SendJSON(ctx)
		return
	}

	email.PendingConfirmationEmailRegisterCache.Delete(emailString)

	result.Ok(200, res).SendJSON(ctx)
}

// # POST - Register a user
//
// raw - USE ONLY IN DEBUG MODE, NO EMAIL CONFIRMATION REQUIRED
//
// needs the following fields in a JSON body
//
// user_id, username, email, password
//
// content:
//
//	`user`: User
//	`jwt`: string
func RegisterRaw(ctx *gin.Context) {
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

	res, errRes := db.CreateUserAndGenerateJWT(registerRequestBody.UserId, registerRequestBody.Username, registerRequestBody.Email, registerRequestBody.Password)
	if errRes != nil {
		errRes.SendJSON(ctx)
		return
	}

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

// POST - logins with `email_or_userid` and `password` in JSON body
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

	res, err := db.LoginAndGenerateJWT(loginRequestBody.EmailOrUserId, loginRequestBody.Password)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	result.Ok(200, res).SendJSON(ctx)
}
