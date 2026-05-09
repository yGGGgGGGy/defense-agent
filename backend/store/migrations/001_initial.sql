-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Agent registry
CREATE TABLE IF NOT EXISTS agents (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_type  TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'idle',
    capabilities TEXT[],
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Scene templates
CREATE TABLE IF NOT EXISTS scenes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    agent_order TEXT[],
    template    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Task instances
CREATE TABLE IF NOT EXISTS task_instances (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id TEXT NOT NULL UNIQUE,
    scene       TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    input_json  JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit records (tamper-evident chain)
CREATE TABLE IF NOT EXISTS audit_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id         TEXT NOT NULL,
    sub_task_id     TEXT,
    agent_type      TEXT NOT NULL,
    agent_id        TEXT,
    decision        TEXT NOT NULL,
    rationale       TEXT NOT NULL,
    evidence_json   JSONB,
    confidence      REAL,
    risk_level      TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    prev_hash       TEXT,
    record_hash     TEXT NOT NULL,
    requires_human  BOOLEAN DEFAULT FALSE,
    approved_by     TEXT,
    approved_at     TIMESTAMPTZ
);

-- DAG nodes
CREATE TABLE IF NOT EXISTS dag_nodes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id         TEXT NOT NULL,
    agent_type      TEXT NOT NULL,
    state           TEXT NOT NULL DEFAULT 'pending',
    input_json      JSONB,
    output_json     JSONB,
    dependencies    TEXT[],
    retry_count     INTEGER DEFAULT 0,
    max_retries     INTEGER DEFAULT 3,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Memory entries with vector embeddings
CREATE TABLE IF NOT EXISTS memory_entries (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id     TEXT NOT NULL,
    agent_type  TEXT NOT NULL,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    embedding   vector(1536),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_task ON audit_records(task_id);
CREATE INDEX IF NOT EXISTS idx_audit_agent ON audit_records(agent_type);
CREATE INDEX IF NOT EXISTS idx_dag_task ON dag_nodes(task_id);
CREATE INDEX IF NOT EXISTS idx_memory_task ON memory_entries(task_id);
CREATE INDEX IF NOT EXISTS idx_memory_agent ON memory_entries(agent_type);
-- IVFFlat index for vector similarity search
CREATE INDEX IF NOT EXISTS idx_memory_embedding ON memory_entries USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Insert default scene templates
INSERT INTO scenes (name, description, agent_order, template) VALUES
('incident_response', 'Emergency response to security incidents',
 ARRAY['perceiver', 'analyst', 'responder', 'operator'],
 '{"version":1,"edges":[{"from":"perceiver","to":"analyst"},{"from":"analyst","to":"responder"},{"from":"analyst","to":"operator"}]}'),
('ops_maintenance', 'Routine system health checks and maintenance',
 ARRAY['operator'],
 '{"version":1,"edges":[]}')
ON CONFLICT (name) DO NOTHING;

-- Insert default agents
INSERT INTO agents (agent_type, name, capabilities) VALUES
('perceiver', 'Perceiver Agent', ARRAY['log_collection', 'traffic_analysis', 'asset_discovery', 'alert_aggregation']),
('analyst', 'Analyst Agent', ARRAY['alert_correlation', 'attack_chain_mapping', 'root_cause_analysis', 'impact_assessment', 'attck_mapping']),
('responder', 'Responder Agent', ARRAY['ip_blocking', 'host_isolation', 'patch_deployment', 'service_recovery']),
('operator', 'Operator Agent', ARRAY['health_check', 'config_compliance', 'patch_management', 'backup_restore', 'capacity_planning', 'service_orchestration', 'log_rotation', 'certificate_management'])
ON CONFLICT (agent_type) DO NOTHING;
