package grokcli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathPrefersExplicitExecutable(t *testing.T) {
	explicit := makeExecutable(t, filepath.Join(t.TempDir(), "custom-grok"))
	resolved, err := resolvePath(explicit, func(string) (string, error) {
		return "/path/grok", nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != explicit {
		t.Fatalf("应优先使用显式路径: got %q want %q", resolved, explicit)
	}
}

func TestResolvePathUsesPathBeforeDarwinCandidates(t *testing.T) {
	pathExecutable := makeExecutable(t, filepath.Join(t.TempDir(), "path-grok"))
	resolved, err := resolvePath("", func(name string) (string, error) {
		if name != "grok" {
			t.Fatalf("查找了错误的命令: %s", name)
		}
		return pathExecutable, nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != pathExecutable {
		t.Fatalf("没有使用 PATH 中的命令: got %q want %q", resolved, pathExecutable)
	}
}

func TestResolvePathFindsDarwinUserCandidate(t *testing.T) {
	home := t.TempDir()
	candidate := makeExecutable(t, filepath.Join(home, ".local", "bin", "grok"))
	resolved, err := resolvePath("", func(string) (string, error) {
		return "", errors.New("not found")
	}, []string{candidate})
	if err != nil {
		t.Fatal(err)
	}
	if resolved != candidate {
		t.Fatalf("没有使用 macOS 用户路径: got %q want %q", resolved, candidate)
	}
}

func TestResolvePathRejectsInvalidExplicitPathWithoutFallback(t *testing.T) {
	_, err := resolvePath(filepath.Join(t.TempDir(), "missing"), func(string) (string, error) {
		return "/path/grok", nil
	}, nil)
	if err == nil || !strings.Contains(err.Error(), pathEnv) {
		t.Fatalf("应返回显式配置错误，实际为 %v", err)
	}
}

func TestResolvePathReturnsActionableError(t *testing.T) {
	_, err := resolvePath("", func(string) (string, error) {
		return "", errors.New("not found")
	}, nil)
	if err == nil || !strings.Contains(err.Error(), pathEnv) {
		t.Fatalf("错误应说明 %s，实际为 %v", pathEnv, err)
	}
}

func makeExecutable(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
