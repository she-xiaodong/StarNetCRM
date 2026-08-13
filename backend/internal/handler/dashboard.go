package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	reponeo4j "github.com/starnet/crm/internal/repository/neo4j"
)

// DashboardHandler 首页看板处理器
type DashboardHandler struct{}

// NewDashboardHandler 创建首页看板处理器
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// Stats 首页统计数据
func (h *DashboardHandler) Stats(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)

	// ─── 联系人数 ───
	var totalContacts int64
	repomysql.DB.Model(&model.Contact{}).Where("tenant_id = ?", tenantIDStr).Count(&totalContacts)

	// ─── 关系连结数：优先从图谱统计，Neo4j 不可用时降级为联系人数量 ───
	var totalRelations int64 = totalContacts
	if reponeo4j.IsAvailable() {
		ctx := context.Background()
		session := reponeo4j.NewSession(ctx)
		defer session.Close(ctx)
		result, err := session.Run(ctx,
			"MATCH (:Person)-[r:RELATES_TO]->(:Person) RETURN count(r) AS cnt", nil)
		if err == nil && result.Next(ctx) {
			if v, ok := result.Record().Get("cnt"); ok {
				switch n := v.(type) {
				case int64:
					totalRelations = n
				case float64:
					totalRelations = int64(n)
				}
			}
		}
	}

	// ─── 引荐进行中（草稿/已发送）───
	var activeReferrals int64
	repomysql.DB.Model(&model.Referral{}).
		Where("tenant_id = ? AND status IN ?", tenantIDStr, []string{"draft", "sent"}).
		Count(&activeReferrals)

	// ─── 标签数量 ───
	var tagCount int64
	repomysql.DB.Model(&model.Tag{}).Where("tenant_id = ?", tenantIDStr).Count(&tagCount)

	// ─── 人脉评分：联系规模(0-60) + 引荐活跃度(0-30) + 标签覆盖(0-10) ───
	networkScore := 0
	if totalContacts > 0 {
		// 联系规模：每 10 人 2 分，上限 60
		scale := int(totalContacts/10) * 2
		if scale > 60 {
			scale = 60
		}
		networkScore += scale
	}
	// 引荐活跃度：每条进行中引荐 5 分，上限 30
	active := int(activeReferrals) * 5
	if active > 30 {
		active = 30
	}
	networkScore += active
	// 标签覆盖：每个标签 1 分，上限 10
	if tagCount > 10 {
		tagCount = 10
	}
	networkScore += int(tagCount)
	if networkScore > 100 {
		networkScore = 100
	}

	// ─── 最近联系人（最近 5 条，标签解析为名称）───
	var recentContacts []model.Contact
	repomysql.DB.Where("tenant_id = ?", tenantIDStr).
		Order("created_at DESC").Limit(5).Find(&recentContacts)
	ptrContacts := make([]*model.Contact, len(recentContacts))
	for i := range recentContacts {
		ptrContacts[i] = &recentContacts[i]
	}
	resolveTagNames(c, ptrContacts)

	dto.Success(c, gin.H{
		"total_contacts":   totalContacts,
		"total_relations":  totalRelations,
		"active_referrals": activeReferrals,
		"network_score":    networkScore,
		"recent_contacts":  recentContacts,
	})
}

// ─── Analytics 人脉分析 ─────────────────────────────────────────

