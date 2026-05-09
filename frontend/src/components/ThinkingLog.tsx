import { useEffect, useRef, useState } from 'react';

interface ThinkingEntry {
  id: number;
  time: string;
  agent: string;
  agentCn: string;
  message: string;
  type: 'thinking' | 'action' | 'result' | 'error';
}

interface Props {
  instanceId: string | null;
  agentNames: Record<string, string>;
}

export function ThinkingLog({ instanceId, agentNames }: Props) {
  const [entries, setEntries] = useState<ThinkingEntry[]>([]);
  const bottomRef = useRef<HTMLDivElement>(null);
  const seqRef = useRef(0);

  useEffect(() => {
    if (!instanceId) return;
    setEntries([]);
    seqRef.current = 0;

    const url = `http://localhost:8080/api/v1/events?instance_id=${instanceId}`;
    const es = new EventSource(url);

    es.addEventListener('node_state', (e) => {
      try {
        const d = JSON.parse(e.data);
        seqRef.current++;
        const cn = agentNames[d.agent_type] || d.agent_type;
        let type: ThinkingEntry['type'] = 'action';
        let msg = `${cn} Agent 执行中...`;
        if (d.state === 'succeeded') {
          type = 'result';
          msg = `${cn} 完成: ${d.summary || ''}`;
        } else if (d.state === 'failed') {
          type = 'error';
          msg = `${cn} 失败: ${d.error || d.summary || ''}`;
        } else if (d.state === 'running') {
          type = 'thinking';
          msg = `${cn} 正在分析处理...`;
        }
        setEntries((prev) => [...prev.slice(-99), {
          id: seqRef.current,
          time: new Date().toLocaleTimeString('zh-CN'),
          agent: d.agent_type,
          agentCn: cn,
          message: msg,
          type,
        }]);
      } catch {}
    });

    es.addEventListener('instance_update', (e) => {
      try {
        const d = JSON.parse(e.data);
        seqRef.current++;
        setEntries((prev) => [...prev.slice(-99), {
          id: seqRef.current,
          time: new Date().toLocaleTimeString('zh-CN'),
          agent: 'system',
          agentCn: '系统',
          message: `实例状态: ${d.status}`,
          type: 'result',
        }]);
      } catch {}
    });

    es.onerror = () => es.close();
    return () => es.close();
  }, [instanceId, agentNames]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [entries]);

  if (!instanceId) return null;

  const typeStyles: Record<string, React.CSSProperties> = {
    thinking: { color: '#1a56db', borderLeftColor: '#1a56db' },
    action: { color: '#d97706', borderLeftColor: '#d97706' },
    result: { color: '#059669', borderLeftColor: '#059669' },
    error: { color: '#dc2626', borderLeftColor: '#dc2626' },
  };

  const agentColors: Record<string, string> = {
    perceiver: '#3b82f6', analyst: '#8b5cf6', responder: '#ef4444',
    operator: '#10b981', researcher: '#f59e0b', developer: '#ec4899',
    executor: '#6366f1', adviser: '#14b8a6', reflector: '#f97316',
    auditor: '#06b6d4', memorist: '#84cc16', system: '#888',
  };

  return (
    <div
      style={{
        background: '#1a1a2e', color: '#e0e0e0', borderRadius: 8, padding: 12,
        fontFamily: '"JetBrains Mono", "Courier New", monospace', fontSize: 12,
        maxHeight: 350, overflow: 'auto', lineHeight: 1.6,
      }}
    >
      <div style={{ color: '#888', marginBottom: 8, borderBottom: '1px solid #333', paddingBottom: 8 }}>
        ▎实时思考过程 — {instanceId}
      </div>
      {entries.length === 0 && (
        <div style={{ color: '#555' }}>等待 Agent 响应...</div>
      )}
      {entries.map((entry) => (
        <div
          key={entry.id}
          style={{
            borderLeft: '2px solid #333',
            paddingLeft: 8,
            marginBottom: 6,
            ...typeStyles[entry.type],
          }}
        >
          <span style={{ color: '#555' }}>{entry.time}</span>{' '}
          <span style={{ color: agentColors[entry.agent] || '#888', fontWeight: 600 }}>
            [{entry.agentCn}]
          </span>{' '}
          {entry.message}
        </div>
      ))}
      <div ref={bottomRef} />
    </div>
  );
}
