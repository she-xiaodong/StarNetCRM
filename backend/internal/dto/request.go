package dto

// LoginRequest 登录请求（独立模式）
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求（独立模式）
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// WecomAuthRequest 企微授权请求
type WecomAuthRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// CreateContactRequest 创建联系人请求
type CreateContactRequest struct {
	Name       string   `json:"name" binding:"required"`
	Company    string   `json:"company"`
	Title      string   `json:"title"`
	Phone      string   `json:"phone"`
	Email      string   `json:"email"`
	Department string   `json:"department"`
	Tags       []string `json:"tags"`
	ReferrerID string   `json:"referrer_id"` // 引荐人PersonID（空=直接人脉）
	Note       string   `json:"note"`

	// 关系信息（可选，同时创建"我→对方"关系记录）
	RelationType     string   `json:"relation_type"`     // 主关系类型：colleague/customer/...
	RelationTags     []string `json:"relation_tags"`     // 关系标签
	RelationStrength int      `json:"relation_strength"` // 亲密度 1-10，缺省 5
}

// UpdateContactRequest 更新联系人请求
type UpdateContactRequest struct {
	Name       string   `json:"name"`
	Company    string   `json:"company"`
	Title      string   `json:"title"`
	Phone      string   `json:"phone"`
	Email      string   `json:"email"`
	Department string   `json:"department"`
	Tags       []string `json:"tags"`
	ReferrerID *string  `json:"referrer_id"` // 引荐人PersonID，传空字符串表示清除引荐人
	Note       string   `json:"note"`

	// 关系信息（可选，同时创建/更新"我→对方"关系记录）
	RelationType     *string  `json:"relation_type"`
	RelationTags     []string `json:"relation_tags"`
	RelationStrength *int     `json:"relation_strength"`
}

// ContactListRequest 联系人列表查询
type ContactListRequest struct {
	PageRequest
	Company    string `form:"company" json:"company"`
	Department string `form:"department" json:"department"`
	Tag        string `form:"tag" json:"tag"`
}

// RelationListRequest 关系列表查询
type RelationListRequest struct {
	PageRequest
	PersonID string `form:"person_id" json:"person_id"`
	RelType  string `form:"rel_type" json:"rel_type"`
}
