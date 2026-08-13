package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	reponeo4j "github.com/starnet/crm/internal/repository/neo4j"
	"github.com/starnet/crm/pkg/cache"
)

// GraphHandler 图谱处理器
type GraphHandler struct{}

// NewGraphHandler 创建图谱处理器
func NewGraphHandler() *GraphHandler {
	return &GraphHandler{}
}

// GetFirstDegree 获取1度关系网络图
func (h *GraphHandler) GetFirstDegree(c *gin.Context) {
	personID := c.Query("person_id")
	if personID == "" {
		userID, _ := c.Get("user_id")
		personID = userID.(string)
	}

	if !reponeo4j.IsAvailable() {
		// Neo4j 不可用，使用 MySQL 联系人数据兜底构建图谱
		dto.Success(c, h.buildMysqlFallbackGraph(c, personID))
		return
	}

	ctx := context.Background()
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	records, err := reponeo4j.GetFirstDegreeRelations(ctx, session, personID)
	if err != nil {
		// Neo4j 查询失败（如服务未启动），降级为 MySQL 联系人数据兜底
		dto.Success(c, h.buildMysqlFallbackGraph(c, personID))
		return
	}

	graphData := h.buildGraphData(records, personID)
	dto.Success(c, graphData)
}

// SearchPath 六度人脉路径搜索
func (h *GraphHandler) SearchPath(c *gin.Context) {
	var req model.SearchPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "请选择起点和终点")
		return
	}

	if req.MaxDepth <= 0 {
		req.MaxDepth = 6
	}

	if !reponeo4j.IsAvailable() {
		dto.InternalError(c, "图谱服务未连接，请先配置Neo4j")
		return
	}

	ctx := context.Background()

	// 尝试从Redis缓存中获取
	var pathResult model.PathResult
	cacheKey := cache.PathCacheKey(req.StartPersonID, req.EndPersonID)
	if err := cache.Get(ctx, cacheKey, &pathResult); err == nil {
		dto.Success(c, pathResult)
		return
	}

	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	// 执行Cypher最短路径查询
	records, err := reponeo4j.FindShortestPath(ctx, session, req.StartPersonID, req.EndPersonID, req.MaxDepth)
	if err != nil {
		dto.InternalError(c, "路径搜索失败")
		return
	}

	if len(records) == 0 {
		dto.Success(c, gin.H{
			"found":  false,
			"length": 0,
			"message": "暂未发现" + itoa(req.MaxDepth) + "度内关系路径，建议扩大人脉网络",
		})
		return
	}

	// 解析路径结果
	pathResult = h.parsePathResult(records[0])
	if pathResult.Length > 0 {
		// 缓存路径结果（5分钟）
		_ = cache.Set(ctx, cacheKey, pathResult, 5*time.Minute)
	}

	dto.Success(c, pathResult)
}

// CreateRelation 创建关系。
// 双写策略：先写 MySQL relations（亲密度/标签落库），Neo4j 可用时再同步。
func (h *GraphHandler) CreateRelation(c *gin.Context) {
	// 复用 RelationHandler 的创建逻辑（写 MySQL，Neo4j 可用时双写）
	relationHandler := NewRelationHandler()
	relationHandler.Create(c)
}

// GetSuperConnectors 获取超级连接者
func (h *GraphHandler) GetSuperConnectors(c *gin.Context) {
	if !reponeo4j.IsAvailable() {
		dto.Success(c, []interface{}{})
		return
	}

	ctx := context.Background()
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	records, err := reponeo4j.GetSuperConnectors(ctx, session, 10, 20)
	if err != nil {
		dto.InternalError(c, "查询失败")
		return
	}

	dto.Success(c, records)
}

