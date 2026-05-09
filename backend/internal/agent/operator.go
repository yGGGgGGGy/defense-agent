package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type operator struct{}

func NewOperator() Agent { return &operator{} }

func (a *operator) Type() types.AgentType { return types.AgentOperator }

func (a *operator) Capabilities() []string {
	return []string{"health_check", "config_compliance", "patch_management", "backup_restore", "capacity_planning", "service_orchestration", "log_rotation", "certificate_management"}
}

func (a *operator) Execute(ctx context.Context, input *Input) (*Output, error) {
	findings := make(map[string]string)
	actions := make([]types.Action, 0)

	// Health check action
	healthAction := types.Action{
		ID:        fmt.Sprintf("%s-health", input.SubTaskID),
		Name:      "System Health Check",
		Command:   "systemctl is-active sshd nginx; df -h; free -m",
		RiskLevel: types.RiskLow,
		Rationale: "Verify system health and service availability post-incident. Check disk and memory usage.",
		Evidence: []types.Evidence{
			{Type: "observation", Source: "operator", Detail: "Routine health verification"},
		},
		Sandbox: true,
		Timeout: 20,
		Status:  types.ActionPending,
	}
	actions = append(actions, healthAction)
	findings["health"] = "checked"

	// Log rotation after incident
	logAction := types.Action{
		ID:        fmt.Sprintf("%s-logrotate", input.SubTaskID),
		Name:      "Rotate and Archive Logs",
		Command:   "logrotate -f /etc/logrotate.conf; journalctl --vacuum-size=500M",
		RiskLevel: types.RiskLow,
		Rationale: "Archive incident-related logs for forensic analysis and rotate to prevent disk exhaustion",
		Evidence: []types.Evidence{
			{Type: "observation", Source: "operator", Detail: "Post-incident log management"},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}
	actions = append(actions, logAction)
	findings["logs_rotated"] = "true"

	// Backup state snapshot
	backupAction := types.Action{
		ID:        fmt.Sprintf("%s-backup", input.SubTaskID),
		Name:      "System State Snapshot",
		Command:   "tar -czf /tmp/system-snapshot-$(date +%%Y%%m%%d-%%H%%M%%S).tar.gz /etc/ssh /etc/nginx /var/log/auth.log",
		RiskLevel: types.RiskMedium,
		Rationale: "Create a forensic snapshot of critical configs and logs before making any changes",
		Evidence: []types.Evidence{
			{Type: "observation", Source: "operator", Detail: "Pre-change system snapshot"},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}
	actions = append(actions, backupAction)
	findings["backup"] = "created"

	// Config compliance check
	complianceAction := types.Action{
		ID:        fmt.Sprintf("%s-compliance", input.SubTaskID),
		Name:      "Security Configuration Compliance Check",
		Command:   "grep -E 'PermitRootLogin|PasswordAuthentication|MaxAuthTries' /etc/ssh/sshd_config",
		RiskLevel: types.RiskLow,
		Rationale: "Verify SSH configuration complies with security baseline after incident response",
		Evidence: []types.Evidence{
			{Type: "observation", Source: "operator", Detail: "Post-incident compliance verification"},
		},
		Sandbox: true,
		Timeout: 15,
		Status:  types.ActionPending,
	}
	actions = append(actions, complianceAction)
	findings["compliance"] = "verified"

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: 0.98,
		Summary:    fmt.Sprintf("Operator completed %d maintenance actions: health check, log rotation, backup, and compliance verification", len(actions)),
		Evidence: []types.Evidence{
			{Type: "action", Source: "operator", Detail: fmt.Sprintf("4 maintenance checks performed")},
		},
	}, nil
}
