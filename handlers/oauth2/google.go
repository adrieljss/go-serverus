package oauth_handlers

import (
	"encoding/json"
	"fmt"

	"github.com/adrieljss/go-serverus/db"
	"github.com/adrieljss/go-serverus/env"
	"github.com/adrieljss/go-serverus/result"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleOAuth2(conf *oauth2.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		ctx.Redirect(307, url)
	}
}

var Oauth2Config *oauth2.Config

// Env variables needs to be loaded first. But i want the config specifications in this file
func SetOauth2Config() {
	Oauth2Config = &oauth2.Config{
		ClientID:     env.CGoogleClientID,
		ClientSecret: env.CGoogleClientSecret,
		RedirectURL:  fmt.Sprintf("%s/auth/oauth2/google/callback", env.CAppRootUrl),
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type GRes struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	PictureURL    string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GoogleOAuth2Callback(conf *oauth2.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := ctx.Query("code")
		t, err := conf.Exchange(ctx.Request.Context(), code)
		if err != nil {
			result.Err(400, err, "BAD_OAUTH2", "failed to login with google oauth2").SendJSON(ctx)
			return
		}

		client := conf.Client(ctx.Request.Context(), t)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			result.Err(400, err, "BAD_OAUTH2", "failed to login with google oauth2").SendJSON(ctx)
			return
		}

		defer resp.Body.Close()

		var gres GRes

		d := json.NewDecoder(resp.Body)
		if err := d.Decode(&gres); err != nil {
			result.ServerErr(err).SendJSON(ctx)
			return
		}

		// user exists
		exists, rerr := db.UserExistsEmail(ctx.Request.Context(), gres.Email)
		if rerr != nil {
			rerr.SendJSON(ctx)
			return
		}

		if exists {
			// login to acc
			res, rerr := db.DangerousLoginAndGenerateJWT(ctx.Request.Context(), gres.Email)
			if rerr != nil {
				rerr.SendJSON(ctx)
				return
			}
			ctx.Redirect(307, fmt.Sprintf("%s/login?jwt=%s", env.CFrontendRootUrl, res.JWT))
		} else {
			// sign up for acc, redirect to signup panel in frontend with some fields already filled out
			ctx.Redirect(307, fmt.Sprintf("%s/signup?email=%s&name=%s&picture=%s", env.CFrontendRootUrl, gres.Email, gres.Name, gres.PictureURL))
		}
	}
}
