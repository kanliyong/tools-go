package main

import (
	"database/sql"
	"fmt"
)

type DBConfig struct {
	host     string
	user     string
	password string
	database string
}

// Dsn
// root:Eqxiu@2019@tcp(10.0.10.49)/nginx?charset=utf8mb4&collation=utf8mb4_general_ci
func (c *DBConfig) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci",
		c.user, c.password, c.host, c.database)
}


var (
	db,maxDB,payDB    *sql.DB
	mysql = &DBConfig{
		host:     "info-zt.mysql.res.eqxiu.com",
		user:     "eqsfinance",
		password: "CI6InK4@mC%A4QNw",
		database: "eqs_max_finance",
	}

	maxMysqlConfig = &DBConfig{
		host:     "info-zt-cw.mysql.res.eqxiu.com",
		user:     "eqs_max",
		password: "9JX4JCHpd3Anq2",
		database: "eqs_max",
	}
	
	payMysqlConfig = &DBConfig{
		host:     "max-slave-pay.mysql.res.eqxiu.com",
		user:     "maxpay",
		password: "Ghx2@nA2eh%oMhW",
		database: "max_pay",
	}
)

func init() {
	var err error
	db, err = sql.Open("mysql", mysql.Dsn())
	if err != nil {
		panic(err)
	}

	maxDB, err = sql.Open("mysql", maxMysqlConfig.Dsn())
	if err != nil{
		panic(err)
	}

	payDB, err = sql.Open("mysql",payMysqlConfig.Dsn())
	if err != nil{
		panic(err)
	}

}
