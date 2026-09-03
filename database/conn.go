package database

import (
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

func Init() *sqlx.DB {
	conn, err := sqlx.Open("postgres", fmt.Sprintf("postgresql://%s:%s@localhost:5432/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("./database/init.sql")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	schema := string(b)

	conn.MustExec(schema)

	return conn
}
