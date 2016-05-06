package db

import (
	"fmt"
	"database/sql"

	_ "github.com/lib/pq"
)

var conn *sql.DB

func Init() error {
	var err error
	conn, err = sql.Open("postgres", "user=dctnr dbname=dctnr sslmode=disable")
	if err != nil {
		fmt.Println("could not connect to db")
		return err
	}
	fmt.Println("db connection established")
	return nil
}

func Ping() error {
	return conn.Ping()
}

func Finalize() {
	conn.Close()
	fmt.Println("db connection finalized")
}
