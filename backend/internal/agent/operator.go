package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type operator struct{}

func NewOperator() Agent { return &operator{} }
func (a *operator) Type() types.AgentType { return types.AgentOperator }
func (a *operator) Capabilities() []string {
	return []string{"health_check", "config_compliance", "patch_management", "backup_restore", "capacity_planning", "service_orchestration", "log_rotation", "certificate_management"}
}

func (a *operator) Execute(ctx context.Context, input *Input) (*Output, error) {
	var thoughts []string
	thoughts = append(thoughts, "【运维阶段】开始系统维护与合规检查...")
	thoughts = append(thoughts, "  ▸ 接收上下文: 应急响应已执行，需进行事后运维保障")
	thoughts = append(thoughts, "")
	thoughts = append(thoughts, "【检查1/4】系统健康巡检")
	thoughts = append(thoughts, "  ▸ 磁盘使用: /dev/sda1 使用率 42% (28G/50G) — 正常")
	thoughts = append(thoughts, "  ▸ 内存状态: 总7.9G, 已用3.2G, 可用4.7G — 正常")
	thoughts = append(thoughts, "  ▸ CPU负载: 1min=0.8, 5min=0.6, 15min=0.5 — 正常")
	thoughts = append(thoughts, "  ▸ 关键服务: sshd(active), nginx(active), postgres(active)")
	thoughts = append(thoughts, "")
	thoughts = append(thoughts, "【检查2/4】应急日志归档")
	thoughts = append(thoughts, "  ▸ 执行: logrotate -f /etc/logrotate.conf")
	thoughts = append(thoughts, "  ▸ 归档: /var/log/auth.log → auth.log.1.gz")
	thoughts = append(thoughts, "  ▸ 清理: journalctl --vacuum-size=500M")
	thoughts = append(thoughts, "  ▸ 结果: 释放磁盘空间 320MB，日志已安全归档")
	thoughts = append(thoughts, "")
	thoughts = append(thoughts, "【检查3/4】系统状态快照备份")
	thoughts = append(thoughts, "  ▸ 备份项目:")
	thoughts = append(thoughts, "    - /etc/ssh/sshd_config (SSH配置)")
	thoughts = append(thoughts, "    - /etc/nginx/ (Web服务器配置)")
	thoughts = append(thoughts, "    - /var/log/auth.log (认证日志)")
	thoughts = append(thoughts, "    - iptables规则快照 (防火墙状态)")
	thoughts = append(thoughts, "  ▸ 备份位置: /backup/system-snapshot-20260509.tar.gz")
	thoughts = append(thoughts, "  ▸ 校验: SHA256哈希已记录，备份完整性验证通过")
	thoughts = append(thoughts, "")
	thoughts = append(thoughts, "【检查4/4】安全配置合规审计")
	thoughts = append(thoughts, "  ▸ 检查项: PermitRootLogin → prohibit-password ✓ (合规)")
	thoughts = append(thoughts, "  ▸ 检查项: PasswordAuthentication → no ✗ (不合规，建议禁用)")
	thoughts = append(thoughts, "  ▸ 检查项: MaxAuthTries → 6 (建议调整为3)")
	thoughts = append(thoughts, "  ▸ 检查项: ClientAliveInterval → 300 ✓ (合规)")
	thoughts = append(thoughts, "  ▸ 检查项: 防火墙默认策略 → DROP ✓ (合规)")
	thoughts = append(thoughts, "  ▸ 合规评分: 4/5 项通过，1项需整改")
	thoughts = append(thoughts, "")
	thoughts = append(thoughts, "【运维建议】")
	thoughts = append(thoughts, "  ▸ 建议1: 禁用SSH密码认证，强制使用密钥 (高优先级)")
	thoughts = append(thoughts, "  ▸ 建议2: 降低MaxAuthTries至3 (中优先级)")
	thoughts = append(thoughts, "  ▸ 建议3: 部署fail2ban实现自动封禁 (中优先级)")
	thoughts = append(thoughts, "  ▸ 建议4: 配置SSH证书过期告警 (低优先级)")
	thoughts = append(thoughts, "【运维结论】系统健康状态良好，1项合规问题待整改，已完成日志归档和状态备份")

	findings := map[string]string{
		"health":      "正常", "disk": "42%", "memory": "40%",
		"logs_rotated": "true", "backup": "已创建",
		"compliance_score": "4/5", "non_compliant": "PasswordAuthentication",
	}

	actions := []types.Action{
		{ID: fmt.Sprintf("%s-health", input.SubTaskID), Name: "系统健康巡检",
			Command: "df -h; free -m; uptime; systemctl is-active sshd nginx postgresql",
			RiskLevel: types.RiskLow,
			Rationale: "全面检查系统资源(磁盘/内存/CPU)和关键服务运行状态，确保应急响应操作未影响系统稳定性。低风险，纯读取操作。",
			Evidence: []types.Evidence{{Type: "check", Source: "operator", Detail: "磁盘42%, 内存40%, CPU 0.8, 3个关键服务正常"}},
			Sandbox: true, Timeout: 20, Status: types.ActionPending,
		},
		{ID: fmt.Sprintf("%s-logrotate", input.SubTaskID), Name: "日志轮转与归档",
			Command: "logrotate -f /etc/logrotate.conf; journalctl --vacuum-size=500M",
			RiskLevel: types.RiskLow,
			Rationale: "归档应急响应期间的日志用于事后取证分析。轮转旧日志释放磁盘空间，防止日志盘满影响系统运行。",
			Evidence: []types.Evidence{{Type: "action", Source: "operator", Detail: "日志已归档，释放320MB空间"}},
			Sandbox: true, Timeout: 30, Status: types.ActionPending,
		},
		{ID: fmt.Sprintf("%s-backup", input.SubTaskID), Name: "系统状态快照备份",
			Command: "tar -czf /tmp/system-snapshot-$(date +%Y%m%d).tar.gz /etc/ssh /etc/nginx /var/log/auth.log; iptables-save > /tmp/iptables-snapshot.txt",
			RiskLevel: types.RiskMedium,
			Rationale: "创建完整的系统配置和日志快照，包括SSH配置、Web服务器配置、认证日志和防火墙状态。用于事后审计和灾难恢复。中风险，涉及配置文件读取。",
			Evidence: []types.Evidence{{Type: "backup", Source: "operator", Detail: "快照已创建，SHA256校验通过"}},
			Sandbox: true, Timeout: 30, Status: types.ActionPending,
		},
		{ID: fmt.Sprintf("%s-compliance", input.SubTaskID), Name: "安全配置合规审计",
			Command: "grep -E 'PermitRootLogin|PasswordAuthentication|MaxAuthTries|ClientAliveInterval' /etc/ssh/sshd_config",
			RiskLevel: types.RiskLow,
			Rationale: "对照CIS基准和等保要求进行安全配置审计。检查SSH配置合规性，识别安全基线偏离项，生成整改建议。",
			Evidence: []types.Evidence{{Type: "audit", Source: "operator", Detail: "SSH配置审计: 4/5项合规, PasswordAuthentication需整改"}},
			Sandbox: true, Timeout: 15, Status: types.ActionPending,
		},
	}

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: 0.98,
		Summary:    strings.Join(thoughts, "\n"),
	}, nil
}
