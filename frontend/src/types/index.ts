export interface AgentInfo {
  type: string;
  capabilities: string[];
}

export interface Alert {
  id: string;
  rule: string;
  source_ip?: string;
  detail: string;
  count?: number;
}

export interface Task {
  id: string;
  scene: string;
  title: string;
  description: string;
  input: string;
  alerts?: Alert[];
}

export interface DagNode {
  id: string;
  agent_type: string;
  state: string;
  dependencies: string[] | null;
  input: string;
  output: string;
  error?: string;
  retry_count: number;
  max_retries: number;
  timeout: number;
}

export interface Dag {
  id: string;
  task_id: string;
  scene: string;
  nodes: Record<string, DagNode>;
  root_nodes: string[];
}

export interface Instance {
  id: string;
  task: Task;
  dag: Dag;
  status: string;
  created_at: string;
  updated_at: string;
  error?: string;
}

export interface AuditRecord {
  id: string;
  task_id: string;
  agent_type: string;
  decision: string;
  rationale: string;
  evidence: Array<{ type: string; source: string; detail: string }>;
  confidence: number;
  risk_level: string;
  status: string;
  timestamp: string;
  prev_hash: string;
  record_hash: string;
  human_approval_required: boolean;
}

export interface GateResult {
  approved: boolean;
  auto_approved: boolean;
  requires_human: boolean;
  message?: string;
}
