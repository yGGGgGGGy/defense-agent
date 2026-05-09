import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from './api/client';
import type { Instance, AuditRecord, AgentInfo } from './types';
import { Sidebar } from './components/Sidebar';
import { ChatView } from './components/ChatView';
import { ConfigPanel } from './components/ConfigPanel';
import { AgentConfigPanel } from './components/AgentConfigPanel';

const agentNames: Record<string, string> = {
  perceiver:'感知',analyst:'分析',responder:'处置',operator:'运维',
  researcher:'研究',developer:'规划',executor:'执行',adviser:'监督',
  reflector:'反思',auditor:'审计',memorist:'记忆',
};

const sceneNames: Record<string, string> = {
  incident_response:'应急响应',ops_maintenance:'运维管理',
  penetration_test:'渗透测试',vulnerability_research:'漏洞挖掘',
  reverse_engineering:'逆向分析',
};

type Page = 'chat' | 'agents' | 'config';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  time: string;
  instanceId?: string;
  dagData?: Instance['dag'];
  auditData?: AuditRecord[];
  thinking?: Array<{ agent: string; agentCn: string; state: string; summary: string; time: string; confidence?: number }>;
}

export default function App() {
  const [page, setPage] = useState<Page>('chat');
  const [instances, setInstances] = useState<Instance[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [llmStatus, setLlmStatus] = useState('mock');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const chatEndRef = useRef<HTMLDivElement>(null!);

  const refreshInstances = useCallback(async () => {
    const insts = await api.getInstances();
    setInstances(insts);
  }, []);

  useEffect(() => {
    refreshInstances();
    api.getAgents().then(d => setAgents(d.agents));
    fetch('http://localhost:8100/v1/health').then(r => r.json()).then(d => setLlmStatus(d.mode || 'unknown')).catch(()=>{});
    const t = setInterval(refreshInstances, 2000);
    return () => clearInterval(t);
  }, [refreshInstances]);

  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);

  // Auto-configure DeepSeek on first load
  useEffect(() => {
    fetch('http://localhost:8100/v1/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        base_url: 'https://api.deepseek.com/anthropic',
        model: 'deepseek-v4-pro',
      }),
    }).catch(() => {});
  }, []);

  // Watch running instances for SSE updates
  const runningInst = instances.find(i => i.status === 'running');
  useEffect(() => {
    if (!runningInst) return;
    const es = new EventSource(`http://localhost:8080/api/v1/events?instance_id=${runningInst.id}`);
    es.addEventListener('node_state', (e) => {
      try {
        const d = JSON.parse(e.data);
        setMessages(prev => {
          const idx = prev.findIndex(m => m.instanceId === runningInst.id && m.role === 'assistant');
          if (idx < 0) return prev;
          const updated = { ...prev[idx] };
          const thinking = updated.thinking ? [...updated.thinking] : [];
          const cn = agentNames[d.agent_type] || d.agent_type;
          thinking.push({
            agent: d.agent_type, agentCn: cn,
            state: d.state, summary: d.summary || '',
            time: new Date().toLocaleTimeString('zh-CN'),
            confidence: d.confidence || 0,
          });
          updated.thinking = thinking;
          if (d.state === 'succeeded' || d.state === 'failed') {
            updated.content = `${cn} Agent ${d.state === 'succeeded' ? '已完成' : '失败'}: ${d.summary || ''}`;
          }
          const copy = [...prev];
          copy[idx] = updated;
          return copy;
        });
      } catch {}
    });
    es.addEventListener('instance_update', (e) => {
      try {
        const d = JSON.parse(e.data);
        if (d.status === 'done' || d.status === 'failed') {
          refreshInstances();
          const instId = runningInst.id;
          setTimeout(async () => {
            try {
              const audit = await api.getAuditTrail(instId);
              setMessages(prev => {
                const idx = prev.findIndex(m => m.instanceId === instId && m.role === 'assistant');
                if (idx < 0) return prev;
                const copy = [...prev];
                copy[idx] = { ...copy[idx], auditData: audit };
                return copy;
              });
            } catch {}
          }, 500);
        }
      } catch {}
    });
    es.onerror = () => es.close();
    return () => es.close();
  }, [runningInst?.id, refreshInstances]);

  const sendMessage = async (text: string) => {
    const userMsg: ChatMessage = {
      id: 'u' + Date.now(), role: 'user', content: text,
      time: new Date().toLocaleTimeString('zh-CN'),
    };
    const assistantMsg: ChatMessage = {
      id: 'a' + Date.now(), role: 'assistant', content: '正在分析任务...',
      time: new Date().toLocaleTimeString('zh-CN'), thinking: [],
    };
    setMessages(prev => [...prev, userMsg, assistantMsg]);
    setLoading(true);

    // Detect scene from natural language
    let scene = 'incident_response';
    const t = text.toLowerCase();
    if (t.includes('巡检') || t.includes('运维') || t.includes('备份') || t.includes('补丁') || t.includes('健康') || t.includes('审计') || t.includes('合规') || (t.includes('检查') && !t.includes('漏洞'))) scene = 'ops_maintenance';
    else if (t.includes('漏洞') || t.includes('cve') || t.includes('fuzz') || t.includes('挖掘')) scene = 'vulnerability_research';
    else if (t.includes('渗透') || t.includes('扫描')) scene = 'penetration_test';
    else if (t.includes('逆向') || t.includes('二进制') || t.includes('恶意')) scene = 'reverse_engineering';
    else if (t.includes('告警') || t.includes('应急') || t.includes('入侵') || t.includes('暴力') || t.includes('ssh') || t.includes('强制')) scene = 'incident_response';

    try {
      const inst = await api.createTask({ scene, title: text.slice(0, 60), description: text, input: text });
      setMessages(prev => {
        const copy = [...prev];
        const idx = copy.findIndex(m => m.id === assistantMsg.id);
        if (idx >= 0) {
          copy[idx] = { ...copy[idx], instanceId: inst.id, dagData: inst.dag, content: '任务已创建，Agent 开始执行...' };
        }
        return copy;
      });
      refreshInstances();
    } catch (e: any) {
      setMessages(prev => {
        const copy = [...prev];
        const idx = copy.findIndex(m => m.id === assistantMsg.id);
        if (idx >= 0) copy[idx] = { ...copy[idx], content: '任务提交失败: ' + e.message };
        return copy;
      });
    }
    setLoading(false);
  };

  // Update completed instances with final DAG data
  useEffect(() => {
    for (const inst of instances) {
      if (inst.status === 'done' || inst.status === 'failed') {
        setMessages(prev => {
          const idx = prev.findIndex(m => m.instanceId === inst.id && m.role === 'assistant');
          if (idx < 0) return prev;
          const copy = [...prev];
          copy[idx] = { ...copy[idx], dagData: inst.dag };
          return copy;
        });
      }
    }
  }, [instances]);

  return (
    <div style={{ display: 'flex', height: '100vh', background: '#0f1117', color: '#e1e4e8', fontFamily: 'system-ui, "Microsoft YaHei", sans-serif' }}>
      <Sidebar
        page={page}
        onPage={setPage}
        llmStatus={llmStatus}
        instanceCount={instances.length}
        agentCount={agents.length}
        selectedId={null}
        messages={messages}
        agentNames={agentNames}
        sceneNames={sceneNames}
      />

      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <div style={{ height: 48, background: '#161b22', borderBottom: '1px solid #21262d', display: 'flex', alignItems: 'center', padding: '0 20px', gap: 16, flexShrink: 0 }}>
          <span style={{ fontWeight: 700, fontSize: 15, color: '#58a6ff' }}>🛡 Defense Agent</span>
          <span style={{ fontSize: 12, color: '#8b949e' }}>自主决策通用防御智能体</span>
          <div style={{ flex: 1 }} />
          <span style={{ fontSize: 11, color: llmStatus === 'live' ? '#3fb950' : '#8b949e', background: '#21262d', padding: '3px 10px', borderRadius: 12 }}>
            {llmStatus === 'live' ? '🟢 DeepSeek Live' : '⚪ DeepSeek Mock'}
          </span>
        </div>

        {page === 'chat' && (
          <ChatView
            messages={messages}
            loading={loading}
            onSend={sendMessage}
            chatEndRef={chatEndRef}
            agentNames={agentNames}
            sceneNames={sceneNames}
          />
        )}
        {page === 'agents' && <AgentConfigPanel agents={agents} agentNames={agentNames} />}
        {page === 'config' && <ConfigPanel onStatusChange={setLlmStatus} />}
      </div>
    </div>
  );
}
