package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
)

// TagHandler 标签管理
type TagHandler struct{}

// NewTagHandler 构造
func NewTagHandler() *TagHandler {
	return &TagHandler{}
}

// List 标签列表（按tenant_id隔离）
func (h *TagHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	var tags []model.Tag
	if err := repomysql.DB.Where("tenant_id = ?", tenantID).
		Order("sort_order ASC, created_at DESC").
		Find(&tags).Error; err != nil {
		dto.InternalError(c, "获取标签列表失败")
		return
	}
	dto.Success(c, tags)
}

// Create 创建标签
func (h *TagHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var req struct {
		Name      string `json:"name" binding:"required"`
		Color     string `json:"color"`
		Type      string `json:"type"`       // "relationship" | "attribute"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	tag := model.Tag{
		ID:        uuid.New().String(),
		TenantID:  tenantID.(string),
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		SortOrder: 0,
	}
	if tg := repomysql.DB.Create(&tag); tg.Error != nil {
		dto.InternalError(c, "创建标签失败")
		return
	}
	dto.Success(c, tag)
}

// Update 更新标签
func (h *TagHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var req struct {
		Name      string `json:"name"`
		Color     string `json:"color"`
		Type      string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	var tag model.Tag
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&tag).Error; err != nil {
		dto.NotFound(c, "标签不存在")
		return
	}

	if req.Name != "" {
		tag.Name = req.Name
	}
	if req.Color != "" {
		tag.Color = req.Color
	}
	if req.Type != "" {
		tag.Type = req.Type
	}
	repomysql.DB.Save(&tag)
	dto.Success(c, tag)
}

// Delete 删除标签
func (h *TagHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var tag model.Tag
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&tag).Error; err != nil {
		dto.NotFound(c, "标签不存在")
		return
	}

	repomysql.DB.Delete(&tag)
	dto.Success(c, nil)
}
