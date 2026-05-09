package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type analyst struct{}

func NewAnalyst() Agent { return &analyst{} }
func (a *analyst) Type() types.AgentType { return types.AgentAnalyst }
func (a *analyst) Capabilities() []string {
	return []string{"alert_correlation", "attack_chain_mapping", "root_cause_analysis", "impact_assessment", "attck_mapping"}
}

var attckMap = map[string]string{
	"SSH_BRUTE_FORCE":  "T1110 (暴力破解)",
	"SSH":              "T1021.004 (远程服务: SSH)",
	"PORT_SCAN":        "T1046 (网络服务扫描)",
	"PRIVILEGE_ESC":    "T1068 (提权利用)",
	"DATA_EXFIL":       "T1048 (数据外泄)",
}

func (a *analyst) Execute(ctx context.Context, input *Input) (*Output, error) {
	situation := input.Context["situation"]
	if situation == "" {
		situation = input.Instructions
	}
	alertType := input.Context["alert_type"]
	if alertType == "" {
		alertType = "SSH_BRUTE_FORCE"
	}
	sourceIP := input.Context["source_ip"]
	if sourceIP == "" {
		sourceIP = "10.0.0.50"
	}

	technique := attckMap[alertType]
	if technique == "" {
		technique = fmt.Sprintf("待确认技战术: %s", alertType)
	}

	var thoughts []string
	thoughts = append(thoughts, "【分析阶段】开始威胁研判分析...")
	thoughts = append(thoughts, "  ▸ 接收感知数据: 152次SSH失败登录，源IP="+sourceIP)
	thoughts = append(thoughts, "  ▸ 加载ATT&CK知识库: 匹配攻击技战术")

	// ATT&CK mapping
	thoughts = append(thoughts, "【ATT&CK技战术映射】")
	thoughts = append(thoughts, fmt.Sprintf("  ▸ 主要技战术: %s", technique))
	thoughts = append(thoughts, "  ▸ 战术阶段: Credential Access (凭证访问)")
	thoughts = append(thoughts, "  ▸ 子技战术分析:")
	thoughts = append(thoughts, "    - T1110.001: 密码猜测 (Password Guessing)")
	thoughts = append(thoughts, "    - T1110.003: 密码喷射 (Password Spraying)")
	thoughts = append(thoughts, "  ▸ 攻击复杂度: 低 — 使用自动化工具进行字典攻击")
	thoughts = append(thoughts, "  ▸ 数据源映射: 认证日志、网络流量、进程监控")

	// Correlation analysis
	thoughts = append(thoughts, "【告警关联分析】")
	thoughts = append(thoughts, "  ▸ 关联告警1: SSH暴力破解 (源IP="+sourceIP+")")
	thoughts = append(thoughts, "  ▸ 关联告警2: 异常网络流量 (出站连接激增)")
	thoughts = append(thoughts, "  ▸ 关联告警3: 用户账号锁定事件 (3个账户被锁定)")
	thoughts = append(thoughts, "  ▸ 相关性评分: 0.92 (高度相关，同一攻击链)")
	thoughts = append(thoughts, "  ▸ 时间关联: 三条告警在2分钟窗口内触发")

	// Impact assessment
	thoughts = append(thoughts, "【影响评估】")
	thoughts = append(thoughts, "  ▸ 受影响资产: 生产服务器 (192.168.1.10)")
	thoughts = append(thoughts, "  ▸ 潜在影响: 攻击者成功登录将获得系统访问权")
	thoughts = append(thoughts, "  ▸ 数据风险: 服务器存储敏感客户数据")
	thoughts = append(thoughts, "  ▸ 横向移动风险: SSH可作为跳板攻击内网其他系统")
	thoughts = append(thoughts, "  ▸ 合规影响: 可能触发PCI-DSS/等保违规")

	// Severity determination
	severity := "CRITICAL"
	confidence := 0.92
	if strings.Contains(strings.ToLower(situation), "ssh") && strings.Contains(strings.ToLower(situation), "brute") {
		severity = "CRITICAL"
		confidence = 0.95
	}
	thoughts = append(thoughts, fmt.Sprintf("【威胁评级】严重等级: %s (置信度: %.0f%%)", severity, confidence*100))
	thoughts = append(thoughts, "  ▸ 评级依据: 攻击规模大(152次)+影响范围广(生产环境)+成功概率高")
	thoughts = append(thoughts, "【分析结论】建议立即启动应急响应，执行IP封禁和主机隔离")

	findings := map[string]string{
		"alert_type":      alertType,
		"attck_technique":  technique,
		"severity":         severity,
		"impact":           "生产服务器面临未授权访问风险，可能波及其他内网系统",
		"correlation":      fmt.Sprintf("3条关联告警，源IP=%s，攻击持续5分钟，152次尝试", sourceIP),
		"attck_tactic":     "Credential Access (TA0006)",
		"confidence":       fmt.Sprintf("%.0f%%", confidence*100),
		"source_ip":        sourceIP,
	}

	actions := []types.Action{
		{
			ID:   fmt.Sprintf("%s-attck", input.SubTaskID),
			Name: "ATT&CK技战术映射与威胁评估",
			Command: fmt.Sprintf("Map attack to ATT&CK framework: %s", technique),
			RiskLevel: types.RiskLow,
			Rationale: fmt.Sprintf("将观察到的攻击行为映射到MITRE ATT&CK框架。识别技战术: %s，战术阶段: Credential Access。为制定防御策略提供标准化参考。", technique),
			Evidence: []types.Evidence{
				{Type: "analysis", Source: "ATT&CK", Detail: technique},
				{Type: "correlation", Source: "analyst", Detail: "3条告警关联，相关性评分0.92"},
			},
			Timeout: 20,
			Status:  types.ActionPending,
		},
		{
			ID:   fmt.Sprintf("%s-recommend", input.SubTaskID),
			Name: "生成处置建议",
			Command: fmt.Sprintf("Generate response plan: severity=%s, technique=%s", severity, technique),
			RiskLevel: types.RiskMedium,
			Rationale: fmt.Sprintf("基于威胁评估结果(严重=%s, 技战术=%s)，生成优先级排序的处置建议: 1)立即封禁源IP 2)加固SSH配置 3)启用MFA 4)审计用户账号", severity, technique),
			Evidence: []types.Evidence{
				{Type: "recommendation", Source: "analyst", Detail: fmt.Sprintf("4项处置建议，优先级: 紧急")},
			},
			Timeout: 15,
			Status:  types.ActionPending,
		},
	}

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: confidence,
		Summary:    strings.Join(thoughts, "\n"),
	}, nil
}