func (h *DashboardHandler) Analytics(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)
	ctx := context.Background()

	// ─── 1. 活跃关系链数：优先图谱关系数，降级为联系人总数 ───
	activeRelations := h.countActiveRelations(ctx, tenantIDStr)

	// ─── 2. 本周新增联系人 ───
	var weekNewContacts int64
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	repomysql.DB.Model(&model.Contact{}).
		Where("tenant_id = ? AND created_at >= ?", tenantIDStr, weekStart).
		Count(&weekNewContacts)

	// ─── 3. 平均路径长度（估算值）───
	var totalContacts int64
	repomysql.DB.Model(&model.Contact{}).Where("tenant_id = ?", tenantIDStr).Count(&totalContacts)
	avgPathLength := 1.0
	if totalContacts > 0 {
		avgPathLength = 1.0 + float64(activeRelations)/float64(totalContacts)
		if avgPathLength > 6.0 {
			avgPathLength = 6.0
		}
	}
	avgPathLength = math.Round(avgPathLength*10) / 10

	// ─── 4. 孤立节点：优先图谱计算，降级为无标签联系人 ───
	var isolatedNodes int64
	if reponeo4j.IsAvailable() {
		session := reponeo4j.NewSession(ctx)
		defer session.Close(ctx)
		result, err := session.Run(ctx,
			"MATCH (p:Person) WHERE NOT (p)-[:RELATES_TO]-() RETURN count(p) AS cnt", nil)
		if err == nil && result.Next(ctx) {
			if v, ok := result.Record().Get("cnt"); ok {
				isolatedNodes = toInt64(v)
			}
		}
	}
	if isolatedNodes == 0 {
		repomysql.DB.Model(&model.Contact{}).
			Where("tenant_id = ? AND (tags IS NULL OR tags = '' OR tags = '[]')", tenantIDStr).
			Count(&isolatedNodes)
	}

	// ─── 5. 关系类型分布（基于联系人标签）───
	relationDistribution := h.relationDistribution(tenantIDStr)

	// ─── 6. 人脉增长趋势（按月，最近 8 个月）───
	growthTrend := h.growthTrend(tenantIDStr)

	// ─── 7. 超级连接者 ───
	superConnectors := h.superConnectors(ctx, tenantIDStr)

	dto.Success(c, gin.H{
		"active_relations":      activeRelations,
		"week_new_contacts":     weekNewContacts,
		"avg_path_length":       avgPathLength,
		"isolated_nodes":        isolatedNodes,
		"relation_distribution": relationDistribution,
		"growth_trend":          growthTrend,
		"super_connectors":      superConnectors,
	})
}

func (h *DashboardHandler) countActiveRelations(ctx context.Context, tenantID string) int64 {
	var totalContacts int64
	repomysql.DB.Model(&model.Contact{}).Where("tenant_id = ?", tenantID).Count(&totalContacts)
	if !reponeo4j.IsAvailable() {
		return totalContacts
	}
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)
	result, err := session.Run(ctx,
		"MATCH (:Person)-[r:RELATES_TO]->(:Person) RETURN count(r) AS cnt", nil)
	if err == nil && result.Next(ctx) {
		if v, ok := result.Record().Get("cnt"); ok {
			return toInt64(v)
		}
	}
	return totalContacts
}

func (h *DashboardHandler) relationDistribution(tenantID string) []gin.H {
	var contacts []model.Contact
	if err := repomysql.DB.Select("tags").Where("tenant_id = ?", tenantID).Find(&contacts).Error; err != nil {
		return []gin.H{}
	}

	// 标签 ID -> 名称（历史 UUID 数据兼容）
	idToName := map[string]string{}
	var tags []model.Tag
	repomysql.DB.Where("tenant_id = ?", tenantID).Find(&tags)
	for _, t := range tags {
		idToName[t.ID] = t.Name
	}

	counts := map[string]int64{}
	var total int64
	for _, c := range contacts {
		var names []string
		if err := json.Unmarshal([]byte(c.Tags), &names); err != nil || len(names) == 0 {
			continue
		}
		for _, n := range names {
			if name, ok := idToName[n]; ok {
				n = name
			}
			if n == "" {
				continue
			}
			counts[n]++
			total++
		}
	}

	type kv struct {
		name  string
		count int64
	}
	list := make([]kv, 0, len(counts))
	for name, cnt := range counts {
		list = append(list, kv{name, cnt})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].count > list[j].count })
	if len(list) > 8 {
		list = list[:8]
	}

	result := make([]gin.H, 0, len(list))
	for _, item := range list {
		percent := int64(0)
		if total > 0 {
			percent = int64(math.Round(float64(item.count) * 100 / float64(total)))
		}
		result = append(result, gin.H{
			"type":    item.name,
			"label":   item.name,
			"count":   item.count,
			"percent": percent,
		})
	}
	return result
}

