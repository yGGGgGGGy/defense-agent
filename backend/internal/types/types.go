package types

import "time"

// AgentType identifies the agent role
type AgentType string

const (
	AgentPerceiver  AgentType = "perceiver"
	AgentAnalyst    AgentType = "analyst"
	AgentResponder  AgentType = "responder"
	AgentOperator   AgentType = "operator"
	AgentResearcher AgentType = "researcher"
	AgentDeveloper  AgentType = "developer"
	AgentExecutor   AgentType = "executor"
	AgentAdviser    AgentType = "adviser"
	AgentReflector  AgentType = "reflector"
	AgentAuditor    AgentType = "auditor"
	AgentMemorist   AgentType = "memorist"
)

// RiskLevel defines the risk classification of an action
type RiskLevel string

const (
	RiskLow       RiskLevel = "low"
	RiskMedium    RiskLevel = "medium"
	RiskHigh      RiskLevel = "high"
	RiskDangerous RiskLevel = "dangerous"
)

// ActionStatus tracks execution status
type ActionStatus string

const (
	ActionPending   ActionStatus = "pending"
	ActionRunning   ActionStatus = "running"
	ActionSucceeded ActionStatus = "succeeded"
	ActionFailed    ActionStatus = "failed"
	ActionBlocked   ActionStatus = "blocked"
)

// SceneType identifies the operational scenario
type SceneType string

const (
	SceneIncidentResponse SceneType = "incident_response"
	ScenePenTest          SceneType = "penetration_test"
	SceneVulnResearch     SceneType = "vulnerability_research"
	SceneReverseEng       SceneType = "reverse_engineering"
	SceneOpsMaintenance   SceneType = "ops_maintenance"
)

// Evidence is a piece of supporting data for a decision
type Evidence struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Detail string `json:"detail"`
}

// Action represents a single unit of work
type Action struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Command   string       `json:"command"`
	Args      []string     `json:"args,omitempty"`
	RiskLevel RiskLevel    `json:"risk_level"`
	Rationale string       `json:"rationale"`
	Evidence  []Evidence   `json:"evidence,omitempty"`
	Sandbox   bool         `json:"sandbox"`
	Timeout   int          `json:"timeout"` // seconds
	Status    ActionStatus `json:"status"`
	Output    string       `json:"output,omitempty"`
}

// Task is the top-level work item
type Task struct {
	ID          string    `json:"id"`
	Scene       SceneType `json:"scene"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Input       string    `json:"input"`
	Alerts      []Alert   `json:"alerts,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Alert is a security alert input
type Alert struct {
	ID       string `json:"id"`
	Rule     string `json:"rule"`
	SourceIP string `json:"source_ip,omitempty"`
	Detail   string `json:"detail"`
	Count    int    `json:"count,omitempty"`
}

// InstanceStatus tracks overall task instance state
type InstanceStatus string

const (
	InstancePending InstanceStatus = "pending"
	InstanceRunning InstanceStatus = "running"
	InstanceDone    InstanceStatus = "done"
	InstanceFailed  InstanceStatus = "failed"
)

// NodeState tracks DAG node state
type NodeState string

const (
	NodePending   NodeState = "pending"
	NodeRunning   NodeState = "running"
	NodeSucceeded NodeState = "succeeded"
	NodeFailed    NodeState = "failed"
	NodeSkipped   NodeState = "skipped"
)

// SubTask maps to a DAG node
type SubTask struct {
	ID           string    `json:"id"`
	TaskID       string    `json:"task_id"`
	AgentType    AgentType `json:"agent_type"`
	Instructions string    `json:"instructions"`
	Context      string    `json:"context"`
	Status       string    `json:"status"`
	Result       string    `json:"result,omitempty"`
}
