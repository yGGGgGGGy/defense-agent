import { useState, useEffect, useCallback } from 'react';
import { api } from './api/client';
import type { Instance, AuditRecord, AgentInfo } from './types';
import { DagView } from './components/DagView';
import { AuditTimeline } from './components/AuditTimeline';
import { AgentPanel } from './components/AgentPanel';
import { useSSE } from './hooks/useSSE';

const styles: Record<string, React.CSSProperties> = {
  container: {
    fontFamily: 'system-ui, sans-serif',
    maxWidth: 1400,
    margin: '0 auto',
    padding: '20px',
    background: '#f5f5f5',
    minHeight: '100vh',
  },
  header: {
    background: 'linear-gradient(135deg, #1a56db, #1e3a5f)',
    color: 'white',
    padding: '20px 24px',
    borderRadius: '8px',
    marginBottom: '20px',
  },
  title: { margin: 0, fontSize: '24px' },
  subtitle: { margin: '4px 0 0', opacity: 0.8, fontSize: '14px' },
  grid: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '20px',
  },
  fullWidth: { gridColumn: '1 / -1' },
  card: {
    background: 'white',
    borderRadius: '8px',
    padding: '16px',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  cardTitle: {
    margin: '0 0 12px',
    fontSize: '16px',
    fontWeight: 600,
    color: '#1e3a5f',
    borderBottom: '2px solid #e5e7eb',
    paddingBottom: '8px',
  },
  button: {
    padding: '8px 16px',
    background: '#1a56db',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '14px',
  },
  input: {
    padding: '8px 12px',
    border: '1px solid #d1d5db',
    borderRadius: '4px',
    fontSize: '14px',
    width: '100%',
    boxSizing: 'border-box' as const,
    marginBottom: '8px',
  },
  badge: {
    display: 'inline-block',
    padding: '2px 8px',
    borderRadius: '10px',
    fontSize: '12px',
    fontWeight: 600,
  } as React.CSSProperties,
  instanceRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '10px 12px',
    borderBottom: '1px solid #e5e7eb',
  },
  select: {
    padding: '8px 12px',
    border: '1px solid #d1d5db',
    borderRadius: '4px',
    fontSize: '14px',
  },
};

const statusColor: Record<string, string> = {
  done: '#059669',
  running: '#1a56db',
  failed: '#dc2626',
  pending: '#d97706',
};

export default function App() {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [selectedId, setSelectedId] = useState<string>('');
  const [auditRecords, setAuditRecords] = useState<AuditRecord[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [scene, setScene] = useState('incident_response');
  const [taskTitle, setTaskTitle] = useState('');
  const [taskInput, setTaskInput] = useState('');
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    const insts = await api.getInstances();
    setInstances(insts);
  }, []);

  useEffect(() => {
    refresh();
    api.getAgents().then((d) => setAgents(d.agents));
    const timer = setInterval(refresh, 3000);
    return () => clearInterval(timer);
  }, [refresh]);

  const selected = instances.find((i) => i.id === selectedId);

  useEffect(() => {
    if (selectedId) {
      api.getAuditTrail(selectedId).then(setAuditRecords);
    }
  }, [selectedId]);

  // SSE real-time updates
  useSSE(selectedId, useCallback((event) => {
    if (event.type === 'node_state' || event.type === 'instance_update') {
      refresh();
    }
    if (event.type === 'audit_append') {
      api.getAuditTrail(selectedId!).then(setAuditRecords);
    }
  }, [selectedId, refresh]));

  const submitTask = async () => {
    if (!taskTitle || !taskInput) return;
    setLoading(true);
    const inst = await api.createTask({
      scene,
      title: taskTitle,
      description: taskInput,
      input: taskInput,
    });
    setSelectedId(inst.id);
    setLoading(false);
    setTaskTitle('');
    setTaskInput('');
    refresh();
  };

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h1 style={styles.title}>Defense Agent System</h1>
        <p style={styles.subtitle}>自主决策通用防御智能体 | 11 Agents | DAG 编排 | 审计链</p>
      </div>

      {/* Task Submission */}
      <div style={{ ...styles.card, marginBottom: 20 }}>
        <h2 style={styles.cardTitle}>Submit Task</h2>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          <select style={styles.select} value={scene} onChange={(e) => setScene(e.target.value)}>
            <option value="incident_response">应急响应</option>
            <option value="ops_maintenance">运维管理</option>
            <option value="penetration_test">渗透测试</option>
            <option value="vulnerability_research">漏洞挖掘</option>
            <option value="reverse_engineering">逆向分析</option>
          </select>
          <input
            style={{ ...styles.input, flex: 1, minWidth: 200 }}
            placeholder="Task title (e.g. SSH Brute Force Alert)"
            value={taskTitle}
            onChange={(e) => setTaskTitle(e.target.value)}
          />
          <input
            style={{ ...styles.input, flex: 2, minWidth: 300 }}
            placeholder="Description (e.g. Detected 150+ failed SSH logins from 10.0.0.50)"
            value={taskInput}
            onChange={(e) => setTaskInput(e.target.value)}
          />
          <button style={styles.button} onClick={submitTask} disabled={loading}>
            {loading ? 'Submitting...' : 'Submit Task'}
          </button>
        </div>
      </div>

      <div style={styles.grid}>
        {/* Instances */}
        <div style={styles.card}>
          <h2 style={styles.cardTitle}>
            Instances ({instances.length})
            <button style={{ ...styles.button, marginLeft: 12, padding: '4px 10px' }} onClick={refresh}>
              Refresh
            </button>
          </h2>
          <div style={{ maxHeight: 400, overflow: 'auto' }}>
            {instances.length === 0 && <p style={{ color: '#888' }}>No tasks submitted yet.</p>}
            {instances.map((inst) => (
              <div
                key={inst.id}
                style={{
                  ...styles.instanceRow,
                  background: inst.id === selectedId ? '#eff6ff' : 'transparent',
                  cursor: 'pointer',
                }}
                onClick={() => setSelectedId(inst.id)}
              >
                <div>
                  <strong>{inst.task.title}</strong>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    {inst.id} | Scene: {inst.task.scene}
                  </div>
                </div>
                <span style={{ ...styles.badge, background: statusColor[inst.status] || '#666', color: 'white' }}>
                  {inst.status}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Agents */}
        <AgentPanel agents={agents} />

        {/* DAG View */}
        {selected && (
          <div style={styles.card}>
            <h2 style={styles.cardTitle}>DAG: {selected.id}</h2>
            <DagView dag={selected.dag} />
          </div>
        )}

        {/* Audit Trail */}
        {selected && auditRecords.length > 0 && (
          <div style={styles.card}>
            <h2 style={styles.cardTitle}>Audit Chain ({auditRecords.length} records)</h2>
            <AuditTimeline records={auditRecords} />
          </div>
        )}
      </div>
    </div>
  );
}
