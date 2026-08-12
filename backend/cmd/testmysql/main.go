package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	passwords := []string{"root", "123456", "admin", "mysql", "phpstudy"}

	for _, pass := range passwords {
		dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/", pass)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Trying pass=%s: OPEN ERR: %v\n", pass, err)
			continue
		}
		err = db.Ping()
		db.Close()
		if err != nil {
			fmt.Printf("Trying pass=%s: FAIL\n", pass)
			continue
		}
		fmt.Printf("SUCCESS! Password is: %s\n", pass)

		// 创建数据库
		db2, _ := sql.Open("mysql", dsn)
		defer db2.Close()
		_, err = db2.Exec("CREATE DATABASE IF NOT EXISTS starnet_crm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
		if err != nil {
			fmt.Printf("  CREATE DB ERR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  Database 'starnet_crm' created!")
		os.Exit(0)
	}

	fmt.Println("ALL FAILED - need correct password")
	os.Exit(1)
}
