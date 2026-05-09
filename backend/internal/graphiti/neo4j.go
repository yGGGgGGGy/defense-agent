package graphiti

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog/log"
)

// Client wraps Neo4j for knowledge graph operations
type Client struct {
	driver   neo4j.DriverWithContext
	database string
}

// NewClient creates a Neo4j client
func NewClient(uri, user, password, database string) (*Client, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		log.Warn().Err(err).Msg("neo4j unreachable, graph features disabled")
		return &Client{driver: nil, database: database}, nil
	}

	log.Info().Str("uri", uri).Msg("connected to neo4j")
	return &Client{driver: driver, database: database}, nil
}

// RecordAgentAction stores an agent's action as a graph node
func (c *Client) RecordAgentAction(ctx context.Context, taskID, agentType, action, result string) error {
	if c.driver == nil {
		return nil // gracefully skip
	}
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: c.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (t:Task {id: $taskID})
			CREATE (a:Action {
				id: randomUUID(),
				agent_type: $agentType,
				action: $action,
				result: $result,
				timestamp: datetime()
			})
			MERGE (t)-[:HAS_ACTION]->(a)
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"taskID":    taskID,
			"agentType": agentType,
			"action":    action,
			"result":    result,
		})
		return nil, err
	})
	return err
}

// RecordAttackPath stores an attack path in the graph
func (c *Client) RecordAttackPath(ctx context.Context, taskID string, nodes []AttackNode, edges []AttackEdge) error {
	if c.driver == nil {
		return nil
	}
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: c.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, n := range nodes {
			query := fmt.Sprintf(`MERGE (n:%s {id: $id}) SET n += $props`, n.Label)
			tx.Run(ctx, query, map[string]any{"id": n.ID, "props": n.Props})
		}
		for _, e := range edges {
			query := fmt.Sprintf(`
				MATCH (a {id: $from}), (b {id: $to})
				MERGE (a)-[:%s]->(b)
			`, e.RelType)
			tx.Run(ctx, query, map[string]any{"from": e.From, "to": e.To})
		}
		return nil, nil
	})
	return err
}

// RecordDecisionChain stores the audit decision chain as a graph
func (c *Client) RecordDecisionChain(ctx context.Context, taskID string, decisions []DecisionNode) error {
	if c.driver == nil {
		return nil
	}
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: c.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for i, d := range decisions {
			tx.Run(ctx, `
				MERGE (t:Task {id: $taskID})
				CREATE (d:Decision {
					id: $id,
					agent: $agent,
					decision: $decision,
					risk: $risk,
					position: $pos
				})
				MERGE (t)-[:HAS_DECISION]->(d)
			`, map[string]any{
				"taskID":   taskID,
				"id":       d.ID,
				"agent":    d.Agent,
				"decision": d.Decision,
				"risk":     d.Risk,
				"pos":      i,
			})
		}
		return nil, nil
	})
	return err
}

// QueryAttackPath queries the graph for attack paths of a task
func (c *Client) QueryAttackPath(ctx context.Context, taskID string) ([]map[string]any, error) {
	if c.driver == nil {
		return nil, nil
	}
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: c.database})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (t:Task {id: $taskID})-[:HAS_ACTION]->(a:Action)
			OPTIONAL MATCH (a)-[r]->(b)
			RETURN a, r, b
			ORDER BY a.timestamp
		`
		records, err := tx.Run(ctx, cypher, map[string]any{"taskID": taskID})
		if err != nil {
			return nil, err
		}
		var results []map[string]any
		for records.Next(ctx) {
			record := records.Record()
			results = append(results, record.AsMap())
		}
		return results, nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.([]map[string]any), nil
}

// AssetGraph stores asset and relationship data
func (c *Client) RecordAssetRelation(ctx context.Context, source, target, relType string, props map[string]any) error {
	if c.driver == nil {
		return nil
	}
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: c.database})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := fmt.Sprintf(`
			MERGE (a:Asset {id: $source})
			MERGE (b:Asset {id: $target})
			MERGE (a)-[:%s {props: $props}]->(b)
		`, relType)
		_, err := tx.Run(ctx, query, map[string]any{
			"source": source,
			"target": target,
			"props":  props,
		})
		return nil, err
	})
	return err
}

// Close shuts down the Neo4j driver
func (c *Client) Close() {
	if c.driver != nil {
		c.driver.Close(context.Background())
	}
}

// Types for graph operations
type AttackNode struct {
	ID    string
	Label string
	Props map[string]any
}

type AttackEdge struct {
	From    string
	To      string
	RelType string
}

type DecisionNode struct {
	ID       string
	Agent    string
	Decision string
	Risk     string
}
