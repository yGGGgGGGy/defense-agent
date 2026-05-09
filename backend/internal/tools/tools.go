package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func healthCheckTool() *Tool {
	return &Tool{
		Name:        "health_check",
		Description: "Check system health: disk usage, memory, running services",
		RiskLevel:   "low",
		Sandbox:     true,
		Handler: func(args map[string]string) (string, error) {
			var results []string

			// Disk usage
			if out, err := execCmd("df", "-h", "/"); err == nil {
				results = append(results, "Disk:\n"+out)
			}

			// Memory
			if out, err := execCmd("free", "-m"); err == nil {
				results = append(results, "Memory:\n"+out)
			}

			return strings.Join(results, "\n\n"), nil
		},
	}
}

func fileReadTool() *Tool {
	return &Tool{
		Name:        "read_file",
		Description: "Read the contents of a file",
		Parameters: []Param{
			{Name: "path", Type: "string", Description: "File path to read", Required: true},
		},
		RiskLevel: "low",
		Sandbox:   false,
		Handler: func(args map[string]string) (string, error) {
			path := args["path"]
			if path == "" {
				return "", fmt.Errorf("path is required")
			}
			// Only allow reading from safe directories
			if strings.Contains(path, "..") || strings.HasPrefix(path, "/etc/shadow") {
				return "", fmt.Errorf("access denied: %s", path)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read error: %w", err)
			}
			if len(data) > 4096 {
				data = data[:4096]
			}
			return string(data), nil
		},
	}
}

func execCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