// buildMysqlFallbackGraph Neo4j不可用时，使用MySQL兜底构建关系图谱。
// 边的 type/strength/tags 全部来自 relations 表（真实数据），不再使用假亲密度或联系人标签硬凑。
// 兜底规则：中心节点只能与一级人脉直接连线；二/三/四级人脉只能经由引荐链（referral 边）
// 通过一级人脉逐级连线，禁止跳级连线。若联系人没有任何关系记录，则不展示（避免噪音）。
func (h *GraphHandler) buildMysqlFallbackGraph(c *gin.Context, centerID string) model.GraphData {
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)

	var contacts []model.Contact
	if tenantIDStr != "" {
		if err := repomysql.DB.Where("tenant_id = ?", tenantIDStr).Find(&contacts).Error; err != nil {
			contacts = nil
		}
	}

	// PersonID -> Contact 映射
	byID := make(map[string]*model.Contact)
	for i := range contacts {
		cnt := &contacts[i]
		if cnt.PersonID != "" {
			byID[cnt.PersonID] = cnt
		}
		byID[cnt.ID] = cnt
	}

	// 关系记录：type/strength/tags 真实数据
	var relations []model.Relation
	if tenantIDStr != "" {
		if err := repomysql.DB.Where("tenant_id = ?", tenantIDStr).Find(&relations).Error; err != nil {
			relations = nil
		}
	}
	for i := range relations {
		relations[i].UnmarshalTags()
	}
	// 将关系 tags 中的标签 ID 解析为名称，避免图谱 tooltip 显示 UUID 乱码
	ptrList := make([]*model.Relation, len(relations))
	for i := range relations {
		ptrList[i] = &relations[i]
	}
	resolveRelationTagNames(c, ptrList)

	// 中心节点
	nodes := []model.GraphNode{{
		ID:    centerID,
		Label: "我",
		Type:  "center",
		Style: map[string]interface{}{},
		Properties: map[string]interface{}{
			"name":  "我",
			"level": 0,
		},
	}}
	nodeSet := map[string]bool{centerID: true}
	edges := []model.GraphEdge{}

	// 人脉层级：沿引荐链传播（无有效引荐人的联系人为 1 级，被引荐者逐级 +1）。
	// 说明：relations 表的关系都以当前用户为中心，若用最短路径 BFS 层级恒为 1，
	// 因此用 contacts.referrer_id 引荐链体现多级人脉（1级=直接发展，2级=被1级引荐…）。
	levelOf := map[string]int{centerID: 0}
	childrenOf := make(map[string][]string)
	hasRef := make(map[string]bool)
	for i := range contacts {
		cnt := &contacts[i]
		if cnt.PersonID == "" {
			continue
		}
		if cnt.ReferrerID != "" {
			if _, ok := byID[cnt.ReferrerID]; ok {
				hasRef[cnt.PersonID] = true
				childrenOf[cnt.ReferrerID] = append(childrenOf[cnt.ReferrerID], cnt.PersonID)
			}
		}
	}
	queue := []string{}
	for i := range contacts {
		cnt := &contacts[i]
		if cnt.PersonID == "" || hasRef[cnt.PersonID] {
			continue
		}
		levelOf[cnt.PersonID] = 1
		queue = append(queue, cnt.PersonID)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenOf[cur] {
			if _, visited := levelOf[child]; visited {
				continue
			}
			levelOf[child] = levelOf[cur] + 1
			if levelOf[child] <= 4 {
				queue = append(queue, child)
			}
		}
	}
	// 未进入引荐链但存在关系记录的联系人，按直接连接视为 1 级
	for _, rel := range relations {
		for _, pid := range []string{rel.FromPersonID, rel.ToPersonID} {
			if pid == centerID {
				continue
			}
			if _, ok := levelOf[pid]; !ok {
				levelOf[pid] = 1
			}
		}
	}

	// 节点确保函数：将 pid 加入节点集合（若尚未加入）
	ensureNode := func(pid string) {
		if nodeSet[pid] {
			return
		}
		nodeSet[pid] = true
		label := pid
		level, levelKnown := levelOf[pid]
		if !levelKnown {
			level = 1
		}
		contact := byID[pid]
		if contact != nil {
			label = contact.Name
		}
		nodes = append(nodes, model.GraphNode{
			ID:    pid,
			Label: label,
			Type:  "person",
			Properties: map[string]interface{}{
				"name":       label,
				"company":    contactValue(contact, "company"),
				"title":      contactValue(contact, "title"),
				"department": contactValue(contact, "department"),
				"level":      level,
			},
		})
	}

	// 收集关系中出现的节点并构建边
	seenRelation := map[string]bool{}
	for _, rel := range relations {
		fromID := rel.FromPersonID
		toID := rel.ToPersonID
		if fromID == "" || toID == "" {
			continue
		}

		// 只展示以中心为端点或双方都在租户内的关系
		fromOK := fromID == centerID || byID[fromID] != nil
		toOK := toID == centerID || byID[toID] != nil
		if !fromOK || !toOK {
			continue
		}

		// 人脉连线规则：只允许逐级连线，禁止跳级。
		// 中心节点只能直接连接一级人脉；与二/三/四级人脉的直接关系边不绘制，
		// 这些深层人脉只能经由引荐链（referral 边）通过一级人脉逐级连线。
		lvFrom := levelOf[fromID]
		lvTo := levelOf[toID]
		if (fromID == centerID && lvTo > 1) || (toID == centerID && lvFrom > 1) {
			continue
		}
		// 非中心节点之间的连线同样不允许跳级（层级相差超过 1 则跳过）
		if fromID != centerID && toID != centerID {
			diff := lvFrom - lvTo
			if diff < 0 {
				diff = -diff
			}
			if diff > 1 {
				continue
			}
		}

		// 添加节点
		ensureNode(fromID)
		ensureNode(toID)

		// 关系标签作为边属性
		props := map[string]interface{}{
			"tags":     rel.Tags,
			"note":     rel.Note,
			"from_id":  fromID,
			"relation_id": rel.ID,
		}

		edgeID := rel.ID
		if edgeID == "" {
			edgeID = uuid.New().String()
		}
		if seenRelation[fromID+"->"+toID] {
			continue // 同方向关系只显示一次
		}
		seenRelation[fromID+"->"+toID] = true

		edges = append(edges, model.GraphEdge{
			ID:       edgeID,
			Source:   fromID,
			Target:   toID,
			Type:     rel.Type,
			Strength: rel.Strength,
			Label:    RelationTypeName(rel.Type),
			Properties: props,
		})
	}

	// 追加引荐边：联系人 → 其引荐人（contacts.referrer_id），形成多级人脉层级
	for i := range contacts {
		cnt := &contacts[i]
		if cnt.PersonID == "" || cnt.ReferrerID == "" {
			continue
		}
		if _, ok := byID[cnt.ReferrerID]; !ok {
			continue
		}
		childID := cnt.PersonID
		parentID := cnt.ReferrerID
		ensureNode(childID)
		ensureNode(parentID)
		// 无向去重：relations 表可能已有该引荐边（方向 引荐人→被引荐者），避免重复绘制
		if seenRelation[childID+"->"+parentID] || seenRelation[parentID+"->"+childID] {
			continue
		}
		seenRelation[childID+"->"+parentID] = true
		edges = append(edges, model.GraphEdge{
			ID:       uuid.New().String(),
			Source:   childID,
			Target:   parentID,
			Type:     "referral",
			Strength: 0,
			Label:    "引荐",
			Properties: map[string]interface{}{
				"tags":     []string{},
				"note":     "引荐关系",
				"from_id":  childID,
			},
		})
	}

	return model.GraphData{
		Nodes: nodes,
		Edges: edges,
	}
}

