package db

import (
	"fmt"
	"time"

	"github.com/adrieljansen/go-serverus/env"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwtToken(user *User) (string, *result.Error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	// claims["username"] = user.Username
	claims["user_uid"] = user.Uid
	claims["hashed_password"] = user.Password
	// jwt token is valid for 7 days after issue
	claims["exp"] = time.Now().AddDate(0, 0, 7).Unix()
	claims["iat"] = time.Now().Unix()

	tokenString, err := token.SignedString(env.CJwtSignature)
	if err != nil {
		return "", result.ServerErr(err)
	}
	return tokenString, nil
}

// User returns nil if token is invalid or no user is found
func ExtractUserFromJwt(tokenString string) (*User, *result.Error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token provided")
		}
		return env.CJwtSignature, nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if err != nil || !ok || !token.Valid {
		return nil, result.Err(401, err, "INVALID_TOKEN", "invalid or expired auth token provided")
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil, result.Err(401, err, "INVALID_TOKEN", "no exp field in auth token or invalid exp field")
	}

	// if exp is before or IS now(),
	// then we declare the token as expired
	if exp.Compare(time.Now()) <= 0 {
		return nil, result.Err(401, err, "EXPIRED_TOKEN", "the given auth token has expired")
	}

	uid, hasUid := claims["user_uid"]
	if !hasUid {
		return nil, result.Err(401, err, "INVALID_TOKEN", "auth token must have user_uid field")
	}

	hashedPsw, hasHashedPsw := claims["hashed_password"]
	if !hasHashedPsw {
		return nil, result.Err(401, err, "INVALID_TOKEN", "auth token must have hashed_password field")
	}

	return FetchUserByUidAndHashedPassword(uid.(string), hashedPsw.(string))
}
