package handlers

import (
	"strings"

	"github.com/adrieljss/go-serverus/db"
	"github.com/adrieljss/go-serverus/result"
	"github.com/gin-gonic/gin"
)

// ! TESTING PURPOSES ONLY
// these handlers are NOT supposed to be exposed to the public
// (security and vulnerability issues)

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

	res, errRes := db.CreateUserAndGenerateJWT(ctx.Request.Context(), registerRequestBody.UserId, registerRequestBody.Username, registerRequestBody.Email, registerRequestBody.Password)
	if errRes != nil {
		errRes.SendJSON(ctx)
		return
	}

	result.Ok(200, res).SendJSON(ctx)
}

// ! THIS IS DANGEROUS PLEASE DONT DO!, only in debug mode
// fetches all users (will be slow), but to test redis caching efficiency
func RedisTestRaw(ctx *gin.Context) {
	var users []db.User
	err := db.FetchManyWithCache(ctx, db.DB, &users, "SELECT * FROM users")
	if err != nil {
		result.ServerErr(err).SendJSON(ctx)
		return
	}
	result.Ok(200, users).SendJSON(ctx)
}
