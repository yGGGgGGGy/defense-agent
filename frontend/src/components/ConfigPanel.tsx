import { useState } from 'react';

const providers = [
  { id: 'openai', name: 'OpenAI', base: 'https://api.openai.com/v1', models: ['gpt-4o', 'gpt-4.1', 'gpt-5.1', 'o4-mini'] },
  { id: 'deepseek', name: 'DeepSeek', base: 'https://api.deepseek.com/v1', models: ['deepseek-chat', 'deepseek-reasoner'] },
  { id: 'anthropic', name: 'Anthropic', base: 'https://api.anthropic.com/v1', models: ['claude-opus-4-7', 'claude-sonnet-4-6', 'claude-haiku-4-5'] },
  { id: 'ollama', name: 'Ollama 本地', base: 'http://localhost:11434/v1', models: ['llama3', 'qwen3', 'mistral', 'deepseek-r1'] },
  { id: 'glm', name: '智谱 GLM', base: 'https://open.bigmodel.cn/api/paas/v4', models: ['glm-4-plus', 'glm-4-flash'] },
  { id: 'custom', name: '自定义', base: '', models: [] },
];

interface Props { onStatusChange: (mode: string) => void; }

const iStyle: React.CSSProperties = { padding: '8px 12px', borderRadius: 6, border: '1px solid #30363d', background: '#0d1117', color: '#c9d1d9', fontSize: 13, width: '100%', boxSizing: 'border-box' };
const labelStyle: React.CSSProperties = { fontSize: 11, color: '#8b949e', display: 'block', marginBottom: 4 };

export function ConfigPanel({ onStatusChange }: Props) {
  const [provider, setProvider] = useState(providers[0]);
  const [model, setModel] = useState(provider.models[0] || '');
  const [customModel, setCustomModel] = useState('');
  const [customBase, setCustomBase] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [status, setStatus] = useState('');
  const [testing, setTesting] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);

  const currentModel = provider.id === 'custom' ? customModel : model;
  const currentBase = provider.id === 'custom' ? customBase : provider.base;

  const testConnection = async () => {
    setTesting(true); setStatus(''); setLogs([]);
    const addLog = (msg: string) => setLogs(prev => [...prev, `[${new Date().toLocaleTimeString('zh-CN')}] ${msg}`]);
    try {
      addLog(`正在连接 ${provider.name}...`);
      const res = await fetch('http://localhost:8100/v1/config', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ api_key: apiKey, base_url: currentBase, model: currentModel }),
      });
      const data = await res.json();
      addLog(`配置响应: ${JSON.stringify(data)}`);
      const hres = await fetch('http://localhost:8100/v1/health');
      const hdata = await hres.json();
      addLog(`健康检查: mode=${hdata.mode}, model=${hdata.model}`);
      onStatusChange(hdata.mode);
      setStatus(hdata.mode === 'live' ? '✅ 已连接' : '⚠️ ' + hdata.mode);
    } catch (e: any) {
      addLog(`错误: ${e.message}`);
      setStatus('❌ 连接失败');
    }
    setTesting(false);
  };

  return (
    <div style={{ maxWidth: 700, margin: '0 auto' }}>
      <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', padding: 24, marginBottom: 16 }}>
        <h2 style={{ margin: '0 0 20px', fontSize: 18, color: '#c9d1d9' }}>⚙️ LLM 模型配置</h2>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, marginBottom: 14 }}>
          <div>
            <label style={labelStyle}>提供商</label>
            <select style={iStyle} value={provider.id} onChange={e => {
              const p = providers.find(x => x.id === e.target.value)!;
              setProvider(p); setModel(p.models[0] || '');
            }}>
              {providers.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
          </div>
          <div>
            <label style={labelStyle}>{provider.models.length > 0 ? '模型选择' : '模型名称'}</label>
            {provider.models.length > 0 ? (
              <select style={iStyle} value={model} onChange={e => setModel(e.target.value)}>
                {provider.models.map(m => <option key={m} value={m}>{m}</option>)}
              </select>
            ) : (
              <input style={iStyle} value={customModel} onChange={e => setCustomModel(e.target.value)} placeholder="输入模型名称" />
            )}
          </div>
          {provider.id === 'custom' && (
            <div style={{ gridColumn: '1/-1' }}>
              <label style={labelStyle}>API Base URL</label>
              <input style={iStyle} value={customBase} onChange={e => setCustomBase(e.target.value)} placeholder="https://api.example.com/v1" />
            </div>
          )}
          <div style={{ gridColumn: '1/-1' }}>
            <label style={labelStyle}>API Key</label>
            <input type="password" style={iStyle} value={apiKey} onChange={e => setApiKey(e.target.value)} placeholder="sk-..." />
          </div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button onClick={testConnection} disabled={testing || !apiKey}
            style={{ padding: '8px 24px', borderRadius: 6, border: 'none', fontSize: 13, fontWeight: 600, cursor: testing ? 'default' : 'pointer',
              background: testing ? '#21262d' : '#238636', color: testing ? '#484f58' : 'white' }}>
            {testing ? '测试中...' : '连接测试'}
          </button>
          <span style={{ fontSize: 13, color: status.includes('✅') ? '#3fb950' : status.includes('❌') ? '#f85149' : '#8b949e' }}>
            {status || (apiKey ? '点击测试' : '不填 Key = Mock 模式')}
          </span>
        </div>
      </div>

      {/* Log output */}
      {logs.length > 0 && (
        <div style={{ background: '#0d1117', borderRadius: 8, border: '1px solid #21262d', padding: 14 }}>
          <div style={{ fontSize: 12, color: '#8b949e', marginBottom: 8 }}>连接日志</div>
          <div style={{ fontFamily: 'monospace', fontSize: 11, lineHeight: 1.6 }}>
            {logs.map((l, i) => <div key={i} style={{ color: l.includes('错误') || l.includes('❌') ? '#f85149' : '#3fb950' }}>{l}</div>)}
          </div>
        </div>
      )}
    </div>
  );
}
