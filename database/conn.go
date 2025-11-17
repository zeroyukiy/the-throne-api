package database

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
)

func Init() *sqlx.DB {
	// conn, err := sqlx.Connect("mysql", "root:password@tcp(127.0.0.1:3306)/thethrone?parseTime=true")
	conn, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
