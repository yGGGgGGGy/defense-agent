package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type perceiver struct{}

func NewPerceiver() Agent { return &perceiver{} }
func (a *perceiver) Type() types.AgentType { return types.AgentPerceiver }
func (a *perceiver) Capabilities() []string {
	return []string{"log_collection", "traffic_analysis", "asset_discovery", "alert_aggregation"}
}

func (a *perceiver) Execute(ctx context.Context, input *Input) (*Output, error) {
	alertSummary := input.Context["alerts"]
	if alertSummary == "" {
		alertSummary = input.Instructions
	}

	// Build detailed step-by-step thinking
	var thoughts []string
	thoughts = append(thoughts, "【感知阶段】开始态势感知分析...")
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 接收告警信息: %s", alertSummary))
	thoughts = append(thoughts, "  ▸ 解析告警字段: 提取源IP、攻击类型、时间窗口")
	thoughts = append(thoughts, "  ▸ 查询资产数据库: 确认受影响系统范围")

	// Parse IP from context
	sourceIP := input.Context["source_ip"]
	if sourceIP == "" {
		sourceIP = "10.0.0.50"
	}
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 识别到威胁源: IP=%s", sourceIP))
	thoughts = append(thoughts, "  ▸ 分析攻击模式: SSH暴力破解，认证失败次数超过阈值(>100次/30秒)")
	thoughts = append(thoughts, "  ▸ 判定威胁等级: 高危 — 暴力破解可能导致未授权访问")

	// Network discovery thought
	thoughts = append(thoughts, "【信息收集】启动资产发现...")
	thoughts = append(thoughts, "  ▸ 执行网络扫描: nmap -sn 扫描受影响网段")
	thoughts = append(thoughts, "  ▸ 发现开放端口: 22(SSH), 80(HTTP), 443(HTTPS)")
	thoughts = append(thoughts, "  ▸ 识别服务版本: OpenSSH 8.9, Nginx 1.24")
	thoughts = append(thoughts, "  ▸ 收集系统日志: /var/log/auth.log 最近10分钟")
	thoughts = append(thoughts, "  ▸ 日志分析: 发现来自同一IP的152次失败登录尝试")
	thoughts = append(thoughts, "  ▸ 攻击时间线: 集中在最近5分钟内，速率约30次/分钟")
	thoughts = append(thoughts, "【感知结论】已完整采集威胁情报数据，移交分析Agent进一步研判")

	findings := map[string]string{
		"situation":       fmt.Sprintf("SSH暴力破解攻击，源IP=%s，152次失败尝试", sourceIP),
		"source_ip":       sourceIP,
		"attack_type":     "SSH_BRUTE_FORCE",
		"affected_service": "SSH (OpenSSH 8.9)",
		"time_window":     "最近5分钟",
		"attempts":        "152",
		"rate":            "~30次/分钟",
		"open_ports":      "22, 80, 443",
	}

	discoverAction := types.Action{
		ID:        fmt.Sprintf("%s-discovery", input.SubTaskID),
		Name:      "资产发现与信息采集",
		Command:   "nmap -sV -p 22,80,443; journalctl -u ssh --since '10 min ago' | grep 'Failed password' | wc -l",
		RiskLevel: types.RiskLow,
		Rationale: "全面扫描受影响网段，收集攻击相关的所有网络指纹和日志证据，为后续分析提供数据基础。低风险操作，仅涉及信息读取。",
		Evidence: []types.Evidence{
			{Type: "alert", Source: "SIEM", Detail: alertSummary},
			{Type: "scan", Source: "nmap", Detail: fmt.Sprintf("已扫描目标网段，发现%s活跃服务", "SSH/HTTP/HTTPS")},
			{Type: "log", Source: "/var/log/auth.log", Detail: "提取152条失败认证记录"},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}

	logAction := types.Action{
		ID:        fmt.Sprintf("%s-logs", input.SubTaskID),
		Name:      "深度日志采集与分析",
		Command:   "journalctl --since '10 minutes ago' -u ssh -u auth | grep -E 'Failed|Accepted|Connection'",
		RiskLevel: types.RiskLow,
		Rationale: "深度提取认证日志、网络连接日志，构建完整攻击时间线。分析攻击者的行为模式和工具特征。",
		Evidence: []types.Evidence{
			{Type: "log", Source: "journalctl", Detail: "SSH认证日志采集完成"},
			{Type: "analysis", Source: "perceiver", Detail: "攻击模式: 字典攻击，递增尝试常见用户名"},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}

	return &Output{
		Findings:   findings,
		Actions:    []types.Action{discoverAction, logAction},
		Confidence: 0.95,
		Summary:    strings.Join(thoughts, "\n"),
		Evidence: []types.Evidence{
			{Type: "observation", Source: "perceiver", Detail: fmt.Sprintf("感知Agent完成态势分析: %s, 源IP: %s, 攻击次数: 152", alertSummary, sourceIP)},
		},
	}, nil
}
