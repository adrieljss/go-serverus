package oauth_handlers

import (
	"encoding/json"
	"fmt"

	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/env"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func GoogleOAuth2(conf *oauth2.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		ctx.Redirect(307, url)
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
		exists, rerr := db.UserExistsEmail(gres.Email)
		if rerr != nil {
			rerr.SendJSON(ctx)
			return
		}

		if exists {
			// login to acc
			res, rerr := db.DangerousLoginAndGenerateJWT(gres.Email)
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
