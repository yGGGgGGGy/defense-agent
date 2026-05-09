import { useState } from 'react';

interface Props {
  onSubmit: (scene: string, title: string, input: string) => void;
  sceneNames: Record<string, string>;
}

export function TaskPanel({ onSubmit, sceneNames }: Props) {
  const [scene, setScene] = useState('incident_response');
  const [title, setTitle] = useState('');
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!title || !input) return;
    setLoading(true);
    await onSubmit(scene, title, input);
    setLoading(false);
    setTitle('');
    setInput('');
  };

  const btnStyle = (loading: boolean): React.CSSProperties => ({
    padding: '8px 24px', borderRadius: 6, border: 'none', fontSize: 13, fontWeight: 600, cursor: loading ? 'default' : 'pointer',
    background: loading ? '#21262d' : '#238636', color: loading ? '#484f58' : 'white',
  });

  return (
    <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', padding: 16 }}>
      <div style={{ fontSize: 13, fontWeight: 600, color: '#c9d1d9', marginBottom: 10 }}>提交新任务</div>
      <div style={{ display: 'flex', gap: 8 }}>
        <select style={{ padding: '8px 12px', borderRadius: 6, border: '1px solid #30363d', background: '#0d1117', color: '#c9d1d9', fontSize: 13, minWidth: 140 }} value={scene} onChange={e => setScene(e.target.value)}>
          {Object.entries(sceneNames).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <input style={{ padding: '8px 12px', borderRadius: 6, border: '1px solid #30363d', background: '#0d1117', color: '#c9d1d9', fontSize: 13, flex: 1 }} placeholder="任务标题，如：SSH 暴力破解告警" value={title} onChange={e => setTitle(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSubmit()} />
        <input style={{ padding: '8px 12px', borderRadius: 6, border: '1px solid #30363d', background: '#0d1117', color: '#c9d1d9', fontSize: 13, flex: 2 }} placeholder="任务描述，如：检测到来自 10.0.0.50 的 150+ 次 SSH 登录失败" value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSubmit()} />
        <button style={btnStyle(loading)} onClick={handleSubmit} disabled={loading}>
          {loading ? '提交中...' : '▶ 执行'}
        </button>
      </div>
    </div>
  );
}
