package model

import (
	"encoding/json"
	"time"
)

// Relation 关系表：A(我) → B(对方) 的关系记录。
// 亲密度(strength)、关系标签(tags)都挂在"关系"上，而非单个联系人。
// 方向为单向：from_person_id 发起方，to_person_id 接收方。
type Relation struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;type:varchar(36);not null;index:idx_tenant_from" json:"tenant_id"`
	FromPersonID string    `gorm:"column:from_person_id;type:varchar(36);not null;index:idx_tenant_from" json:"from_person_id"` // 我
	ToPersonID   string    `gorm:"column:to_person_id;type:varchar(36);not null;index:idx_to" json:"to_person_id"`              // 对方
	Type         string    `gorm:"column:type;type:varchar(50);not null;default:'friend'" json:"type"`                          // 主关系类型：colleague/customer/alumni/friend/custom...
	TagsJSON     string    `gorm:"column:tags;type:json" json:"-"`                                                             // 关系标签（JSON数组）
	Strength     int       `gorm:"column:strength;type:int;not null;default:5" json:"strength"`                                 // 亲密度 1-10
	Note         string    `gorm:"column:note;type:text" json:"note"`
	CreatedBy    string    `gorm:"column:created_by;type:varchar(36)" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// 关系标签（运行时解析，不存储 MySQL，仅用于传输）
	Tags []string `gorm:"-" json:"tags,omitempty"`

	// 对方/发起方联系人展示信息（运行时填充，不存储 MySQL）
	ToName    string `gorm:"-" json:"to_name,omitempty"`
	ToCompany string `gorm:"-" json:"to_company,omitempty"`
	ToTitle   string `gorm:"-" json:"to_title,omitempty"`
	FromName  string `gorm:"-" json:"from_name,omitempty"`
}

func (Relation) TableName() string { return "relations" }

// UnmarshalTags 将 TagsJSON 解析为 Tags
func (r *Relation) UnmarshalTags() {
	r.Tags = nil
	if r.TagsJSON == "" {
		return
	}
	_ = json.Unmarshal([]byte(r.TagsJSON), &r.Tags)
}

// MarshalTags 将 Tags 序列化为 TagsJSON
func (r *Relation) MarshalTags() {
	if len(r.Tags) == 0 {
		r.TagsJSON = "[]"
		return
	}
	if b, err := json.Marshal(r.Tags); err == nil {
		r.TagsJSON = string(b)
	}
}

// CreateRelationRequest 创建关系请求
type CreateRelationRequest struct {
	FromPersonID string   `json:"from_person_id"`                   // 我（缺省由后端用当前用户填充）
	ToPersonID   string   `json:"to_person_id" binding:"required"`  // 对方
	Type         string   `json:"type"`                             // 关系类型：colleague/customer/...
	Source       string   `json:"source"`                           // 建立来源
	Tags         []string `json:"tags"`                             // 关系标签
	Strength     int      `json:"strength" binding:"min=1,max=10"`  // 亲密度 1-10
	ValidUntil   string   `json:"valid_until"`
	IsShared     string   `json:"is_shared"`                        // private/department/company
	Note         string   `json:"note"`
}
