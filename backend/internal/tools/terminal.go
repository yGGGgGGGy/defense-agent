package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

// safeCommands are commands allowed without sandbox
var safeCommands = map[string]bool{
	"echo":   true,
	"date":   true,
	"whoami": true,
	"hostname": true,
	"uptime": true,
	"uname":  true,
	"pwd":    true,
	"env":    true,
}

func terminalTool() *Tool {
	return &Tool{
		Name:        "terminal",
		Description: "Execute a shell command in the sandbox environment for security testing",
		Parameters: []Param{
			{Name: "command", Type: "string", Description: "Shell command to execute", Required: true},
			{Name: "timeout", Type: "integer", Description: "Timeout in seconds (default 30)", Required: false},
		},
		Sandbox:   true,
		RiskLevel: "medium",
		Handler: func(args map[string]string) (string, error) {
			cmd := args["command"]
			if cmd == "" {
				return "", fmt.Errorf("command is required")
			}

			// Validate safety
			parts := strings.Fields(cmd)
			if len(parts) == 0 {
				return "", fmt.Errorf("empty command")
			}
			baseCmd := parts[0]

			if !safeCommands[baseCmd] && !strings.HasPrefix(baseCmd, "nmap") &&
				!strings.HasPrefix(baseCmd, "curl") && !strings.HasPrefix(baseCmd, "grep") {
				return fmt.Sprintf("Command '%s' requires sandbox. Use sandbox_exec tool instead.", baseCmd), nil
			}

			ctx := exec.Command("timeout", "30", "bash", "-c", cmd)
			output, err := ctx.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Command failed: %s\nOutput: %s", err, string(output)), nil
			}
			return strings.TrimSpace(string(output)), nil
		},
	}
}
