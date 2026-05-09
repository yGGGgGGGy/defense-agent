import type { AgentInfo } from '../types';

const agentColors: Record<string, string> = {
  perceiver: '#3b82f6', analyst: '#8b5cf6', responder: '#ef4444', operator: '#10b981',
  researcher: '#f59e0b', developer: '#ec4899', executor: '#6366f1', adviser: '#14b8a6',
  reflector: '#f97316', auditor: '#06b6d4', memorist: '#84cc16',
};

const capNames: Record<string, string> = {
  log_collection:'日志采集',traffic_analysis:'流量分析',asset_discovery:'资产发现',alert_aggregation:'告警聚合',
  alert_correlation:'告警关联',attack_chain_mapping:'攻击链映射',root_cause_analysis:'根因分析',impact_assessment:'影响评估',attck_mapping:'ATT&CK映射',
  ip_blocking:'IP封禁',host_isolation:'主机隔离',patch_deployment:'补丁部署',service_recovery:'服务恢复',
  health_check:'健康巡检',config_compliance:'配置合规',patch_management:'补丁管理',backup_restore:'备份恢复',
  capacity_planning:'容量规划',service_orchestration:'服务编排',log_rotation:'日志轮转',certificate_management:'证书管理',
  threat_intel:'威胁情报',cve_lookup:'CVE查询',ioc_correlation:'IOC关联',exploit_search:'漏洞搜索',
  attack_planning:'攻击规划',exploit_selection:'工具选择',tool_recommendation:'工具推荐',strategy_generation:'策略生成',
  tool_execution:'工具执行',command_dispatch:'命令调度',result_collection:'结果收集',sandbox_orchestration:'沙箱编排',
  execution_monitoring:'执行监控',loop_detection:'循环检测',progress_evaluation:'进展评估',mentor_guidance:'导师引导',
  failure_analysis:'失败分析',guidance_generation:'引导生成',error_recovery:'错误恢复',tool_suggestion:'工具建议',
  decision_review:'决策复核',evidence_verification:'证据验证',compliance_check:'合规检查',risk_reassessment:'风险重评',
  memory_store:'记忆存储',memory_retrieve:'记忆检索',pattern_extraction:'模式提取',experience_replay:'经验回放',
};

interface Props { agents: AgentInfo[]; agentNames: Record<string, string>; }

export function AgentConfigPanel({ agents, agentNames }: Props) {
  return (
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <div style={{ background: '#161b22', borderRadius: 8, border: '1px solid #21262d', padding: 24, marginBottom: 16 }}>
        <h2 style={{ margin: '0 0 8px', fontSize: 18, color: '#c9d1d9' }}>🤖 Agent 配置</h2>
        <p style={{ margin: '0 0 20px', fontSize: 12, color: '#8b949e' }}>共 {agents.length} 个 Agent，分为通用级(100次调用)和受限级(20次调用)</p>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: 12 }}>
          {agents.map(a => (
            <div key={a.type} style={{
              background: '#0d1117', borderRadius: 8, border: `1px solid ${agentColors[a.type]}22`,
              padding: 16, borderLeft: `3px solid ${agentColors[a.type] || '#888'}`,
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <div>
                  <span style={{ fontWeight: 700, fontSize: 15, color: agentColors[a.type] }}>{agentNames[a.type] || a.type}</span>
                  <span style={{ fontSize: 11, color: '#484f58', marginLeft: 8 }}>{a.type}</span>
                </div>
                <span style={{ fontSize: 10, padding: '2px 8px', borderRadius: 8, background: '#21262d', color: '#8b949e' }}>
                  {['perceiver','analyst','responder','operator','executor'].includes(a.type) ? '通用级' : '受限级'}
                </span>
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                {a.capabilities.map(cap => (
                  <span key={cap} style={{ fontSize: 11, padding: '2px 7px', borderRadius: 4, background: '#21262d', color: '#8b949e' }}>
                    {capNames[cap] || cap}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
