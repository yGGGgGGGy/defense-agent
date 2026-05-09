import type { AuditRecord } from '../types';

const rColors: Record<string, string> = { low: '#3fb950', medium: '#d2991d', high: '#f85149', dangerous: '#ff0000' };
const rNames: Record<string, string> = { low: '低', medium: '中', high: '高', dangerous: '危险' };

interface Props { records: AuditRecord[]; agentNames: Record<string, string>; }

export function AuditTimeline({ records, agentNames }: Props) {
  const valid = records.every((_, i) => i === 0 || records[i-1].record_hash === records[i].prev_hash);

  return (
    <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', padding: 16 }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: '#c9d1d9', marginBottom: 8 }}>
        审计链 · {records.length} 条记录
      </div>
      <div style={{ marginBottom: 10, fontSize: 12 }}>
        <span style={{ color: '#8b949e' }}>链完整性: </span>
        <span style={{ fontWeight: 600, color: valid ? '#3fb950' : '#f85149' }}>
          {valid ? '✓ 有效' : '✗ 断裂'}
        </span>
        <span style={{ color: '#484f58', marginLeft: 16 }}>SHA256 哈希链表</span>
      </div>
      <div style={{ maxHeight: 320, overflow: 'auto' }}>
        {records.map((r, i) => (
          <div key={r.id} style={{
            display: 'flex', gap: 10, padding: '7px 0', borderBottom: '1px solid #21262d',
            fontSize: 12, alignItems: 'center',
          }}>
            <span style={{ color: '#484f58', width: 20, flexShrink: 0, textAlign: 'right' }}>{i}</span>
            <span style={{
              padding: '1px 6px', borderRadius: 8, fontSize: 10, background: `${rColors[r.risk_level]}22`,
              color: rColors[r.risk_level], fontWeight: 600, minWidth: 28, textAlign: 'center', flexShrink: 0,
            }}>{rNames[r.risk_level] || r.risk_level}</span>
            <span style={{ fontWeight: 600, minWidth: 40, flexShrink: 0, color: '#c9d1d9' }}>
              {agentNames[r.agent_type] || r.agent_type}
            </span>
            <span style={{ flex: 1, color: '#8b949e' }}>{r.decision}</span>
            <span style={{ fontSize: 10, color: '#484f58', flexShrink: 0 }}>{new Date(r.timestamp).toLocaleTimeString('zh-CN')}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
