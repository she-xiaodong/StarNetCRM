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

// RelationHandler 关系处理器：管理"我→对方"的关系（亲密度 + 关系标签）
type RelationHandler struct{}

// NewRelationHandler 创建关系处理器
func NewRelationHandler() *RelationHandler {
	return &RelationHandler{}
}

// List 关系列表（按 from_person_id 查询，即"我"发起的关系）
func (h *RelationHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var req dto.RelationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.BadRequest(c, "参数错误")
		return
	}
	req.Default()

	// personID 缺省时取当前用户（"我"）
	personID := req.PersonID
	if personID == "" {
		userID, _ := c.Get("user_id")
		personID = userID.(string)
	}

	query := repomysql.DB.Where("tenant_id = ? AND from_person_id = ?", tenantID, personID)
	if req.RelType != "" {
		query = query.Where("type = ?", req.RelType)
	}

	var total int64
	query.Model(&model.Relation{}).Count(&total)

	var relations []model.Relation
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("updated_at DESC").Find(&relations).Error; err != nil {
		dto.InternalError(c, "查询失败")
		return
	}

	// 解析标签 JSON
	ptrList := make([]*model.Relation, len(relations))
	for i := range relations {
		relations[i].UnmarshalTags()
		ptrList[i] = &relations[i]
	}
	resolveRelationNames(c, ptrList)
	resolveRelationTagNames(c, ptrList)

	dto.SuccessPage(c, dto.PageData{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     relations,
	})
}

// Create 创建关系（单向：我 → 对方）
func (h *RelationHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	var req model.CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	// 校验亲密度范围
	if req.Strength < 1 || req.Strength > 10 {
		dto.BadRequest(c, "亲密度需在 1-10 之间")
		return
	}

	// 校验对方存在且属于当前租户
	var toContact model.Contact
	if err := repomysql.DB.Where("tenant_id = ? AND person_id = ?", tenantID, req.ToPersonID).First(&toContact).Error; err != nil {
		dto.BadRequest(c, "对方联系人不存在")
		return
	}

	// FromPersonID 缺省时取当前用户（"我"）
	fromPersonID := req.FromPersonID
	if fromPersonID == "" {
		fromPersonID = userID.(string)
	}

	// 避免重复：同一对 (from,to) 已有关系时直接更新
	var existing model.Relation
	if err := repomysql.DB.Where("tenant_id = ? AND from_person_id = ? AND to_person_id = ?", tenantID, fromPersonID, req.ToPersonID).First(&existing).Error; err == nil {
		existing.Type = req.Type
		existing.Strength = req.Strength
		existing.Note = req.Note
		existing.Tags = req.Tags
		existing.MarshalTags()
		if err := repomysql.DB.Save(&existing).Error; err != nil {
			dto.InternalError(c, "更新关系失败")
			return
		}
		existing.UnmarshalTags()
		h.syncToNeo4j(c, &existing)
		resolveRelationTagNames(c, []*model.Relation{&existing})
		dto.Success(c, existing)
		return
	}

	relation := model.Relation{
		ID:           uuid.New().String(),
		TenantID:     tenantID.(string),
		FromPersonID: fromPersonID,
		ToPersonID:   req.ToPersonID,
		Type:         req.Type,
		Strength:     req.Strength,
		Note:         req.Note,
		CreatedBy:    userID.(string),
	}
	relation.Tags = req.Tags
	relation.MarshalTags()

	if err := repomysql.DB.Create(&relation).Error; err != nil {
		dto.InternalError(c, "创建关系失败")
		return
	}
	relation.UnmarshalTags()

	// 同步写入 Neo4j（如果可用）
	h.syncToNeo4j(c, &relation)
	resolveRelationTagNames(c, []*model.Relation{&relation})

	dto.Success(c, relation)
}

