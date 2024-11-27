package middlewares

import (
	"strings"

	"github.com/adrieljss/go-serverus/db"
	"github.com/adrieljss/go-serverus/result"
	"github.com/gin-gonic/gin"
)

// Middleware to protect routes.
// Fetches user informations from jwt token and sets them to the request context.
// Uses a JWT-like authentication.
// User information can be accessed via
//
//	ctx.Get("user") // -> returns db.User
func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			result.Err(401, nil, "NO_AUTH_HEADER", "no Authorization header is sent, but this is a protected route").SendJSON(ctx)
			ctx.Abort()
			return
		}
		headerArr := strings.Split(header, " ")
		if len(headerArr) != 2 || headerArr[0] != "Bearer" {
			result.Err(401, nil, "INVALID_AUTH_HEADER", "invalid auth header").SendJSON(ctx)
			ctx.Abort()
			return
		}

		token := headerArr[1]
		user, err := db.ExtractUserFromJwt(ctx.Request.Context(), token)
		if err != nil {
			err.SendJSON(ctx)
			ctx.Abort()
			return
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}
