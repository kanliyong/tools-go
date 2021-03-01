package main

import (
	"context"
	"database/sql"
	"log"
)

func withTx(db *sql.DB, callback func() error) (err error) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	err = callback()
	if err != nil {
		rb := tx.Rollback()
		if rb != nil {
			log.Printf("rollback tx err %v", err)
		}
		return err
	}
	cm := tx.Commit()
	if cm != nil {
		log.Printf("commit tx err %v", err)
	}
	return err
}


func main() {

}
