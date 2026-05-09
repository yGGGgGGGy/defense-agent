import type { Instance } from '../types';

const statusColor: Record<string, string> = { done: '#3fb950', running: '#58a6ff', failed: '#f85149', pending: '#d2991d' };
const statusName: Record<string, string> = { done: '完成', running: '运行中', failed: '失败', pending: '等待' };

interface Props {
  instances: Instance[];
  selectedId: string;
  onSelect: (id: string) => void;
  sceneNames: Record<string, string>;
}

export function InstanceList({ instances, selectedId, onSelect, sceneNames }: Props) {
  return (
    <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
      <div style={{ padding: '12px 14px', borderBottom: '1px solid #21262d', fontSize: 13, fontWeight: 600, color: '#c9d1d9' }}>
        实例列表 ({instances.length})
      </div>
      <div style={{ flex: 1, overflow: 'auto' }}>
        {instances.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: '#484f58', fontSize: 12 }}>暂无任务</div>
        )}
        {instances.map(inst => (
          <div
            key={inst.id}
            onClick={() => onSelect(inst.id)}
            style={{
              padding: '10px 14px', cursor: 'pointer', borderBottom: '1px solid #21262d',
              background: inst.id === selectedId ? '#1f6feb18' : 'transparent',
              borderLeft: inst.id === selectedId ? '3px solid #58a6ff' : '3px solid transparent',
            }}
          >
            <div style={{ fontSize: 13, fontWeight: 600, color: '#c9d1d9', marginBottom: 3 }}>
              {inst.task.title || inst.id.slice(0, 12)}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 11 }}>
              <span style={{ color: '#8b949e' }}>{sceneNames[inst.task.scene] || inst.task.scene}</span>
              <span style={{ color: statusColor[inst.status] || '#888', background: '#21262d', padding: '1px 8px', borderRadius: 8 }}>
                {statusName[inst.status] || inst.status}
              </span>
            </div>
            <div style={{ fontSize: 10, color: '#484f58', marginTop: 2 }}>{inst.id}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
