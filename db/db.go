package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var DB *pgxpool.Pool
var RDB *redis.Client

func ConnectToDB(databaseUri string) {
	db, err := pgxpool.New(context.Background(), databaseUri)
	if err != nil {
		logrus.Fatal("cannot connect to postgres DB: ", err)
	}

	if err = db.Ping(context.Background()); err != nil {
		logrus.Fatal("cannot ping postgres DB, make sure DB is online")
	}

	logrus.Warn("successfully connected to postgres database")
	DB = db
}

func ConnectToRedis(redisUri string) {
	opts, err := redis.ParseURL(redisUri)
	if err != nil {
		logrus.Fatal("cannot connect to Redis DB: ", err)
	}

	rdb := redis.NewClient(opts)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		logrus.Fatal("cannot ping redis service, make sure redis is online")
	}

	logrus.Warn("successfully connected to redis service")
	RDB = rdb
}
