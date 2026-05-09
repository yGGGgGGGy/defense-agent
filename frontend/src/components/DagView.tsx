import { useMemo } from 'react';
import type { Dag } from '../types';

const nColors: Record<string, string> = {
  perceiver:'#3b82f6',analyst:'#8b5cf6',responder:'#ef4444',operator:'#10b981',
  researcher:'#f59e0b',developer:'#ec4899',executor:'#6366f1',adviser:'#14b8a6',
  reflector:'#f97316',auditor:'#06b6d4',memorist:'#84cc16',
};
const sColors: Record<string, string> = { pending:'#484f58',running:'#58a6ff',succeeded:'#3fb950',failed:'#f85149',skipped:'#d2991d' };
const sNames: Record<string, string> = { pending:'等待',running:'执行中',succeeded:'成功',failed:'失败',skipped:'跳过' };

interface Props { dag: Dag; agentNames: Record<string, string>; }

export function DagView({ dag, agentNames }: Props) {
  const nodes = useMemo(() => Object.values(dag.nodes), [dag.nodes]);
  const layers = useMemo(() => {
    const visited = new Set<string>();
    const result: Array<Array<typeof nodes[0]>> = [];
    const walk = (ids: string[]) => {
      const cur: typeof nodes = []; const nxt: string[] = [];
      for (const id of ids) {
        if (visited.has(id)) continue; visited.add(id);
        const n = dag.nodes[id]; if (n) { cur.push(n); if (n.dependencies) nxt.push(...n.dependencies); }
      }
      if (cur.length) result.push(cur);
    };
    const isDep: Record<string, boolean> = {};
    for (const n of nodes) { if (n.dependencies) for (const d of n.dependencies) isDep[d] = true; }
    walk(nodes.filter(n => !isDep[n.id]).map(n => n.id));
    for (const n of nodes) { if (!visited.has(n.id)) walk([n.id]); }
    return result.reverse();
  }, [dag, nodes]);

  return (
    <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', padding: 16 }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: '#c9d1d9', marginBottom: 12 }}>DAG 执行流程</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, alignItems: 'center' }}>
        {layers.map((layer, li) => (
          <div key={li}>
            {li > 0 && <div style={{ textAlign: 'center', color: '#30363d', fontSize: 16, marginBottom: 4 }}>↓</div>}
            <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
              {layer.map(node => {
                const isPar = layer.length > 1;
                return (
                  <div key={node.id} style={{
                    padding: '10px 18px', borderRadius: 8, background: '#0d1117',
                    border: `2px solid ${(nColors[node.agent_type] || '#666')}44`,
                    borderTop: `3px solid ${nColors[node.agent_type] || '#666'}`,
                    textAlign: 'center', minWidth: 90,
                    boxShadow: isPar ? `0 0 12px ${(nColors[node.agent_type] || '#666')}22` : 'none',
                  }}>
                    <div style={{ fontWeight: 700, fontSize: 13, color: nColors[node.agent_type] }}>
                      {agentNames[node.agent_type] || node.agent_type}
                    </div>
                    <div style={{ fontSize: 10, color: '#484f58', margin: '2px 0' }}>{node.agent_type}</div>
                    <span style={{
                      display: 'inline-block', padding: '2px 8px', borderRadius: 8, fontSize: 10,
                      background: `${sColors[node.state] || '#888'}22`, color: sColors[node.state], fontWeight: 600,
                    }}>
                      {sNames[node.state] || node.state}
                    </span>
                  </div>
                );
              })}
            </div>
            {layer.length > 1 && (
              <div style={{ textAlign: 'center', fontSize: 10, color: '#d2991d', marginTop: 2 }}>∥ 并行执行</div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
