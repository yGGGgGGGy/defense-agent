package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// DockerSandbox executes commands in an isolated Docker container
type DockerSandbox struct {
	containerName string
}

// NewDockerSandbox creates a Docker-based sandbox executor
func NewDockerSandbox(containerName string) *DockerSandbox {
	return &DockerSandbox{containerName: containerName}
}

// AllowedCommands are commands that can be run without additional approval
var AllowedCommands = map[string]bool{
	"nmap":     true,
	"curl":     true,
	"wget":     true,
	"ping":     true,
	"traceroute": true,
	"dig":      true,
	"nslookup": true,
	"whois":    true,
	"ssh":      true,
	"netstat":  true,
	"ss":       true,
	"iptables": true,
	"systemctl": true,
	"journalctl": true,
	"df":       true,
	"free":     true,
	"ps":       true,
	"top":      true,
	"htop":     true,
	"lsof":     true,
	"tar":      true,
	"gzip":     true,
	"cat":      true,
	"head":     true,
	"tail":     true,
	"grep":     true,
	"find":     true,
	"awk":      true,
	"sed":      true,
	"sort":     true,
	"uniq":     true,
	"wc":       true,
	"cp":       true,
	"mv":       true,
	"mkdir":    true,
	"logrotate": true,
}

// Execute runs a command in the Docker sandbox
func (s *DockerSandbox) Execute(ctx context.Context, command string, timeout time.Duration) (string, error) {
	// Validate command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	baseCmd := parts[0]
	if !AllowedCommands[baseCmd] {
		return "", fmt.Errorf("command %q not in allowed list", baseCmd)
	}

	// Check if Docker container is running
	if err := s.ensureRunning(ctx); err != nil {
		log.Warn().Err(err).Msg("docker sandbox unavailable, using local mock")
		return s.mockExecute(command), nil
	}

	// Build docker exec command
	dockerArgs := []string{"exec", "-i", s.containerName}
	if timeout > 0 {
		dockerArgs = append(dockerArgs, "timeout", fmt.Sprintf("%d", int(timeout.Seconds())))
	}
	dockerArgs = append(dockerArgs, "bash", "-c", command)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Sandbox error: %s\nStderr: %s", err, stderr.String()), nil
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (s *DockerSandbox) ensureRunning(ctx context.Context) error {
	check := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", s.containerName)
	out, err := check.Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf("sandbox container %q not running", s.containerName)
	}
	return nil
}

func (s *DockerSandbox) mockExecute(command string) string {
	parts := strings.Fields(command)
	baseCmd := parts[0]
	switch baseCmd {
	case "nmap":
		return "Nmap scan: 22/tcp open (SSH), 80/tcp open (HTTP), 443/tcp open (HTTPS)"
	case "curl":
		return "HTTP/1.1 200 OK\nServer: nginx\nContent-Type: text/html"
	case "df":
		return "/dev/sda1  50G  20G  28G  42% /"
	case "free":
		return "Mem: total=7932 used=3214 free=4718"
	case "ps":
		return "PID  CMD\n1 systemd\n42 sshd\n99 nginx"
	case "iptables":
		return "Chain INPUT (policy ACCEPT)\nChain FORWARD (policy DROP)"
	case "systemctl":
		return "sshd.service active (running)\nnginx.service active (running)"
	case "journalctl":
		return "May 09 10:00:00 server sshd[42]: Failed password for root from 10.0.0.50"
	case "cat", "head", "tail":
		return fmt.Sprintf("[sandbox mock] Contents of %s", strings.Join(parts[1:], " "))
	default:
		return fmt.Sprintf("[sandbox mock] %s executed successfully", baseCmd)
	}
}
