package handler

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	reponeo4j "github.com/starnet/crm/internal/repository/neo4j"
	"github.com/starnet/crm/pkg/logger"
	"go.uber.org/zap"
)

// ContactHandler 联系人处理器
type ContactHandler struct{}

// NewContactHandler 创建联系人处理器
func NewContactHandler() *ContactHandler {
	return &ContactHandler{}
}

// List 联系人列表
func (h *ContactHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var req dto.ContactListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.BadRequest(c, "参数错误")
		return
	}
	req.Default()

	query := repomysql.DB.Where("tenant_id = ?", tenantID)
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR company LIKE ? OR title LIKE ?", keyword, keyword, keyword)
	}
	if req.Company != "" {
		query = query.Where("company LIKE ?", "%"+req.Company+"%")
	}
	if req.Department != "" {
		query = query.Where("department LIKE ?", "%"+req.Department+"%")
	}

	var total int64
	query.Model(&model.Contact{}).Count(&total)

	var contacts []model.Contact
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&contacts).Error; err != nil {
		dto.InternalError(c, "查询失败")
		return
	}

	dto.SuccessPage(c, dto.PageData{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     contacts,
	})
}

// Create 创建联系人
func (h *ContactHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	var req dto.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	personID := uuid.New().String()

	// 序列化标签
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		if b, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(b)
		}
	}

	// 写入MySQL
	contact := model.Contact{
		ID:         uuid.New().String(),
		TenantID:   tenantID.(string),
		PersonID:   personID,
		Name:       req.Name,
		Company:    req.Company,
		Title:      req.Title,
		Phone:      req.Phone,
		Email:      req.Email,
		Department: req.Department,
		Tags:       tagsJSON,
		CreatedBy:  userID.(string),
	}

	if err := repomysql.DB.Create(&contact).Error; err != nil {
		dto.InternalError(c, "创建联系人失败")
		return
	}

	// 写入Neo4j（如果可用）
	if reponeo4j.IsAvailable() {
		ctx := context.Background()
		session := reponeo4j.NewSession(ctx)
		defer session.Close(ctx)

		nodeProps := map[string]interface{}{
			"id":         personID,
			"name":       req.Name,
			"company":    req.Company,
			"title":      req.Title,
			"phone":      req.Phone,
			"email":      req.Email,
			"wecom_id":   "",
			"avatar":     "",
			"department": req.Department,
			"is_active":  true,
		}

		if err := reponeo4j.CreatePerson(ctx, session, nodeProps); err != nil {
			logger.Log.Warn("创建图谱节点失败（联系人已保存）", zap.String("person_id", personID), zap.Error(err))
		} else {
			// 确保当前用户也有Person节点
			uid := userID.(string)
			userName := ""
			var userModel model.User
			if err := repomysql.DB.Where("id = ?", uid).First(&userModel).Error; err == nil {
				userName = userModel.Name
			}
			if userName == "" {
				userName = uid
			}

			if err := reponeo4j.EnsurePerson(ctx, session, uid, userName); err != nil {
				logger.Log.Warn("为用户创建图谱节点失败", zap.String("user_id", uid), zap.Error(err))
			}

			// 创建 RELATES_TO 关系
			relProps := map[string]interface{}{
				"from_id":    uid,
				"to_id":      personID,
				"type":       "knows",
				"source":     "manual",
				"strength":   1.0,
				"valid_until": nil,
				"is_shared":  false,
				"created_by": uid,
			}
			if err := reponeo4j.CreateRelation(ctx, session, relProps); err != nil {
				logger.Log.Warn("创建人脉关系失败", zap.String("from", uid), zap.String("to", personID), zap.Error(err))
			}
		}
	}

	dto.Success(c, contact)
}

// Get 获取联系人详情
func (h *ContactHandler) Get(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var contact model.Contact
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		dto.NotFound(c, "联系人不存在")
		return
	}

	dto.Success(c, contact)
}

// Update 更新联系人
func (h *ContactHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var req dto.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败")
		return
	}

	var contact model.Contact
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		dto.NotFound(c, "联系人不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Company != "" {
		updates["company"] = req.Company
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Department != "" {
		updates["department"] = req.Department
	}
	if req.Tags != nil {
		if b, err := json.Marshal(req.Tags); err == nil {
			updates["tags"] = string(b)
		}
	}

	if len(updates) > 0 {
		if err := repomysql.DB.Model(&contact).Updates(updates).Error; err != nil {
			dto.InternalError(c, "更新失败")
			return
		}
	}

	dto.Success(c, contact)
}

// Delete 删除联系人
func (h *ContactHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var contact model.Contact
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		dto.NotFound(c, "联系人不存在")
		return
	}

	if err := repomysql.DB.Delete(&contact).Error; err != nil {
		dto.InternalError(c, "删除失败")
		return
	}

	// TODO: 同时删除Neo4j中的节点

	dto.Success(c, gin.H{"message": "删除成功"})
}
