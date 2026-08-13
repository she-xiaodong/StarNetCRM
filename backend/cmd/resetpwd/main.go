package main

import (
	"fmt"
	"log"

	"github.com/starnet/crm/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL失败: %v", err)
	}

	// 生成 123456 的 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("生成哈希失败: %v", err)
	}

	// 自检：确保该哈希确实能匹配 123456
	if err := bcrypt.CompareHashAndPassword(hash, []byte("123456")); err != nil {
		log.Fatalf("哈希自检失败: %v", err)
	}
	fmt.Println("哈希自检通过：hash 正确对应 123456")

	// 参数化更新，避免 SQL 转义问题
	result := db.Exec("UPDATE users SET password_hash = ? WHERE username = ?", string(hash), "xiaodong")
	if result.Error != nil {
		log.Fatalf("更新密码失败: %v", result.Error)
	}
	fmt.Printf("更新成功：影响 %d 行\n", result.RowsAffected)
}
