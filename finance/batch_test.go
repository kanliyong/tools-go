package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestBatch(t *testing.T){
	config := DBConfig{
		host: "127.0.0.1",
		user: "root",
		password: "123456",
		database: "test",
	}

	db,err := sql.Open("mysql", config.Dsn())
	if err != nil{
		panic(err)
	}

	b := &BatchParam{
		Target:   db,
		Source:   db,
		Columns:  []string{"id","name"},
		Table:    "t1",
		Size:     2,
		FetchSQL: "select * from t",
	}
	b.Sync()
}