// Update 更新关系（亲密度/标签/类型/备注）
func (h *RelationHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var relation model.Relation
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&relation).Error; err != nil {
		dto.NotFound(c, "关系不存在")
		return
	}

	var req struct {
		Type     string   `json:"type"`
		Tags     []string `json:"tags"`
		Strength *int     `json:"strength"`
		Note     string   `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败")
		return
	}

	updates := map[string]interface{}{}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Tags != nil {
		if b, err := jsonMarshal(req.Tags); err == nil {
			updates["tags"] = string(b)
		}
	}
	if req.Strength != nil {
		if *req.Strength < 1 || *req.Strength > 10 {
			dto.BadRequest(c, "亲密度需在 1-10 之间")
			return
		}
		updates["strength"] = *req.Strength
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}

	if len(updates) > 0 {
		if err := repomysql.DB.Model(&relation).Updates(updates).Error; err != nil {
			dto.InternalError(c, "更新失败")
			return
		}
		repomysql.DB.First(&relation, "id = ?", id)
	}
	relation.UnmarshalTags()
	resolveRelationTagNames(c, []*model.Relation{&relation})

	dto.Success(c, relation)
}

// Delete 删除关系
func (h *RelationHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	id := c.Param("id")

	var relation model.Relation
	if err := repomysql.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&relation).Error; err != nil {
		dto.NotFound(c, "关系不存在")
		return
	}

	if err := repomysql.DB.Delete(&relation).Error; err != nil {
		dto.InternalError(c, "删除失败")
		return
	}

	// 同步删除 Neo4j 关系（如果可用）
	if reponeo4j.IsAvailable() {
		ctx := context.Background()
		session := reponeo4j.NewSession(ctx)
		defer session.Close(ctx)
		if err := reponeo4j.DeleteRelation(ctx, session, relation.FromPersonID, relation.ToPersonID); err != nil {
			logger.Log.Warn("删除Neo4j关系失败", zap.String("relation_id", id), zap.Error(err))
		}
	}

	dto.Success(c, gin.H{"message": "删除成功"})
}

// UpsertMyRelation 创建或更新"我→对方"关系记录。
// 供联系人创建/编辑时调用：有则更新（亲密度/标签/类型），无则新建。
func UpsertMyRelation(tenantID, fromPersonID, toPersonID, relType string, tags []string, strength int, c *gin.Context) model.Relation {
	if relType == "" {
		relType = "friend"
	}
	if strength < 1 || strength > 10 {
		strength = 5
	}

	var existing model.Relation
	if err := repomysql.DB.Where("tenant_id = ? AND from_person_id = ? AND to_person_id = ?", tenantID, fromPersonID, toPersonID).First(&existing).Error; err == nil {
		existing.Type = relType
		existing.Strength = strength
		existing.Tags = tags
		existing.MarshalTags()
		if err := repomysql.DB.Save(&existing).Error; err != nil {
			logger.Log.Warn("更新关系失败", zap.String("from", fromPersonID), zap.String("to", toPersonID), zap.Error(err))
		}
		existing.UnmarshalTags()
		return existing
	}

	relation := model.Relation{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		FromPersonID: fromPersonID,
		ToPersonID:   toPersonID,
		Type:         relType,
		Strength:     strength,
		CreatedBy:    fromPersonID,
	}
	relation.Tags = tags
	relation.MarshalTags()
	if err := repomysql.DB.Create(&relation).Error; err != nil {
		logger.Log.Warn("创建关系失败", zap.String("from", fromPersonID), zap.String("to", toPersonID), zap.Error(err))
		return relation
	}
	relation.UnmarshalTags()
	return relation
}

// syncToNeo4j 将 MySQL 关系同步写入 Neo4j（双写，失败仅告警不阻断）
func (h *RelationHandler) syncToNeo4j(c *gin.Context, relation *model.Relation) {
	if !reponeo4j.IsAvailable() {
		return
	}
	ctx := context.Background()
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	// 确保两端 Person 节点存在
	_ = reponeo4j.EnsurePerson(ctx, session, relation.FromPersonID, relation.FromPersonID)
	_ = reponeo4j.EnsurePerson(ctx, session, relation.ToPersonID, relation.ToPersonID)

	relProps := map[string]interface{}{
		"from_id":     relation.FromPersonID,
		"to_id":       relation.ToPersonID,
		"relation_id": relation.ID,
		"type":        relation.Type,
		"source":      "manual",
		"strength":    relation.Strength,
		"valid_until": "2099-12-31",
		"is_shared":   model.VisibilityPrivate,
		"created_by":  relation.CreatedBy,
	}
	if err := reponeo4j.CreateRelation(ctx, session, relProps); err != nil {
		logger.Log.Warn("同步关系到Neo4j失败", zap.String("relation_id", relation.ID), zap.Error(err))
	}
}

// resolveRelationNames 将关系对方 PersonID 解析为联系人名称（便于前端展示）
func resolveRelationNames(c *gin.Context, relations []*model.Relation) {
	if len(relations) == 0 {
		return
	}
	tenantID, _ := c.Get("tenant_id")

	personIDs := make([]string, 0, len(relations))
	for _, r := range relations {
		if r.ToPersonID != "" {
			personIDs = append(personIDs, r.ToPersonID)
		}
		if r.FromPersonID != "" {
			personIDs = append(personIDs, r.FromPersonID)
		}
	}
	if len(personIDs) == 0 {
		return
	}

	var contacts []model.Contact
	if err := repomysql.DB.Where("tenant_id = ? AND person_id IN ?", tenantID, personIDs).Find(&contacts).Error; err != nil {
		return
	}
	idToContact := make(map[string]*model.Contact)
	for i := range contacts {
		idToContact[contacts[i].PersonID] = &contacts[i]
	}

	for _, r := range relations {
		if toContact, ok := idToContact[r.ToPersonID]; ok {
			r.ToName = toContact.Name
			r.ToCompany = toContact.Company
			r.ToTitle = toContact.Title
		}
		if fromContact, ok := idToContact[r.FromPersonID]; ok {
			r.FromName = fromContact.Name
		}
	}
}

// resolveRelationTagNames 将关系 tags 中的标签 ID 解析为标签名称（与联系人标签处理保持一致）
func resolveRelationTagNames(c *gin.Context, relations []*model.Relation) {
	if len(relations) == 0 {
		return
	}
	tenantID, _ := c.Get("tenant_id")

	// 收集所有标签 ID
	idSet := make(map[string]struct{})
	for _, r := range relations {
		for _, t := range r.Tags {
			idSet[t] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return
	}

	// 查询标签表，建立 ID → 名称映射
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

	// 将 ID 替换为名称（无法匹配的保留原值，兼容已是名称的情况）
	for _, r := range relations {
		for i, t := range r.Tags {
			if name, ok := idToName[t]; ok {
				r.Tags[i] = name
			}
		}
	}
}

// jsonMarshal 便捷 JSON 序列化
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// relationTypeLabel 关系类型中文名映射
var relationTypeLabel = map[string]string{
	"colleague": "同事",
	"manager":   "上下级",
	"customer":  "客户",
	"partner":   "合作伙伴",
	"alumni":    "校友",
	"friend":    "朋友",
	"referral":  "引荐",
	"custom":    "自定义",
}

// RelationTypeName 关系类型中文名
func RelationTypeName(t string) string {
	if name, ok := relationTypeLabel[t]; ok {
		return name
	}
	// 直接返回原值（可能是自定义类型或中文）
	return strings.TrimSpace(t)
}
