package model

import (
	"time"
)

// Role 用户角色
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleMember  Role = "member"
)

// DeployMode 部署模式
type DeployMode string

const (
	DeploySaaS       DeployMode = "saas"
	DeployStandalone DeployMode = "standalone"
)

// Tenant 租户（企业）
type Tenant struct {
	ID         string     `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	Name       string     `gorm:"column:name;type:varchar(100);not null" json:"name"`
	CorpID     string     `gorm:"column:corp_id;type:varchar(100)" json:"corp_id"` // 企微企业ID
	DeployMode DeployMode `gorm:"column:deploy_mode;type:enum('saas','standalone');default:saas" json:"deploy_mode"`
	Config     string     `gorm:"column:config;type:json" json:"config"` // JSON配置
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }

// User 用户表
type User struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_tenant_id" json:"tenant_id"`
	WecomUserID  *string   `gorm:"column:wecom_user_id;type:varchar(100);uniqueIndex:uk_tenant_wecom" json:"wecom_user_id"` // 企微UserID (允许NULL避免唯一索引空字符串冲突)
	Username     string    `gorm:"column:username;type:varchar(50)" json:"username"`       // 独立模式账号
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255)" json:"-"`         // 独立模式密码（不序列化）
	Name         string    `gorm:"column:name;type:varchar(50)" json:"name"`
	Phone        string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email        string    `gorm:"column:email;type:varchar(100)" json:"email"`
	Role         Role      `gorm:"column:role;type:enum('admin','manager','member');default:member" json:"role"`
	Department   string    `gorm:"column:department;type:varchar(100)" json:"department"`
	Avatar       string    `gorm:"column:avatar;type:varchar(255)" json:"avatar"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }
