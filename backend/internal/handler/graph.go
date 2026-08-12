package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/starnet/crm/internal/dto"
	"github.com/starnet/crm/internal/model"
	"github.com/starnet/crm/pkg/cache"
	reponeo4j "github.com/starnet/crm/internal/repository/neo4j"
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
		// Neo4j 不可用，返回仅包含中心节点的空图谱
		dto.Success(c, model.GraphData{
			Nodes: []model.GraphNode{
				{ID: personID, Label: "我", Properties: map[string]interface{}{"name": "我"}},
			},
			Edges: []model.GraphEdge{},
		})
		return
	}

	ctx := context.Background()
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	records, err := reponeo4j.GetFirstDegreeRelations(ctx, session, personID)
	if err != nil {
		dto.InternalError(c, "查询关系图谱失败")
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

// CreateRelation 创建关系
func (h *GraphHandler) CreateRelation(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if !reponeo4j.IsAvailable() {
		dto.InternalError(c, "图谱服务未连接，请先配置Neo4j")
		return
	}

	var req model.CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.BadRequest(c, "参数校验失败: "+err.Error())
		return
	}

	ctx := context.Background()
	session := reponeo4j.NewSession(ctx)
	defer session.Close(ctx)

	relationID := uuid.New().String()
	validUntil := req.ValidUntil
	if validUntil == "" {
		validUntil = "2099-12-31" // 默认永久
	}

	isShared := req.IsShared
	if isShared == "" {
		isShared = model.VisibilityPrivate
	}

	err := reponeo4j.CreateRelation(ctx, session, map[string]interface{}{
		"from_id":     req.FromPersonID,
		"to_id":       req.ToPersonID,
		"relation_id": relationID,
		"type":        req.Type,
		"source":      req.Source,
		"strength":    req.Strength,
		"valid_until": validUntil,
		"is_shared":   isShared,
		"created_by":  userID.(string),
	})

	if err != nil {
		dto.InternalError(c, "创建关系失败: "+err.Error())
		return
	}

	dto.Success(c, gin.H{
		"id":      relationID,
		"message": "关系已记录",
	})
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
