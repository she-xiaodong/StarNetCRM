package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/starnet/crm/internal/config"
)

var Driver neo4j.DriverWithContext

// Init 初始化Neo4j连接
func Init(cfg config.Neo4jConfig) error {
	var err error
	Driver, err = neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.User, cfg.Password, ""),
	)
	if err != nil {
		return fmt.Errorf("failed to connect Neo4j: %w", err)
	}

	// 验证连接
	ctx := context.Background()
	if err := Driver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	// 创建约束
	if err := createConstraints(ctx, cfg.Database); err != nil {
		return fmt.Errorf("failed to create constraints: %w", err)
	}

	return nil
}

// Close 关闭连接
func Close() {
	if Driver != nil {
		_ = Driver.Close(context.Background())
	}
}

// IsAvailable 检查Neo4j是否可用
func IsAvailable() bool {
	return Driver != nil
}

// NewSession 创建新会话（调用前需确保IsAvailable()为true）
func NewSession(ctx context.Context) neo4j.SessionWithContext {
	return Driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: config.AppConfig.Neo4j.Database,
	})
}

// createConstraints 创建Neo4j约束
func createConstraints(ctx context.Context, database string) error {
	session := Driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
	})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		"CREATE CONSTRAINT person_id_unique IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE",
		nil,
	)
	return err
}

// CreatePerson 创建Person节点
func CreatePerson(ctx context.Context, session neo4j.SessionWithContext, p map[string]interface{}) error {
	_, err := session.Run(ctx,
		`CREATE (p:Person {
			id: $id, name: $name, company: $company, title: $title,
			phone: $phone, email: $email, wecom_id: $wecom_id,
			avatar: $avatar, department: $department,
			is_active: $is_active, create_time: datetime(), update_time: datetime()
		})`,
		p,
	)
	return err
}

// CreateRelation 创建RELATES_TO关系（使用MERGE确保节点存在）
func CreateRelation(ctx context.Context, session neo4j.SessionWithContext, props map[string]interface{}) error {
	_, err := session.Run(ctx,
		`MERGE (a:Person {id: $from_id})
		 MERGE (b:Person {id: $to_id})
		 MERGE (a)-[r:RELATES_TO {
		 	type: $type, source: $source, strength: $strength,
		 	valid_until: $valid_until, is_shared: $is_shared,
		 	created_by: $created_by, create_time: datetime(), update_time: datetime()
		 }]->(b)
		 RETURN r`,
		props,
	)
	return err
}

// UpdateRelation 更新关系属性（type/strength）
func UpdateRelation(ctx context.Context, session neo4j.SessionWithContext, props map[string]interface{}) error {
	_, err := session.Run(ctx,
		`MATCH (a:Person {id: $from_id})-[r:RELATES_TO {relation_id: $relation_id}]->(b:Person {id: $to_id})
		 SET r.type = $type, r.strength = $strength, r.update_time = datetime()
		 RETURN r`,
		props,
	)
	return err
}

// DeleteRelation 删除两个 Person 之间的 RELATES_TO 关系
func DeleteRelation(ctx context.Context, session neo4j.SessionWithContext, fromID, toID string) error {
	_, err := session.Run(ctx,
		`MATCH (a:Person {id: $from_id})-[r:RELATES_TO]->(b:Person {id: $to_id})
		 DELETE r`,
		map[string]interface{}{
			"from_id": fromID,
			"to_id":   toID,
		},
	)
	return err
}

// EnsurePerson 确保Person节点存在（不存在则创建，存在则更新name）
func EnsurePerson(ctx context.Context, session neo4j.SessionWithContext, id, name string) error {
	_, err := session.Run(ctx,
		`MERGE (p:Person {id: $id})
		 ON CREATE SET p.name = $name, p.create_time = datetime(), p.update_time = datetime()
		 ON MATCH SET p.name = $name, p.update_time = datetime()`,
		map[string]interface{}{
			"id":   id,
			"name": name,
		},
	)
	return err
}

// FindShortestPath 六度人脉最短路径查询
func FindShortestPath(ctx context.Context, session neo4j.SessionWithContext, startID, endID string, maxDepth int) ([]map[string]interface{}, error) {
	if maxDepth <= 0 {
		maxDepth = 6
	}

	query := fmt.Sprintf(`
		MATCH path = shortestPath((p1:Person {id: $start_id})-[*..%d]-(p2:Person {id: $end_id}))
		WHERE ALL(r IN RELATIONSHIPS(path) WHERE r.valid_until >= date() OR r.valid_until IS NULL)
		RETURN path
	`, maxDepth)

	result, err := session.Run(ctx, query, map[string]interface{}{
		"start_id": startID,
		"end_id":   endID,
	})
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		records = append(records, result.Record().AsMap())
	}

	return records, result.Err()
}

// GetFirstDegreeRelations 获取1度关系（直接联系人）
func GetFirstDegreeRelations(ctx context.Context, session neo4j.SessionWithContext, personID string) ([]map[string]interface{}, error) {
	result, err := session.Run(ctx,
		`MATCH (p:Person {id: $person_id})-[r:RELATES_TO]-(friend:Person)
		 WHERE r.valid_until >= date() OR r.valid_until IS NULL
		 RETURN friend, r, type(r) as rel_type`,
		map[string]interface{}{"person_id": personID},
	)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		records = append(records, result.Record().AsMap())
	}

	return records, result.Err()
}

// GetSuperConnectors 识别超级连接者
func GetSuperConnectors(ctx context.Context, session neo4j.SessionWithContext, threshold int, limit int) ([]map[string]interface{}, error) {
	result, err := session.Run(ctx,
		`MATCH (p:Person)-[r:RELATES_TO]-()
		 WHERE r.valid_until >= date() OR r.valid_until IS NULL
		 WITH p, COUNT(r) AS degree
		 WHERE degree > $threshold
		 RETURN p.name AS name, p.id AS id, degree
		 ORDER BY degree DESC
		 LIMIT $limit`,
		map[string]interface{}{
			"threshold": threshold,
			"limit":     limit,
		},
	)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		records = append(records, result.Record().AsMap())
	}

	return records, result.Err()
}
