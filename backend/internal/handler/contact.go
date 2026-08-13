package handler

import (
	"context"
	"encoding/json"
	"strings"

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

	// 将 tags 中的标签 ID 解析为名称（兼容已存名称的情况）
	ptrContacts := make([]*model.Contact, len(contacts))
	for i := range contacts {
		ptrContacts[i] = &contacts[i]
	}
	resolveTagNames(c, ptrContacts)
	fillReferrerPaths(c, ptrContacts)

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

	// 校验引荐人属于当前租户且不能引荐自己
	if req.ReferrerID != "" {
		if req.ReferrerID == personID {
			dto.BadRequest(c, "引荐人不能是自己")
			return
		}
		var referrer model.Contact
		if err := repomysql.DB.Where("tenant_id = ? AND person_id = ?", tenantID, req.ReferrerID).First(&referrer).Error; err != nil {
			dto.BadRequest(c, "引荐人不存在")
			return
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
		ReferrerID: req.ReferrerID,
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

	// 同步创建"我→对方"关系记录（亲密度/关系标签），若指定了关系信息
	if req.RelationType != "" || len(req.RelationTags) > 0 || req.RelationStrength > 0 {
		strength := req.RelationStrength
		if strength == 0 {
			strength = 5
		}
		UpsertMyRelation(tenantID.(string), userID.(string), personID, req.RelationType, req.RelationTags, strength, c)
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

	// 将 tags 中的标签 ID 解析为名称
	resolveTagNames(c, []*model.Contact{&contact})
	fillReferrerPaths(c, []*model.Contact{&contact})

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
	if req.ReferrerID != nil {
		if *req.ReferrerID != "" && *req.ReferrerID != contact.PersonID {
			var referrer model.Contact
			if err := repomysql.DB.Where("tenant_id = ? AND person_id = ?", tenantID, *req.ReferrerID).First(&referrer).Error; err != nil {
				dto.BadRequest(c, "引荐人不存在")
				return
			}
			updates["referrer_id"] = *req.ReferrerID
		} else {
			updates["referrer_id"] = ""
		}
	}

	if len(updates) > 0 {
		if err := repomysql.DB.Model(&contact).Updates(updates).Error; err != nil {
			dto.InternalError(c, "更新失败")
			return
		}
	}

	// 同步更新"我→对方"关系记录（若传了关系信息）
	if req.RelationType != nil || req.RelationTags != nil || req.RelationStrength != nil {
		relType := ""
		if req.RelationType != nil {
			relType = *req.RelationType
		}
		tags := req.RelationTags
		if tags == nil {
			tags = []string{}
		}
		strength := 0
		if req.RelationStrength != nil {
			strength = *req.RelationStrength
		}
		userID, _ := c.Get("user_id")
		UpsertMyRelation(tenantID.(string), userID.(string), contact.PersonID, relType, tags, strength, c)
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

// resolveTagNames 将联系人 tags 中的标签 ID 解析为名称（兼容已直接存名称的情况）
func resolveTagNames(c *gin.Context, contacts []*model.Contact) {
	if len(contacts) == 0 {
		return
	}
	tenantID, _ := c.Get("tenant_id")

	// 收集所有 tag 引用
	idSet := map[string]struct{}{}
	for _, contact := range contacts {
		var tags []string
		if err := json.Unmarshal([]byte(contact.Tags), &tags); err != nil || len(tags) == 0 {
			continue
		}
		for _, t := range tags {
			idSet[t] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return
	}

	// 查询标签表，建立 ID -> 名称 映射
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var tags []model.Tag
	if err := repomysql.DB.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&tags).Error; err != nil {
		return
	}
	idToName := make(map[string]string)
	for _, tag := range tags {
		idToName[tag.ID] = tag.Name
	}

	// 替换 tag ID 为名称，未命中的保留原值（已是名称）
	for _, contact := range contacts {
		var tagList []string
		if err := json.Unmarshal([]byte(contact.Tags), &tagList); err != nil || len(tagList) == 0 {
			continue
		}
		changed := false
		for i, t := range tagList {
			if name, ok := idToName[t]; ok {
				tagList[i] = name
				changed = true
			}
		}
		if changed {
			if b, err := json.Marshal(tagList); err == nil {
				contact.Tags = string(b)
			}
		}
	}
}

// fillReferrerPaths 计算联系人的引荐路径（沿 referrer_id 逐级上溯到当前用户）
// 直接人脉：["我"]；多级人脉：["我","张三","李四"]，文本格式 "我 → 张三 → 李四"
func fillReferrerPaths(c *gin.Context, contacts []*model.Contact) {
	if len(contacts) == 0 {
		return
	}
	tenantID, _ := c.Get("tenant_id")

	// 查询该租户全量联系人，建立 PersonID -> Contact 映射
	var all []model.Contact
	if err := repomysql.DB.Where("tenant_id = ?", tenantID).Find(&all).Error; err != nil {
		return
	}
	byPersonID := make(map[string]*model.Contact, len(all))
	for i := range all {
		byPersonID[all[i].PersonID] = &all[i]
	}

	for _, contact := range contacts {
		// 沿 referrer 链从联系人上溯，收集 [联系人, 引荐人, 引荐人的引荐人...]
		chain := []string{contact.Name}
		pid := contact.ReferrerID
		visited := map[string]bool{contact.PersonID: true}
		for pid != "" && len(chain) < 10 {
			if visited[pid] {
				break
			}
			visited[pid] = true
			referrer, ok := byPersonID[pid]
			if !ok {
				break
			}
			chain = append(chain, referrer.Name)
			pid = referrer.ReferrerID
		}

		if len(chain) == 1 {
			// 直接人脉（一级）
			contact.ReferrerPath = []string{"我"}
			contact.ReferrerPathText = "我"
			continue
		}
		// 反转链：["李四","张三"] -> ["张三","李四"]，再拼接"我"
		for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
			chain[i], chain[j] = chain[j], chain[i]
		}
		contact.ReferrerPath = append([]string{"我"}, chain...)
		contact.ReferrerPathText = strings.Join(contact.ReferrerPath, " → ")
	}
}
