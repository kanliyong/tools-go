package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// BatchParam MySQL 数据迁移工具
type BatchParam struct{
	Target *sql.DB
	Source *sql.DB
	Columns []string
	Table string
	Size int
	FetchSQL string
}


func (b *BatchParam) Sync() (err error ){

	SQL := b.FetchSQL
	rows , err := b.Source.Query(SQL)
	if err != nil{
		log.Printf("query err sql = %s, err = %v", SQL, err)
		return
	}

	rowsBuffer := make([]interface{}, 0, b.Size)
	cols, _ := rows.Columns()
	for rows.Next(){
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			break
		}
		rowsBuffer = append(rowsBuffer, columns)
		if len(rowsBuffer) >= b.Size{
			b.insertBatch(b.Target, rowsBuffer)
			//rowsBuffer = make([]interface{}, 0, b.Size)
			rowsBuffer = rowsBuffer[:0]
		}
	}
	if len(rowsBuffer) > 0{
		b.insertBatch(b.Target, rowsBuffer)
	}

	if closeErr := rows.Close(); closeErr != nil {
		return closeErr
	}
	// Check for row scan error.
	if err != nil {
		return err
	}
	// Check for errors during row iteration.
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}

func (b *BatchParam) insertBatch(db *sql.DB, rows []interface{}) (err error){
	placeholder := make([]string, 0,len(rows))
	data := make([]interface{},0, len(rows) * len(b.Columns))
	for _, row := range rows{
		placeholder = append(placeholder, rowPlaceholder(len(b.Columns)))
		if row , ok := row.([]interface{}); ok{
			data = append(data, row...)
		}
	}
	sqltext := "insert into %s (%s) values %s"
	SQL := fmt.Sprintf(sqltext, b.Table, strings.Join(b.Columns,","), strings.Join(placeholder, ","))
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	result , err := db.Exec(SQL, data...)
	if err != nil{
		tx.Rollback()
		return
	}
	tx.Commit()
	log.Println(result)
	return nil
}
//rowPlaceholder
//(?,?,?,...)
func rowPlaceholder(len int) string{
	var b strings.Builder
	b.Grow(2 * len + 2)
	b.WriteString("(")
	for i := 0; i< len; i++ {
		b.WriteString("?")
		if i != len - 1{
			b.WriteString(",")
		}
	}
	b.WriteString(")")
	return b.String()
}