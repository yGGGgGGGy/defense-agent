import type { AgentInfo } from '../types';

const agentColors: Record<string, string> = {
  perceiver: '#3b82f6',
  analyst: '#8b5cf6',
  responder: '#ef4444',
  operator: '#10b981',
  researcher: '#f59e0b',
  developer: '#ec4899',
  executor: '#6366f1',
  adviser: '#14b8a6',
  reflector: '#f97316',
  auditor: '#06b6d4',
  memorist: '#84cc16',
};

interface Props {
  agents: AgentInfo[];
}

export function AgentPanel({ agents }: Props) {
  return (
    <div>
      <h2 style={{ margin: '0 0 12px', fontSize: '16px', fontWeight: 600, color: '#1e3a5f' }}>
        Agents ({agents.length})
      </h2>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '8px' }}>
        {agents.map((a) => (
          <div
            key={a.type}
            style={{
              padding: '10px 12px',
              borderRadius: '6px',
              borderLeft: `3px solid ${agentColors[a.type] || '#888'}`,
              background: 'white',
              boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
            }}
          >
            <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '4px' }}>
              {a.type}
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
              {a.capabilities.slice(0, 3).map((cap) => (
                <span
                  key={cap}
                  style={{
                    fontSize: '11px',
                    padding: '1px 6px',
                    borderRadius: '4px',
                    background: '#f3f4f6',
                    color: '#555',
                  }}
                >
                  {cap}
                </span>
              ))}
              {a.capabilities.length > 3 && (
                <span style={{ fontSize: '11px', color: '#888' }}>
                  +{a.capabilities.length - 3}
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