func (h *DashboardHandler) growthTrend(tenantID string) []gin.H {
	type row struct {
		Month string
		Count int64
	}
	var rows []row
	err := repomysql.DB.Raw(
		"SELECT DATE_FORMAT(created_at, '%Y-%m') AS month, COUNT(*) AS count "+
			"FROM contacts WHERE tenant_id = ? GROUP BY month ORDER BY month DESC LIMIT 8",
		tenantID).Scan(&rows).Error
	if err != nil {
		return []gin.H{}
	}

	result := make([]gin.H, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		label := rows[i].Month
		if len(label) >= 3 && label[2] == '-' {
			if n, err := strconv.Atoi(label[3:]); err == nil {
				label = fmt.Sprintf("%d月", n)
			}
		}
		result = append(result, gin.H{"month": label, "count": rows[i].Count})
	}
	return result
}

func (h *DashboardHandler) superConnectors(ctx context.Context, tenantID string) []gin.H {
	if reponeo4j.IsAvailable() {
		session := reponeo4j.NewSession(ctx)
		defer session.Close(ctx)
		result, err := session.Run(ctx,
			`MATCH (p:Person)-[r:RELATES_TO]-()
			 WHERE r.valid_until >= date() OR r.valid_until IS NULL
			 WITH p, COUNT(r) AS degree
			 WHERE degree > 0
			 RETURN p.id AS id, p.name AS name, p.company AS company, degree
			 ORDER BY degree DESC LIMIT 5`, nil)
		if err == nil {
			var list []gin.H
			for result.Next(ctx) {
				rec := result.Record()
				id, _ := rec.Get("id")
				name, _ := rec.Get("name")
				company, _ := rec.Get("company")
				degree, _ := rec.Get("degree")
				companyStr := ""
				if company != nil {
					companyStr = fmt.Sprintf("%v", company)
				}
				list = append(list, gin.H{
					"id":             fmt.Sprintf("%v", id),
					"name":           fmt.Sprintf("%v", name),
					"company":        companyStr,
					"degree":         toInt64(degree),
					"top_connection": "",
				})
			}
			if len(list) > 0 {
				return list
			}
		}
	}

	// 降级：按标签覆盖广度取 Top5（无图谱时基于联系人标签估算）
	var contacts []model.Contact
	repomysql.DB.Where("tenant_id = ?", tenantID).Find(&contacts)

	// 标签 ID -> 名称（历史 UUID 数据兼容）
	idToName := map[string]string{}
	var tags []model.Tag
	repomysql.DB.Where("tenant_id = ?", tenantID).Find(&tags)
	for _, t := range tags {
		idToName[t.ID] = t.Name
	}

	type conn struct {
		id, name, company string
		degree            int
		top               string
	}
	var list []conn
	for _, ct := range contacts {
		var names []string
		if err := json.Unmarshal([]byte(ct.Tags), &names); err != nil {
			continue
		}
		seen := map[string]bool{}
		var uniq []string
		for _, n := range names {
			if name, ok := idToName[n]; ok {
				n = name
			}
			if n != "" && !seen[n] {
				seen[n] = true
				uniq = append(uniq, n)
			}
		}
		if len(uniq) == 0 {
			continue
		}
		top := strings.Join(uniq, "、")
		if len(uniq) > 3 {
			top = strings.Join(uniq[:3], "、") + " 等"
		}
		list = append(list, conn{ct.ID, ct.Name, ct.Company, len(uniq), top})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].degree > list[j].degree })
	if len(list) > 5 {
		list = list[:5]
	}
	result := make([]gin.H, 0, len(list))
	for _, item := range list {
		result = append(result, gin.H{
			"id":             item.id,
			"name":           item.name,
			"company":        item.company,
			"degree":         item.degree,
			"top_connection": item.top,
		})
	}
	return result
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}
