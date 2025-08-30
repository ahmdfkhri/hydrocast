package database

import (
	"context"
	"fmt"
	"log"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(conf *config.DatabaseConfig) *pgxpool.Pool {
	dbstring := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		conf.User, conf.Pass, conf.Host, conf.Port, conf.Name)

	db, err := pgxpool.New(context.Background(), dbstring)
	if err != nil {
		log.Fatalln(err)
	}

	return db
}
