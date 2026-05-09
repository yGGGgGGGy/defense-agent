import { useEffect, useRef, useState } from 'react';

interface Entry {
  id: number; time: string; agent: string; agentCn: string;
  message: string; type: 'thinking' | 'action' | 'result' | 'error';
}

interface Props { instanceId: string | null; agentNames: Record<string, string>; }

const typeStyle: Record<string, { color: string; icon: string }> = {
  thinking: { color: '#58a6ff', icon: '●' },
  action:   { color: '#d2991d', icon: '▶' },
  result:   { color: '#3fb950', icon: '✓' },
  error:    { color: '#f85149', icon: '✗' },
};

export function ThinkingTerminal({ instanceId, agentNames }: Props) {
  const [entries, setEntries] = useState<Entry[]>([]);
  const bottomRef = useRef<HTMLDivElement>(null);
  const seq = useRef(0);

  useEffect(() => {
    if (!instanceId) { setEntries([]); return; }
    setEntries([]); seq.current = 0;

    const es = new EventSource(`http://localhost:8080/api/v1/events?instance_id=${instanceId}`);
    es.addEventListener('node_state', (e) => {
      try {
        const d = JSON.parse(e.data);
        seq.current++;
        const cn = agentNames[d.agent_type] || d.agent_type;
        let type: Entry['type'] = 'action';
        let msg = `${cn} Agent 开始执行`;
        if (d.state === 'running') { type = 'thinking'; msg = `${cn} 正在分析处理中...`; }
        else if (d.state === 'succeeded') { type = 'result'; msg = `${cn} 完成 → ${(d.summary || '').slice(0, 80)}`; }
        else if (d.state === 'failed') { type = 'error'; msg = `${cn} 失败: ${d.error || d.summary || ''}`; }
        setEntries(prev => [...prev.slice(-199), {
          id: seq.current, time: new Date().toLocaleTimeString('zh-CN'),
          agent: d.agent_type, agentCn: cn, message: msg, type,
        }]);
      } catch {}
    });
    es.onerror = () => es.close();
    return () => es.close();
  }, [instanceId, agentNames]);

  useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [entries]);

  if (!instanceId) return null;

  return (
    <div style={{ background: '#0d1117', borderRadius: 8, border: '1px solid #21262d', overflow: 'hidden' }}>
      <div style={{ padding: '8px 14px', borderBottom: '1px solid #21262d', fontSize: 12, color: '#8b949e', display: 'flex', justifyContent: 'space-between' }}>
        <span>▎实时思考过程 — {instanceId}</span>
        <span style={{ color: '#3fb950' }}>● LIVE</span>
      </div>
      <div style={{ maxHeight: 260, overflow: 'auto', padding: '10px 14px', fontFamily: '"JetBrains Mono", "Cascadia Code", monospace', fontSize: 12, lineHeight: 1.7 }}>
        {entries.map(e => (
          <div key={e.id} style={{ color: typeStyle[e.type].color, marginBottom: 2 }}>
            <span style={{ color: '#484f58' }}>{e.time}</span>{' '}
            <span style={{ fontWeight: 600 }}>[{e.agentCn}]</span>{' '}
            <span style={{ color: typeStyle[e.type].color }}>{typeStyle[e.type].icon}</span>{' '}
            <span style={{ color: '#c9d1d9' }}>{e.message}</span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
