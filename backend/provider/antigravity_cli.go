package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const cliTimeout = 120 * time.Second

type commandExecutor interface {
	Output() ([]byte, error)
}

type AntigravityCLIProvider struct {
	command    string
	args       string
	lookPath   func(string) (string, error)
	newCommand func(context.Context, string, ...string) commandExecutor
	now        func() time.Time
	timeout    time.Duration
}

func NewAntigravityCLIProvider(command string, args string) *AntigravityCLIProvider {
	return &AntigravityCLIProvider{
		command:  strings.TrimSpace(command),
		args:     args,
		lookPath: exec.LookPath,
		newCommand: func(ctx context.Context, name string, args ...string) commandExecutor {
			return exec.CommandContext(ctx, name, args...)
		},
		now:     time.Now,
		timeout: cliTimeout,
	}
}

func (p *AntigravityCLIProvider) Name() string {
	return "antigravity-cli"
}

func hasModelFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == "--model" || strings.HasPrefix(args[i], "--model=") {
			return true
		}
	}
	return false
}

func parseShellArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && !inSingle:
			escaped = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t' || r == '\n') && !inSingle && !inDouble:
			flush()
		default:
			current.WriteRune(r)
		}
	}

	if escaped {
		return nil, errors.New("unterminated escape in args")
	}
	if inSingle || inDouble {
		return nil, errors.New("unterminated quote in args")
	}
	flush()

	return args, nil
}

func (p *AntigravityCLIProvider) resolveCommand(command string) (string, error) {
	if strings.TrimSpace(command) != "" {
		bin, err := p.lookPath(command)
		if err != nil {
			return "", fmt.Errorf("command not found: %s", command)
		}
		return bin, nil
	}

	if bin, err := p.lookPath("agy"); err == nil {
		return bin, nil
	}
	if bin, err := p.lookPath("antigravity-cli"); err == nil {
		return bin, nil
	}

	return "", errors.New("neither 'agy' nor 'antigravity-cli' found in PATH")
}

func (p *AntigravityCLIProvider) run(ctx context.Context, prompt string, cfg GenerateConfig) (GenerateResult, string, error) {
	model := strings.TrimSpace(cfg.Model)
	argText := cfg.Args
	if argText == "" {
		argText = p.args
	}
	args, err := parseShellArgs(argText)
	if err != nil {
		return GenerateResult{}, "", err
	}
	if model != "" && !hasModelFlag(args) {
		args = append(args, "--model", model)
	}

	command := strings.TrimSpace(cfg.Command)
	if command == "" {
		command = p.command
	}
	bin, err := p.resolveCommand(command)
	if err != nil {
		return GenerateResult{}, "", err
	}

	timeout := p.timeout
	if timeout <= 0 {
		timeout = cliTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := p.newCommand(runCtx, bin, args...)
	execCmd, ok := cmd.(*exec.Cmd)
	if !ok {
		return GenerateResult{}, "", errors.New("invalid command executor")
	}
	execCmd.Stdin = strings.NewReader(prompt)
	var stderr bytes.Buffer
	execCmd.Stderr = &stderr

	stdout, err := execCmd.Output()
	stderrText := strings.TrimSpace(stderr.String())
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return GenerateResult{}, stderrText, fmt.Errorf("antigravity-cli timed out after %s", timeout)
		}
		if stderrText != "" {
			return GenerateResult{}, stderrText, fmt.Errorf("antigravity-cli failed: %s", stderrText)
		}
		return GenerateResult{}, stderrText, fmt.Errorf("antigravity-cli failed: %w", err)
	}

	content := strings.TrimSpace(string(stdout))
	if content == "" {
		return GenerateResult{}, stderrText, errors.New("antigravity-cli returned empty stdout")
	}

	result := GenerateResult{
		Content:    content,
		TokenUsage: estimateTokens(len(prompt), len(content)),
	}

	return result, stderrText, nil
}

func (p *AntigravityCLIProvider) GenerateContent(ctx context.Context, prompt string, cfg GenerateConfig) (GenerateResult, error) {
	result, _, err := p.run(ctx, prompt, cfg)
	return result, err
}

func (p *AntigravityCLIProvider) StreamContent(ctx context.Context, prompt string, cfg GenerateConfig) (<-chan string, <-chan error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)
		result, err := p.GenerateContent(ctx, prompt, cfg)
		if err != nil {
			errCh <- err
			return
		}
		ch <- result.Content
	}()

	return ch, errCh
}

func (p *AntigravityCLIProvider) TestConnection(ctx context.Context, _ string, model string) error {
	testCfg := GenerateConfig{
		Model: model,
	}
	_, _, err := p.run(ctx, "ping", testCfg)
	return err
}
