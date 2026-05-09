import type { ChatMessage } from '../App';
import { useState } from 'react';

interface Props {
  page: string;
  onPage: (p: 'chat' | 'agents' | 'config') => void;
  llmStatus: string;
  instanceCount: number;
  agentCount: number;
  selectedId: string | null;
  messages: ChatMessage[];
  agentNames: Record<string, string>;
  sceneNames: Record<string, string>;
}

export function Sidebar({ page, onPage, llmStatus, instanceCount, agentCount, messages }: Props) {
  const items: Array<{ id: 'chat' | 'agents' | 'config'; label: string; icon: string; badge?: string }> = [
    { id: 'chat', label: '对话', icon: '💬', badge: messages.length > 0 ? String(messages.filter(m => m.role === 'user').length) : undefined },
    { id: 'agents', label: 'Agent', icon: '🤖', badge: String(agentCount) },
    { id: 'config', label: 'LLM 设置', icon: '⚙', badge: llmStatus === 'live' ? 'Live' : 'Mock' },
  ];

  return (
    <div style={{ width: 200, background: '#161b22', borderRight: '1px solid #21262d', display: 'flex', flexDirection: 'column', flexShrink: 0 }}>
      <div style={{ padding: '18px 14px 12px', borderBottom: '1px solid #21262d' }}>
        <div style={{ fontSize: 15, fontWeight: 800, color: '#58a6ff' }}>🛡 Defense Agent</div>
        <div style={{ fontSize: 10, color: '#484f58', marginTop: 2 }}>DeepSeek V4 Pro</div>
      </div>

      <div style={{ flex: 1, padding: '8px 0' }}>
        {items.map(item => (
          <div
            key={item.id}
            onClick={() => onPage(item.id)}
            style={{
              display: 'flex', alignItems: 'center', gap: 10, padding: '10px 14px', cursor: 'pointer',
              background: page === item.id ? '#1f6feb18' : 'transparent',
              borderLeft: page === item.id ? '3px solid #58a6ff' : '3px solid transparent',
              color: page === item.id ? '#58a6ff' : '#8b949e', fontSize: 13, fontWeight: page === item.id ? 600 : 400,
              transition: 'all .12s',
            }}
          >
            <span>{item.icon}</span>
            <span>{item.label}</span>
            {item.badge && (
              <span style={{ marginLeft: 'auto', fontSize: 10, background: '#21262d', padding: '1px 7px', borderRadius: 10, color: '#8b949e' }}>
                {item.badge}
              </span>
            )}
          </div>
        ))}
      </div>

      <div style={{ padding: '12px 14px', borderTop: '1px solid #21262d', fontSize: 10, color: '#484f58' }}>
        <div style={{ marginBottom: 2 }}>模型: deepseek-v4-pro</div>
        <div style={{ marginBottom: 2 }}>API: deepseek/anthropic</div>
        <div>实例: {instanceCount} | Agent: {agentCount}</div>
      </div>
    </div>
  );
}
