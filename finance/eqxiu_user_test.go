package main

import "database/sql"

// eqxiuUserTestRemoveAll
func eqxiuUserTestRemoveAll(db *sql.DB) (err error) {
	s := `delete from eqs_user_test`
	return withTx(db, func() error {
		result, err := db.Exec(s)
		if err != nil {
			return err
		}
		_, err = result.RowsAffected()
		return err
	})
}

func InitSync(source, dist *sql.DB, batchSize int)( err error ){
	bp := &BatchParam{
		Target:    dist,
		Source:    source,
		Columns:   []string{"id","type","create_time","is_calculate","property","group_id","belong_product_line","belong_person" },
		Table:     "eqs_user_test",
		Size:      100,
		FetchSQL:  `insert into eqs_user_test 
        (  id,type,create_time,is_calculate,property,group_id,belong_product_line,belong_person ) 
        values 
        (%s, %s ,%s ,%s, %s ,%s ,%s, %s ) `,
	}
	err = bp.Sync()
	return err
}
