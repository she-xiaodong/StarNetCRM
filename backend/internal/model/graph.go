package model

import "time"

// PersonNode Neo4j中Person节点
type PersonNode struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Company    string    `json:"company"`
	Title      string    `json:"title"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	WecomID    string    `json:"wecom_id"`
	Avatar     string    `json:"avatar"`
	Department string    `json:"department"`
	IsActive   bool      `json:"is_active"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// RelationEdge 关系边（Neo4j RELATES_TO）
type RelationEdge struct {
	ID         string    `json:"id"` // Neo4j relationship elementId
	FromID     string    `json:"from_id"`
	ToID       string    `json:"to_id"`
	Type       string    `json:"type"`        // colleague/manager/customer/partner/alumni/friend/custom
	Source     string    `json:"source"`      // 建立来源
	Strength   int       `json:"strength"`    // 亲密度 1-10
	ValidUntil string    `json:"valid_until"` // 有效期日期
	IsShared   string    `json:"is_shared"`   // private/department/company
	CreatedBy  string    `json:"created_by"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// RelationshipType 预设关系类型
var RelationshipTypes = []string{
	"colleague",  // 同事
	"manager",    // 上下级
	"customer",   // 客户
	"partner",    // 合作伙伴
	"alumni",     // 校友
	"friend",     // 朋友
	"custom",     // 自定义
}

// VisibilityLevel 可见性级别
const (
	VisibilityPrivate    = "private"
	VisibilityDepartment = "department"
	VisibilityCompany    = "company"
)

// GraphNode 图谱可视化节点（G6格式）
type GraphNode struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Type       string                 `json:"type"`   // 节点样式类型
	Size       int                    `json:"size"`   // 根据人脉评分计算的尺寸
	Style      map[string]interface{} `json:"style,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GraphEdge 图谱可视化边（G6格式）
type GraphEdge struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Label      string                 `json:"label,omitempty"`
	Type       string                 `json:"type"` // 关系类型
	Strength   int                    `json:"strength"`
	Style      map[string]interface{} `json:"style,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GraphData 图谱数据（G6格式）
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// PathResult 路径查询结果
type PathResult struct {
	Path       []PathNode  `json:"path"`
	Length     int         `json:"length"`
	Strength   string      `json:"strength"` // strong/medium/weak
	GraphData  *GraphData  `json:"graph_data,omitempty"` // 用于G6高亮渲染
}

// PathNode 路径中的节点
type PathNode struct {
	Person     PersonNode    `json:"person"`
	Relation   *RelationEdge `json:"relation,omitempty"` // 与上一节点的关系
	Degree     int           `json:"degree"`              // 距起点的度数
}

// SearchPathRequest 路径搜索请求
type SearchPathRequest struct {
	StartPersonID string `json:"start_person_id" binding:"required"`
	EndPersonID   string `json:"end_person_id" binding:"required"`
	MaxDepth      int    `json:"max_depth"` // 默认6
}


