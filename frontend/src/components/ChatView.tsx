import { useState } from 'react';
import type { ChatMessage } from '../App';
import { DagView } from './DagView';
import { AuditTimeline } from './AuditTimeline';

interface Props {
  messages: ChatMessage[];
  loading: boolean;
  onSend: (text: string) => void;
  chatEndRef: React.RefObject<HTMLDivElement>;
  agentNames: Record<string, string>;
  sceneNames: Record<string, string>;
}

const tColors: Record<string, string> = {
  thinking: '#58a6ff', running: '#d2991d', succeeded: '#3fb950', failed: '#f85149',
};

export function ChatView({ messages, loading, onSend, chatEndRef, agentNames, sceneNames }: Props) {
  const [input, setInput] = useState('');

  const handleSend = () => {
    if (!input.trim() || loading) return;
    onSend(input.trim());
    setInput('');
  };

  return (
    <>
      {/* Messages */}
      <div style={{ flex: 1, overflow: 'auto', padding: '20px 40px' }}>
        {messages.length === 0 && (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', color: '#484f58' }}>
            <div style={{ fontSize: 48, marginBottom: 16 }}>🛡</div>
            <div style={{ fontSize: 18, fontWeight: 600, color: '#8b949e', marginBottom: 8 }}>防御智能体系统</div>
            <div style={{ fontSize: 13, color: '#484f58', textAlign: 'center', maxWidth: 500, lineHeight: 1.8 }}>
              我是你的网络安全智能助手。直接用自然语言描述安全任务，<br/>
              我会自动识别场景并调度 Agent 执行。<br/>
              <br/>
              试试说：<br/>
              <span style={{ color: '#58a6ff' }}>"检测到 10.0.0.50 有大量 SSH 暴力破解，请处理"</span><br/>
              <span style={{ color: '#58a6ff' }}>"对 example.com 进行渗透测试"</span><br/>
              <span style={{ color: '#58a6ff' }}>"检查服务器健康状态并做合规审计"</span>
            </div>
          </div>
        )}

        {messages.map(msg => (
          <div key={msg.id} style={{ marginBottom: 20 }}>
            {/* User message */}
            {msg.role === 'user' && (
              <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <div style={{ maxWidth: '70%', background: '#1f6feb', color: 'white', padding: '12px 18px', borderRadius: '16px 16px 4px 16px', fontSize: 14, lineHeight: 1.6 }}>
                  {msg.content}
                </div>
              </div>
            )}

            {/* Assistant message */}
            {msg.role === 'assistant' && (
              <div style={{ display: 'flex', gap: 12 }}>
                <div style={{ width: 32, height: 32, borderRadius: '50%', background: '#238636', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16, flexShrink: 0 }}>🤖</div>
                <div style={{ flex: 1, maxWidth: '85%' }}>
                  <div style={{ background: '#161b22', border: '1px solid #21262d', borderRadius: '12px 12px 12px 4px', padding: '14px 18px', fontSize: 13, lineHeight: 1.7, color: '#c9d1d9' }}>
                    {msg.content}
                  </div>

                  {/* Thinking log */}
                  {msg.thinking && msg.thinking.length > 0 && (
                    <div style={{ marginTop: 8, background: '#0d1117', borderRadius: 8, border: '1px solid #21262d', overflow: 'hidden' }}>
                      <div style={{ padding: '8px 14px', borderBottom: '1px solid #21262d', fontSize: 11, color: '#8b949e', display: 'flex', justifyContent: 'space-between' }}>
                        <span>▎Agent 详细执行报告</span>
                        <span style={{ color: '#3fb950' }}>● LIVE</span>
                      </div>
                      <div style={{ maxHeight: 500, overflow: 'auto', padding: '10px 14px', fontFamily: '"JetBrains Mono", "Cascadia Code", monospace', fontSize: 11, lineHeight: 1.7 }}>
                        {msg.thinking.map((t, i) => (
                          <div key={i} style={{ marginBottom: 12 }}>
                            {/* Agent header */}
                            <div style={{
                              background: `${tColors[t.state] || '#58a6ff'}15`,
                              border: `1px solid ${tColors[t.state] || '#58a6ff'}33`,
                              borderRadius: 6, padding: '6px 10px', marginBottom: 4,
                            }}>
                              <span style={{ color: tColors[t.state], fontWeight: 700, fontSize: 12 }}>
                                {t.state === 'running' ? '●' : t.state === 'succeeded' ? '✓' : '▶'} [{t.agentCn} Agent]
                              </span>
                              <span style={{ color: '#484f58', marginLeft: 8 }}>{t.time}</span>
                              {(t.confidence ?? 0) > 0 && (
                                <span style={{ float: 'right', color: '#d2991d', fontSize: 10 }}>
                                  置信度: {((t.confidence ?? 0) * 100).toFixed(0)}%
                                </span>
                              )}
                            </div>
                            {/* Detailed thinking - multiline */}
                            <div style={{ color: '#8b949e', paddingLeft: 4, whiteSpace: 'pre-wrap' }}>
                              {t.summary}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* DAG */}
                  {msg.dagData && (
                    <div style={{ marginTop: 8 }}>
                      <DagView dag={msg.dagData} agentNames={agentNames} />
                    </div>
                  )}

                  {/* Audit */}
                  {msg.auditData && msg.auditData.length > 0 && (
                    <div style={{ marginTop: 8 }}>
                      <AuditTimeline records={msg.auditData} agentNames={agentNames} />
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* System message */}
            {msg.role === 'system' && (
              <div style={{ textAlign: 'center', fontSize: 11, color: '#484f58', padding: '4px 0' }}>
                {msg.content}
              </div>
            )}
          </div>
        ))}
        <div ref={chatEndRef} />
      </div>

      {/* Input */}
      <div style={{ padding: '16px 40px 20px', borderTop: '1px solid #21262d', background: '#161b22', flexShrink: 0 }}>
        <div style={{ display: 'flex', gap: 10, maxWidth: 900, margin: '0 auto' }}>
          <input
            style={{
              flex: 1, padding: '12px 18px', borderRadius: 12, border: '1px solid #30363d',
              background: '#0d1117', color: '#c9d1d9', fontSize: 14, outline: 'none',
              boxShadow: '0 0 0 1px #1f6feb00', transition: 'box-shadow .2s',
            }}
            placeholder="描述你的安全任务，如：检测到 SSH 暴力破解，IP 为 10.0.0.50，请立即处理..."
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend(); } }}
            disabled={loading}
            onFocus={e => e.target.style.boxShadow = '0 0 0 2px #1f6feb44'}
            onBlur={e => e.target.style.boxShadow = '0 0 0 1px #1f6feb00'}
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            style={{
              padding: '12px 28px', borderRadius: 12, border: 'none', fontSize: 14, fontWeight: 600,
              cursor: loading || !input.trim() ? 'default' : 'pointer',
              background: loading || !input.trim() ? '#21262d' : '#238636',
              color: loading || !input.trim() ? '#484f58' : 'white',
              transition: 'all .15s',
            }}
          >
            {loading ? '执行中...' : '发送'}
          </button>
        </div>
        <div style={{ textAlign: 'center', fontSize: 10, color: '#484f58', marginTop: 8 }}>
          Enter 发送 · 自动识别应急/渗透/挖洞/逆向/运维场景 · DeepSeek V4 Pro
        </div>
      </div>
    </>
  );
}
