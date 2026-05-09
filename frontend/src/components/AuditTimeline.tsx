import type { AuditRecord } from '../types';

const riskColors: Record<string, string> = {
  low: '#22c55e',
  medium: '#f59e0b',
  high: '#ef4444',
  dangerous: '#dc2626',
};

interface Props {
  records: AuditRecord[];
}

export function AuditTimeline({ records }: Props) {
  const valid = records.every((_, i) => {
    if (i === 0) return true;
    return records[i - 1].record_hash === records[i].prev_hash;
  });

  return (
    <div>
      <div style={{ marginBottom: '12px', fontSize: '13px' }}>
        Chain integrity:{' '}
        <span
          style={{
            fontWeight: 600,
            color: valid ? '#22c55e' : '#ef4444',
          }}
        >
          {valid ? 'VALID' : 'BROKEN'}
        </span>
        <span style={{ marginLeft: 16, color: '#666' }}>
          {records.length} linked records (SHA256)
        </span>
      </div>

      <div style={{ maxHeight: 400, overflow: 'auto' }}>
        {records.map((r, i) => (
          <div
            key={r.id}
            style={{
              display: 'flex',
              gap: 12,
              padding: '8px 0',
              borderBottom: '1px solid #e5e7eb',
              fontSize: '13px',
            }}
          >
            <div style={{ width: 28, textAlign: 'center', color: '#888' }}>
              [{i}]
            </div>
            <span
              style={{
                display: 'inline-block',
                padding: '1px 6px',
                borderRadius: '10px',
                fontSize: '11px',
                background: riskColors[r.risk_level] || '#888',
                color: 'white',
                height: 'fit-content',
                minWidth: '60px',
                textAlign: 'center',
              }}
            >
              {r.risk_level}
            </span>
            <span style={{ fontWeight: 600, minWidth: '70px' }}>{r.agent_type}</span>
            <span style={{ flex: 1 }}>{r.decision}</span>
            <span style={{ fontSize: '11px', color: '#888' }}>
              {new Date(r.timestamp).toLocaleTimeString()}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
