package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseShellArgs(t *testing.T) {
	args, err := parseShellArgs(`translate --to vi --title "abc def" --note 'x y'`)
	if err != nil {
		t.Fatalf("parseShellArgs() error = %v", err)
	}

	want := []string{"translate", "--to", "vi", "--title", "abc def", "--note", "x y"}
	if len(args) != len(want) {
		t.Fatalf("len(args) = %d, want %d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestResolveCommandFallback(t *testing.T) {
	p := NewAntigravityCLIProvider("", "")
	p.lookPath = func(name string) (string, error) {
		switch name {
		case "agy":
			return "", errors.New("missing")
		case "antigravity-cli":
			return "/usr/bin/antigravity-cli", nil
		default:
			return "", errors.New("missing")
		}
	}

	bin, err := p.resolveCommand("")
	if err != nil {
		t.Fatalf("resolveCommand() error = %v", err)
	}
	if bin != "/usr/bin/antigravity-cli" {
		t.Fatalf("resolveCommand() = %q, want %q", bin, "/usr/bin/antigravity-cli")
	}
}

func TestGenerateContentSuccess(t *testing.T) {
	p := NewAntigravityCLIProvider("sh", "-c 'cat'")
	result, err := p.GenerateContent(context.Background(), "xin chao", GenerateConfig{})
	if err != nil {
		t.Fatalf("GenerateContent() error = %v", err)
	}
	if result.Content != "xin chao" {
		t.Fatalf("Content = %q, want %q", result.Content, "xin chao")
	}
}

func TestGenerateContentEmptyStdoutIsError(t *testing.T) {
	p := NewAntigravityCLIProvider("sh", "-c 'true'")
	_, err := p.GenerateContent(context.Background(), "ignored", GenerateConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "empty stdout") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "empty stdout")
	}
}

func TestGenerateContentNonZeroExitIsError(t *testing.T) {
	p := NewAntigravityCLIProvider("sh", "-c 'echo boom >&2; exit 7'")
	_, err := p.GenerateContent(context.Background(), "ignored", GenerateConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "boom")
	}
}

func TestGenerateContentTimeout(t *testing.T) {
	p := NewAntigravityCLIProvider("sh", "-c 'sleep 2; cat'")
	p.timeout = 100 * time.Millisecond

	_, err := p.GenerateContent(context.Background(), "ignored", GenerateConfig{})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "timed out")
	}
}
