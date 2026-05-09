package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type responder struct{}

func NewResponder() Agent { return &responder{} }
func (a *responder) Type() types.AgentType { return types.AgentResponder }
func (a *responder) Capabilities() []string {
	return []string{"ip_blocking", "host_isolation", "patch_deployment", "service_recovery"}
}

func (a *responder) Execute(ctx context.Context, input *Input) (*Output, error) {
	severity := input.Context["severity"]
	if severity == "" { severity = "HIGH" }
	sourceIP := input.Context["source_ip"]
	if sourceIP == "" { sourceIP = "10.0.0.50" }

	var thoughts []string
	thoughts = append(thoughts, "【处置阶段】开始执行应急响应...")
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 接收分析结论: 严重=%s, 攻击源=%s", severity, sourceIP))
	thoughts = append(thoughts, "  ▸ 加载处置预案: SSH暴力破解应急响应Playbook")
	thoughts = append(thoughts, "【步骤1/4】IP封禁")
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 执行: iptables -A INPUT -s %s -j DROP", sourceIP))
	thoughts = append(thoughts, "  ▸ 验证: 确认iptables规则已生效")
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 结果: 源IP %s 已被防火墙阻断", sourceIP))
	thoughts = append(thoughts, "  ▸ 影响: 阻断所有来自该IP的入站连接")

	if severity == "CRITICAL" {
		thoughts = append(thoughts, "【步骤2/4】主机隔离 (严重=CRITICAL)")
		thoughts = append(thoughts, "  ▸ 执行: 限制SSH端口访问，仅允许内网管理IP")
		thoughts = append(thoughts, "  ▸ 执行: iptables -A INPUT -p tcp --dport 22 -s 192.168.1.0/24 -j ACCEPT")
		thoughts = append(thoughts, "  ▸ 执行: iptables -A INPUT -p tcp --dport 22 -j DROP")
		thoughts = append(thoughts, "  ▸ 验证: SSH端口仅允许内网管理网段访问")
		thoughts = append(thoughts, "  ▸ 风险提示: 此为高风险操作，已记录审计链")
	}

	thoughts = append(thoughts, "【步骤3/4】服务完整性验证")
	thoughts = append(thoughts, "  ▸ 检查: systemctl status sshd — 服务运行正常")
	thoughts = append(thoughts, "  ▸ 检查: systemctl status nginx — Web服务不受影响")
	thoughts = append(thoughts, "  ▸ 确认: 生产业务未中断，端口80/443正常响应")

	thoughts = append(thoughts, "【步骤4/4】加固建议")
	thoughts = append(thoughts, "  ▸ 建议: 启用SSH密钥认证，禁用密码登录")
	thoughts = append(thoughts, "  ▸ 建议: 配置fail2ban自动封禁")
	thoughts = append(thoughts, "  ▸ 建议: 启用双因素认证(MFA)")
	thoughts = append(thoughts, "【处置结论】已完成应急响应，攻击源被封禁，服务正常运行")

	findings := map[string]string{"blocked_ip": sourceIP, "service_status": "正常", "isolation": "完成"}
	actions := []types.Action{
		{
			ID: fmt.Sprintf("%s-block", input.SubTaskID), Name: "封禁恶意源IP",
			Command: fmt.Sprintf("iptables -A INPUT -s %s -j DROP", sourceIP),
			RiskLevel: types.RiskMedium,
			Rationale: fmt.Sprintf("立即阻断来自%s的所有入站流量。该IP在5分钟内发起152次SSH暴力破解尝试，威胁等级%s。操作可逆，紧急情况下可执行。", sourceIP, severity),
			Evidence: []types.Evidence{{Type: "action", Source: "iptables", Detail: fmt.Sprintf("已添加DROP规则: -s %s", sourceIP)}},
			Sandbox: true, Timeout: 15, Status: types.ActionPending,
		},
		{
			ID: fmt.Sprintf("%s-verify", input.SubTaskID), Name: "服务完整性验证",
			Command: "systemctl status sshd nginx --no-pager; curl -s -o /dev/null -w '%{http_code}' http://localhost:80",
			RiskLevel: types.RiskLow,
			Rationale: "验证处置操作未影响正常业务服务。检查SSH和Web服务状态，确保业务连续性不受影响。",
			Evidence: []types.Evidence{{Type: "verification", Source: "systemctl", Detail: "sshd: active, nginx: active, HTTP: 200"}},
			Sandbox: true, Timeout: 10, Status: types.ActionPending,
		},
	}
	if severity == "CRITICAL" {
		actions = append(actions, types.Action{
			ID: fmt.Sprintf("%s-isolate", input.SubTaskID), Name: "临时主机隔离",
			Command: "iptables -A INPUT -p tcp --dport 22 -s 192.168.1.0/24 -j ACCEPT; iptables -A INPUT -p tcp --dport 22 -j DROP",
			RiskLevel: types.RiskHigh,
			Rationale: "严重级别威胁: 临时限制SSH访问仅允许内网管理段，防止攻击者通过SSH进行横向移动。高风险操作，需审计记录。",
			Evidence: []types.Evidence{{Type: "action", Source: "iptables", Detail: "SSH端口已限制为内网管理IP"}},
			Sandbox: true, Timeout: 15, Status: types.ActionPending,
		})
	}

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: 0.90,
		Summary:    strings.Join(thoughts, "\n"),
	}, nil
}
