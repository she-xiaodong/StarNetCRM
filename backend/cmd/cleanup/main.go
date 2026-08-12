package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/starnet_crm?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("DB ERROR:", err)
		return
	}

	tables := []string{"contacts", "tags", "referrals", "operation_logs", "users", "tenants"}
	for _, t := range tables {
		if db.Migrator().HasTable(t) {
			if err := db.Migrator().DropTable(t); err != nil {
				fmt.Printf("Drop %s FAIL: %v\n", t, err)
			} else {
				fmt.Printf("Drop %s OK\n", t)
			}
		} else {
			fmt.Printf("Table %s not found\n", t)
		}
	}
	fmt.Println("cleanup done")
}
