package main

import (
	"github.com/adrieljansen/go-serverus/db"
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
	env.CServerAddress = env.LoadString("SERVER_ADDRESS")
	env.CProductionMode = env.LoadBool("PRODUCTION_MODE")
	env.CPostgresURI = env.LoadString("POSTGRES_URI")
	env.CJwtSignature = env.LoadByteSlice("JWT_SIGNATURE")

	if env.CProductionMode {
		gin.SetMode(gin.ReleaseMode)
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetLevel(logrus.WarnLevel)
		// TODO: add logrus.SetOutput()
	} else {
		// development mode
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}

	db.ConnectToDB(env.CPostgresURI)
	r := gin.Default()

	// is server ok
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "server is ok")
	})

	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/users/@me", middlewares.AuthRequired(), handlers.GetMe)
	r.GET("/users/:user_id", handlers.GetByUserId)

	r.Run(env.CServerAddress)
}
