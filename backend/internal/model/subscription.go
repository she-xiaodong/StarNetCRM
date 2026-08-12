package model

import "time"

// Plan 订阅方案
type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanEnterprise Plan = "enterprise"
)

// SubscriptionStatus 订阅状态
type SubscriptionStatus string

const (
	SubActive    SubscriptionStatus = "active"
	SubTrial     SubscriptionStatus = "trial"
	SubExpired   SubscriptionStatus = "expired"
	SubCancelled SubscriptionStatus = "cancelled"
)

// Subscription 租户订阅记录
type Subscription struct {
	ID         string             `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID   string             `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	Plan       Plan               `gorm:"column:plan;type:varchar(20);default:free" json:"plan"`
	Status     SubscriptionStatus `gorm:"column:status;type:varchar(20);default:trial" json:"status"`
	MaxUsers   int                `gorm:"column:max_users;default:1" json:"max_users"`
	Price      float64            `gorm:"column:price;type:decimal(10,2);default:0" json:"price"`
	TrialEnds  *time.Time         `gorm:"column:trial_ends" json:"trial_ends"`
	ExpiresAt  *time.Time         `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt  time.Time          `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Subscription) TableName() string { return "subscriptions" }
