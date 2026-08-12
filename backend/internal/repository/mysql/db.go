package mysql

import (
	"fmt"
	"time"

	"github.com/starnet/crm/internal/config"
	"github.com/starnet/crm/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化MySQL连接
func Init(cfg config.MySQLConfig) error {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect MySQL: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

// AutoMigrate 自动迁移
func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Contact{},
		&model.Tag{},
		&model.Referral{},
		&model.OperationLog{},
		&model.Subscription{},
	)
}

// Close 关闭连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

// SeedDefaultTenant 创建默认租户（独立部署模式）
func SeedDefaultTenant() (*model.Tenant, error) {
	var tenant model.Tenant
	result := DB.Where("name = ?", "default").First(&tenant)
	if result.Error == nil {
		return &tenant, nil
	}

	tenant = model.Tenant{
		ID:         "tenant-default",
		Name:       "default",
		DeployMode: model.DeployStandalone,
		Config:     "{}",
	}
	if err := DB.Create(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}