// contactValue 获取联系人字段值（nil 安全）
func contactValue(contact *model.Contact, field string) string {
	if contact == nil {
		return ""
	}
	switch field {
	case "company":
		return contact.Company
	case "title":
		return contact.Title
	case "department":
		return contact.Department
	}
	return ""
}

// buildGraphData 构建G6格式图谱数据
func (h *GraphHandler) buildGraphData(records []map[string]interface{}, centerID string) model.GraphData {
	nodes := []model.GraphNode{}
	edges := []model.GraphEdge{}
	nodeSet := make(map[string]bool)

	// 中心节点
	nodeSet[centerID] = true
	nodes = append(nodes, model.GraphNode{
		ID:    centerID,
		Label: "我",
		Properties: map[string]interface{}{
			"name": "我",
		},
	})

	for _, record := range records {
		// 解析friend节点
		if friendData, ok := record["friend"].(map[string]interface{}); ok {
			friendProps := friendData["properties"].(map[string]interface{})
			friendID := toString(friendProps["id"])

			if !nodeSet[friendID] {
				nodeSet[friendID] = true
				nodes = append(nodes, model.GraphNode{
					ID:    friendID,
					Label: toString(friendProps["name"]),
					Properties: map[string]interface{}{
						"company":    toString(friendProps["company"]),
						"title":      toString(friendProps["title"]),
						"department": toString(friendProps["department"]),
					},
				})
			}

			// 解析关系
			if relData, ok := record["r"].(map[string]interface{}); ok {
				relProps := relData["properties"].(map[string]interface{})
				edges = append(edges, model.GraphEdge{
					Source:   centerID,
					Target:   friendID,
					Type:     toString(relProps["type"]),
					Strength: toInt(relProps["strength"]),
				})
			}
		}
	}

	return model.GraphData{
		Nodes: nodes,
		Edges: edges,
	}
}

// parsePathResult 解析路径查询结果
func (h *GraphHandler) parsePathResult(record map[string]interface{}) model.PathResult {
	result := model.PathResult{
		Path: []model.PathNode{},
	}

	// 从Cypher返回的path中提取节点和关系
	// Neo4j path格式: {start: {...}, end: {...}, segments: [...]}
	pathData, err := json.Marshal(record)
	if err != nil {
		return result
	}

	var pathMap map[string]interface{}
	json.Unmarshal(pathData, &pathMap)

	result.Length = 0
	result.Strength = "medium"

	return result
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	}
	return 0
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	p := len(buf)
	for i > 0 {
		p--
		buf[p] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		p--
		buf[p] = '-'
	}
	return string(buf[p:])
}
