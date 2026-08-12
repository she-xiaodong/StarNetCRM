package main

import (
	"fmt"
	"log"
	"os"

	"github.com/starnet/crm/internal/config"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("用法: go run cmd/seed/main.go <用户名>\n示例: go run cmd/seed/main.go xiaodong")
	}
	username := os.Args[1]

	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 连接MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL失败: %v", err)
	}
	repomysql.DB = db

	// 3. 自动迁移（确保表存在）
	if err := repomysql.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 4. 录入356条联系人数据
	if err := seedContactsForUser(username); err != nil {
		log.Fatalf("录入联系人数��失败: %v", err)
	}
}
