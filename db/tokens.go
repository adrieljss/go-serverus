package db

import (
	"context"
	"fmt"
	"time"

	"github.com/adrieljss/go-serverus/env"
	"github.com/adrieljss/go-serverus/result"
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
func ExtractUserFromJwt(ctx context.Context, tokenString string) (*User, *result.Error) {
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

	return FetchUserByUidAndHashedPassword(ctx, uid.(string), hashedPsw.(string))
}

type GenericUserJWTRes struct {
	JWT  string `json:"jwt"`
	User *User  `json:"user"`
}

// this is dangerous because only email is needed to authenticate
//
// use only in OAuth2 stuff, or other of which you are sure that
// the owner of the email is correct
func DangerousLoginAndGenerateJWT(ctx context.Context, email string) (*GenericUserJWTRes, *result.Error) {
	user, err := FetchUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	jwt, err := GenerateJwtToken(user)
	if err != nil {
		return nil, err
	}

	return &GenericUserJWTRes{
		JWT:  jwt,
		User: user,
	}, nil
}

// login with the essential credentials
func LoginAndGenerateJWT(ctx context.Context, emailOrUserId string, password string) (*GenericUserJWTRes, *result.Error) {
	user, err := FetchUserByCredentials(ctx, emailOrUserId, password)
	if err != nil {
		return nil, err
	}

	jwt, err := GenerateJwtToken(user)
	if err != nil {
		return nil, err
	}

	return &GenericUserJWTRes{
		JWT:  jwt,
		User: user,
	}, nil
}

// create a user and generate jwt token
func CreateUserAndGenerateJWT(ctx context.Context, userId string, username string, email string, password string) (*GenericUserJWTRes, *result.Error) {
	user, err := CreateUser(ctx, userId, username, email, password)
	if err != nil {
		return nil, err
	}

	jwt, err := GenerateJwtToken(user)
	if err != nil {
		return nil, err
	}

	return &GenericUserJWTRes{
		JWT:  jwt,
		User: user,
	}, nil
}
