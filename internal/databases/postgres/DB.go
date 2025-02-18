package postgres

import (
	"fmt"
	"isekai-shop/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

)

func NewPosgres(config *config.Database) *sqlx.DB {
	dsn := fmt.Sprint(config.FullURL)
	db, err := sqlx.Open(config.Driver, dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}
