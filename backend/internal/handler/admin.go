package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler { return &AdminHandler{} }

// ─── 统计概览 ───
func (h *AdminHandler) Stats(c *gin.Context) {
	var totalTenants, totalUsers int64
	repomysql.DB.Model(&model.Tenant{}).Count(&totalTenants)
	repomysql.DB.Model(&model.User{}).Count(&totalUsers)

	// 活跃租户
	var activeTenants int64
	repomysql.DB.Model(&model.Subscription{}).
		Where("status IN ?", []string{"active", "trial"}).
		Count(&activeTenants)

	dto.Success(c, gin.H{
		"total_tenants":  totalTenants,
		"active_tenants": activeTenants,
		"total_users":    totalUsers,
		"new_this_month": 0, // TODO: 按创建时间筛选
	})
}

// ─── 租户列表 ───
func (h *AdminHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	type TenantRow struct {
		model.Tenant
		AdminName   string  `json:"admin_name"`
		AdminPhone  string  `json:"admin_phone"`
		Plan        string  `json:"plan"`
		Status      string  `json:"status"`
		ExpiresAt   *string `json:"expires_at"`
		UserCount   int     `json:"user_count"`
	}

	var rows []TenantRow
	var total int64

	query := repomysql.DB.Table("tenants AS t").
		Select(`t.*, 
			u.name AS admin_name, u.phone AS admin_phone,
			COALESCE(s.plan, 'free') AS plan,
			COALESCE(s.status, 'trial') AS status,
			COALESCE(DATE_FORMAT(s.expires_at, '%Y-%m-%d'), '') AS expires_at,
			(SELECT COUNT(*) FROM users WHERE tenant_id = t.id) AS user_count`).
		Joins("LEFT JOIN users u ON u.tenant_id = t.id AND u.role = 'admin'").
		Joins("LEFT JOIN subscriptions s ON s.tenant_id = t.id")

	if keyword != "" {
		query = query.Where("t.name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("t.created_at DESC").Scan(&rows)

	dto.Success(c, gin.H{
		"list":  rows,
		"total": total,
	})
}

// ─── 开通租户 ───
func (h *AdminHandler) CreateTenant(c *gin.Context) {
	var req struct {
		CompanyName   string `json:"company_name" binding:"required"`
		AdminName     string `json:"admin_name" binding:"required"`
		AdminUsername string `json:"admin_username" binding:"required"`
		AdminPassword string `json:"admin_password" binding:"required,min=6"`
		AdminPhone    string `json:"admin_phone"`
		AdminEmail    string `json:"admin_email"`
		Plan          string `json:"plan" binding:"required"`
		TrialDays     int    `json:"trial_days"`
		MaxUsers      int    `json:"max_users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var existing model.User
	if err := repomysql.DB.Where("username = ?", req.AdminUsername).First(&existing).Error; err == nil {
		dto.BadRequest(c, "用户名已存在")
		return
	}

	tenantID := uuid.New().String()
	userID := uuid.New().String()
	subID := uuid.New().String()

	// 创建租户
	tenant := model.Tenant{
		ID:         tenantID,
		Name:       req.CompanyName,
		DeployMode: model.DeploySaaS,
	}
	if err := repomysql.DB.Create(&tenant).Error; err != nil {
		dto.InternalError(c, "创建租户失败: "+err.Error())
		return
	}

	// 创建管理员
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	admin := model.User{
		ID:           userID,
		TenantID:     tenantID,
		Username:     req.AdminUsername,
		PasswordHash: string(hash),
		Name:         req.AdminName,
		Phone:        req.AdminPhone,
		Email:        req.AdminEmail,
		Role:         model.RoleAdmin,
	}
	if err := repomysql.DB.Create(&admin).Error; err != nil {
		repomysql.DB.Delete(&tenant)
		dto.InternalError(c, "创建管理员失败: "+err.Error())
		return
	}

	// 创建订阅
	maxUsers := req.MaxUsers
	if maxUsers <= 0 {
		maxUsers = 5
	}
	sub := model.Subscription{
		ID:       subID,
		TenantID: tenantID,
		Plan:     model.Plan(req.Plan),
		Status:   model.SubActive,
		MaxUsers: maxUsers,
	}
	if req.TrialDays > 0 {
		sub.Status = model.SubTrial
	}
	if err := repomysql.DB.Create(&sub).Error; err != nil {
		dto.InternalError(c, "创建订阅失败: "+err.Error())
		return
	}

	dto.Success(c, gin.H{
		"tenant_id": tenantID,
		"user_id":   userID,
	})
}

// ─── 删除租户 ───
func (h *AdminHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		dto.BadRequest(c, "缺少租户ID")
		return
	}
	repomysql.DB.Where("tenant_id = ?", id).Delete(&model.User{})
	repomysql.DB.Where("tenant_id = ?", id).Delete(&model.Subscription{})
	repomysql.DB.Delete(&model.Tenant{}, "id = ?", id)
	dto.Success(c, nil)
}

// ─── 租户详情 ───
func (h *AdminHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	var tenant model.Tenant
	if err := repomysql.DB.First(&tenant, "id = ?", id).Error; err != nil {
		dto.NotFound(c, "租户不存在")
		return
	}

	var sub model.Subscription
	repomysql.DB.Where("tenant_id = ?", id).First(&sub)

	var userCount int64
	repomysql.DB.Model(&model.User{}).Where("tenant_id = ?", id).Count(&userCount)

	dto.Success(c, gin.H{
		"tenant":       tenant,
		"subscription": sub,
		"user_count":   userCount,
	})
}
