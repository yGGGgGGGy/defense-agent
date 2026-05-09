import { useMemo } from 'react';
import type { Dag } from '../types';

const nodeColors: Record<string, string> = {
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

const stateColors: Record<string, string> = {
  pending: '#d1d5db',
  running: '#3b82f6',
  succeeded: '#22c55e',
  failed: '#ef4444',
  skipped: '#f59e0b',
};

interface Props {
  dag: Dag;
}

export function DagView({ dag }: Props) {
  const nodes = useMemo(() => Object.values(dag.nodes), [dag.nodes]);

  // Simple top-down layout
  const layers = useMemo(() => {
    const visited = new Set<string>();
    const result: Array<Array<typeof nodes[0]>> = [];

    const walk = (nodeIds: string[]) => {
      const current: typeof nodes = [];
      const next: string[] = [];
      for (const id of nodeIds) {
        if (visited.has(id)) continue;
        visited.add(id);
        const n = dag.nodes[id];
        if (n) {
          current.push(n);
          if (n.dependencies) {
            next.push(...n.dependencies);
          }
        }
      }
      if (current.length > 0) {
        result.push(current);
      }
    };

    // Start from leaves (nodes that no one depends on)
    const isDepOf: Record<string, boolean> = {};
    for (const n of nodes) {
      if (n.dependencies) {
        for (const d of n.dependencies) {
          isDepOf[d] = true;
        }
      }
    }
    const leaves = nodes.filter((n) => !isDepOf[n.id]).map((n) => n.id);
    walk(leaves);
    // Then walk remaining
    for (const n of nodes) {
      if (!visited.has(n.id)) {
        walk([n.id]);
      }
    }
    return result.reverse(); // top-down
  }, [dag, nodes]);

  return (
    <div style={{ padding: '10px 0' }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12, alignItems: 'center' }}>
        {layers.map((layer, li) => (
          <div key={li}>
            {li > 0 && (
              <div style={{ textAlign: 'center', color: '#888', fontSize: '20px', marginBottom: 4 }}>
                ↓
              </div>
            )}
            <div style={{ display: 'flex', gap: 16, justifyContent: 'center' }}>
              {layer.map((node) => (
                <div
                  key={node.id}
                  style={{
                    padding: '10px 16px',
                    borderRadius: '8px',
                    border: `2px solid ${nodeColors[node.agent_type] || '#666'}`,
                    background: 'white',
                    minWidth: '120px',
                    textAlign: 'center',
                  }}
                >
                  <div style={{ fontWeight: 600, fontSize: '14px' }}>
                    {node.agent_type}
                  </div>
                  <div style={{ fontSize: '11px', color: '#666', margin: '4px 0' }}>
                    {node.id.split('-')[1]?.slice(0, 8)}
                  </div>
                  <span
                    style={{
                      display: 'inline-block',
                      padding: '2px 8px',
                      borderRadius: '10px',
                      fontSize: '11px',
                      background: stateColors[node.state] || '#888',
                      color: 'white',
                    }}
                  >
                    {node.state}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Flow direction */}
      <div style={{ marginTop: 20, fontSize: '12px', color: '#888', textAlign: 'center' }}>
        Phase 1 (IR): Perceiver → Analyst → (Responder ∥ Operator)
      </div>
    </div>
  );
}
