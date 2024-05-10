package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var DB *pgxpool.Pool

func ConnectToDB(databaseUri string) {
	db, err := pgxpool.New(context.Background(), databaseUri)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(context.Background()); err != nil {
		panic(err)
	}

	logrus.Warn("successfully connected to postgres database")
	DB = db
}

func RowIsEmpty(rows pgx.Rows) bool {
	return rows.Next()
}
