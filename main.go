package main

import (
	"fmt"

	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/email"
	"github.com/adrieljansen/go-serverus/env"
	"github.com/adrieljansen/go-serverus/handlers"
	"github.com/adrieljansen/go-serverus/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// load env variables first
	godotenv.Load()

	// cache
	env.CAppName = env.LoadString("APP_NAME")
	env.CServerAddress = env.LoadString("SERVER_ADDRESS")
	env.CAppRootUrl = env.LoadString("APP_ROOT_URL")
	env.CProductionMode = env.LoadBool("PRODUCTION_MODE")
	env.CPostgresURI = env.LoadString("POSTGRES_URI")
	env.CJwtSignature = env.LoadByteSlice("JWT_SIGNATURE")
	env.CSMTPHost = env.LoadString("SMTP_HOST")
	env.CSMTPPort = env.LoadUint16("SMTP_PORT")
	env.CSMTPFrom = env.LoadString("SMTP_FROM")
	env.CSMTPPass = env.LoadString("SMTP_PASS")

	var r *gin.Engine
	if env.CProductionMode {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetLevel(logrus.WarnLevel)
		// TODO: add logrus.SetOutput()
	} else {
		// development mode
		r = gin.Default()
		r.POST("/auth/registerRaw", handlers.RegisterRaw)
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}

	db.ConnectToDB(env.CPostgresURI)

	email.StartEmailService()
	// email.SendEmailVerification("arvinhijinks@gmail.com", "123")

	middlewares.StartIPRateLimiterService(2, 5)

	r.Use(middlewares.RateLimitRequired())

	// is server ok
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(200, fmt.Sprintf("%s server is ok, root dir: %s", env.CAppName, env.CAppRootUrl))
	})

	r.GET("/verifyEmail")
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/verifyEmail", handlers.VerifyEmail)
	r.POST("/auth/login", handlers.Login)
	r.GET("/users/@me", middlewares.AuthRequired(), handlers.GetMe)
	r.GET("/users/:user_id", handlers.GetByUserId)
	r.PATCH("/users/@me", middlewares.AuthRequired(), handlers.PatchMe)
	r.PATCH("/users/@me/password", middlewares.AuthRequired(), handlers.PatchMeCredentials)
	r.POST("/users/@me/verifyResetPass", middlewares.AuthRequired(), handlers.VerifyResetPass)

	r.Run(env.CServerAddress)
}
