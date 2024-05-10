package handlers

import (
	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type RegisterRequestBody struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// pfp url can only be set after login
	// PfpUrl   string `json:"pfp_url"`
	Password string `json:"password"`
}

func (a RegisterRequestBody) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.UserId, validation.Required, validation.Length(3, 35), is.Alphanumeric),
		validation.Field(&a.Username, validation.Required, validation.Length(2, 35)),
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, validation.Required, validation.Length(5, 30)),
	)
}

// POST - register a user
// needs the following fields in a JSON body
//
// user_id, username, email, password
//
// content:
//
//	`user`: User
//	`jwt`: string
func Register(ctx *gin.Context) {
	var registerRequestBody RegisterRequestBody
	if err := ctx.BindJSON(&registerRequestBody); err != nil {
		result.ErrBodyBind().SendJSON(ctx)
		return
	}

	validateErr := registerRequestBody.Validate()
	if validateErr != nil {
		result.ErrValidate(validateErr).SendJSON(ctx)
		return
	}

	user, errRes := db.CreateUser(registerRequestBody.UserId, registerRequestBody.Username, registerRequestBody.Email, registerRequestBody.Password)
	if errRes != nil {
		errRes.SendJSON(ctx)
		return
	}

	jwt, err := db.GenerateJwtToken(user)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	mp := make(map[string]any)
	mp["user"] = user
	mp["jwt"] = jwt
	result.Ok(200, mp).SendJSON(ctx)

	// TODO
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

	user, err := db.FetchUserByCredentials(loginRequestBody.EmailOrUserId, loginRequestBody.Password)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	jwt, err := db.GenerateJwtToken(user)
	if err != nil {
		err.SendJSON(ctx)
		return
	}

	mp := make(map[string]any)
	mp["user"] = user
	mp["jwt"] = jwt
	result.Ok(200, mp).SendJSON(ctx)
}
