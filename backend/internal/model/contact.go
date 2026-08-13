package model

import (
	"time"
)

// Contact 联系人扩展表（MySQL中Person的镜像，用于复杂条件查询）
type Contact struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID   string    `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_tenant" json:"tenant_id"`
	PersonID   string    `gorm:"column:person_id;type:varchar(36);not null" json:"person_id"` // 对应Neo4j节点ID
	Name       string    `gorm:"column:name;type:varchar(50)" json:"name"`
	Company    string    `gorm:"column:company;type:varchar(100);index:idx_company" json:"company"`
	Title      string    `gorm:"column:title;type:varchar(100)" json:"title"`
	Phone      string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email      string    `gorm:"column:email;type:varchar(100)" json:"email"`
	Department string    `gorm:"column:department;type:varchar(100)" json:"department"`
	Tags       string    `gorm:"column:tags;type:json" json:"tags"`                    // JSON数组
	ReferrerID string    `gorm:"column:referrer_id;type:varchar(36);index:idx_referrer" json:"referrer_id"` // 引荐人PersonID（空=直接人脉）
	CreatedBy  string    `gorm:"column:created_by;type:varchar(36)" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// 引荐路径（运行时计算，不存储MySQL）：["我","张三","李四"]
	ReferrerPath     []string `gorm:"-" json:"referrer_path,omitempty"`
	ReferrerPathText string   `gorm:"-" json:"referrer_path_text,omitempty"`

	// Neo4j中Person的额外属性（不存储MySQL，仅用于传输）
	WecomID  string `gorm:"-" json:"wecom_id"`
	IsActive bool   `gorm:"-" json:"is_active"`
}

func (Contact) TableName() string { return "contacts" }

// Tag 标签表
type Tag struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;type:varchar(36);not null;uniqueIndex:uk_tenant_tag" json:"tenant_id"`
	Name      string    `gorm:"column:name;type:varchar(50);not null;uniqueIndex:uk_tenant_tag" json:"name"`
	Type      string    `gorm:"column:type;type:varchar(50);default:'relationship'" json:"type"` // relationship | attribute
	Color     string    `gorm:"column:color;type:varchar(20)" json:"color"`
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Tag) TableName() string { return "tags" }

// Referral 引荐记录表
type Referral struct {
	ID              string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	FromPersonID    string    `gorm:"column:from_person_id;type:varchar(36);not null;index:idx_from" json:"from_person_id"`
	MiddlePersonID  string    `gorm:"column:middle_person_id;type:varchar(36);not null;index:idx_middle" json:"middle_person_id"`
	TargetPersonID  string    `gorm:"column:target_person_id;type:varchar(36);not null;index:idx_target" json:"target_person_id"`
	PathJSON        string    `gorm:"column:path_json;type:json" json:"path_json"`
	Status          string    `gorm:"column:status;type:enum('draft','sent','accepted','rejected','connected');default:draft" json:"status"`
	MessageTemplate string    `gorm:"column:message_template;type:text" json:"message_template"`
	SentAt          *time.Time `gorm:"column:sent_at" json:"sent_at"`
	RespondedAt     *time.Time `gorm:"column:responded_at" json:"responded_at"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Referral) TableName() string { return "referrals" }

// OperationLog 操作日志
type OperationLog struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID   string    `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_tenant_user" json:"tenant_id"`
	UserID     string    `gorm:"column:user_id;type:varchar(36);not null;index:idx_tenant_user" json:"user_id"`
	Action     string    `gorm:"column:action;type:varchar(50)" json:"action"`
	TargetType string    `gorm:"column:target_type;type:varchar(50)" json:"target_type"`
	TargetID   string    `gorm:"column:target_id;type:varchar(36)" json:"target_id"`
	Detail     string    `gorm:"column:detail;type:json" json:"detail"`
	IP         string    `gorm:"column:ip;type:varchar(50)" json:"ip"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime;index:idx_created" json:"created_at"`
}

func (OperationLog) TableName() string { return "operation_logs" }
