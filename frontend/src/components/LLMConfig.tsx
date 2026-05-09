import { useState } from 'react';

const providers = [
  { id: 'openai', name: 'OpenAI', base: 'https://api.openai.com/v1', models: ['gpt-4o', 'gpt-4.1', 'gpt-5.1', 'o4-mini'] },
  { id: 'deepseek', name: 'DeepSeek', base: 'https://api.deepseek.com/v1', models: ['deepseek-chat', 'deepseek-reasoner'] },
  { id: 'anthropic', name: 'Anthropic', base: 'https://api.anthropic.com/v1', models: ['claude-sonnet-4-6', 'claude-opus-4-7', 'claude-haiku-4-5'] },
  { id: 'ollama', name: 'Ollama', base: 'http://localhost:11434/v1', models: ['llama3', 'qwen3', 'mistral'] },
  { id: 'custom', name: '自定义', base: '', models: [] },
];

interface Props {
  onConfigured: (info: { provider: string; model: string }) => void;
}

export function LLMConfig({ onConfigured }: Props) {
  const [provider, setProvider] = useState(providers[0]);
  const [model, setModel] = useState(provider.models[0] || '');
  const [customModel, setCustomModel] = useState('');
  const [customBase, setCustomBase] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [status, setStatus] = useState('');
  const [testing, setTesting] = useState(false);
  const [expanded, setExpanded] = useState(false);

  const currentModel = provider.id === 'custom' ? customModel : model;
  const currentBase = provider.id === 'custom' ? customBase : provider.base;

  const testConnection = async () => {
    setTesting(true);
    setStatus('配置中...');
    try {
      const res = await fetch('http://localhost:8100/v1/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ api_key: apiKey, base_url: currentBase, model: currentModel }),
      });
      const data = await res.json();
      if (data.status === 'configured') {
        const hres = await fetch('http://localhost:8100/v1/health');
        const hdata = await hres.json();
        setStatus(`${hdata.mode === 'live' ? '✅ 已连接' : '⚠️ ' + hdata.mode} (${hdata.model})`);
        onConfigured({ provider: provider.name, model: currentModel });
      } else {
        setStatus('❌ 配置失败');
      }
    } catch {
      setStatus('❌ AI 服务未响应');
    }
    setTesting(false);
  };

  if (!expanded) {
    return (
      <div
        style={{
          background: '#1e3a5f', color: 'white', padding: '8px 16px', borderRadius: 6,
          cursor: 'pointer', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          marginBottom: 16, fontSize: 14,
        }}
        onClick={() => setExpanded(true)}
      >
        <span>⚙️ LLM 配置 {status && `— ${status}`}</span>
        <span style={{ fontSize: 12, opacity: 0.6 }}>点击展开</span>
      </div>
    );
  }

  return (
    <div style={{ background: 'white', borderRadius: 8, padding: 16, marginBottom: 16, boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <h3 style={{ margin: 0, fontSize: 16, color: '#1e3a5f' }}>⚙️ LLM 模型配置</h3>
        <button
          onClick={() => setExpanded(false)}
          style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 18, color: '#888' }}
        >
          ✕
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 10 }}>
        <div>
          <label style={{ fontSize: 12, color: '#666', display: 'block', marginBottom: 4 }}>提供商</label>
          <select
            style={{ width: '100%', padding: '6px 10px', borderRadius: 4, border: '1px solid #d1d5db', fontSize: 13 }}
            value={provider.id}
            onChange={(e) => {
              const p = providers.find((x) => x.id === e.target.value)!;
              setProvider(p);
              setModel(p.models[0] || '');
            }}
          >
            {providers.map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label style={{ fontSize: 12, color: '#666', display: 'block', marginBottom: 4 }}>
            {provider.id === 'custom' ? '模型名称' : '模型'}
          </label>
          {provider.models.length > 0 ? (
            <select
              style={{ width: '100%', padding: '6px 10px', borderRadius: 4, border: '1px solid #d1d5db', fontSize: 13 }}
              value={model}
              onChange={(e) => setModel(e.target.value)}
            >
              {provider.models.map((m) => (
                <option key={m} value={m}>{m}</option>
              ))}
            </select>
          ) : (
            <input
              style={{ width: '100%', padding: '6px 10px', borderRadius: 4, border: '1px solid #d1d5db', fontSize: 13 }}
              value={customModel}
              onChange={(e) => setCustomModel(e.target.value)}
              placeholder="输入模型名"
            />
          )}
        </div>
        {provider.id === 'custom' && (
          <div style={{ gridColumn: '1 / -1' }}>
            <label style={{ fontSize: 12, color: '#666', display: 'block', marginBottom: 4 }}>API 地址</label>
            <input
              style={{ width: '100%', padding: '6px 10px', borderRadius: 4, border: '1px solid #d1d5db', fontSize: 13 }}
              value={customBase}
              onChange={(e) => setCustomBase(e.target.value)}
              placeholder="https://api.example.com/v1"
            />
          </div>
        )}
      </div>

      <div style={{ marginBottom: 10 }}>
        <label style={{ fontSize: 12, color: '#666', display: 'block', marginBottom: 4 }}>API Key</label>
        <input
          type="password"
          style={{ width: '100%', padding: '6px 10px', borderRadius: 4, border: '1px solid #d1d5db', fontSize: 13 }}
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="sk-..."
        />
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <button
          onClick={testConnection}
          disabled={testing || !apiKey}
          style={{
            padding: '6px 20px', background: testing ? '#888' : '#059669', color: 'white',
            border: 'none', borderRadius: 4, cursor: testing ? 'default' : 'pointer', fontSize: 13,
          }}
        >
          {testing ? '测试中...' : '连接测试'}
        </button>
        <span style={{ fontSize: 13, color: status.includes('✅') ? '#059669' : status.includes('❌') ? '#dc2626' : '#666' }}>
          {status || (apiKey ? '点击测试连接' : '不填 Key 使用 Mock 模式')}
        </span>
      </div>
    </div>
  );
}
