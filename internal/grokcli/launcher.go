package grokcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const pathEnv = "GROK_CLI"

// Resolve 查找可执行的 Grok CLI，并返回绝对路径。
func Resolve() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates := []string{}
	if runtime.GOOS == "darwin" {
		candidates = darwinCandidates(home)
	}
	return resolvePath(os.Getenv(pathEnv), exec.LookPath, candidates)
}

// StartLogin 启动 Grok CLI 官方账号登录流程。
func StartLogin() error {
	path, err := Resolve()
	if err != nil {
		return err
	}
	if err := exec.Command(path, "login").Start(); err != nil {
		return fmt.Errorf("启动 %s login 失败: %w", path, err)
	}
	return nil
}

func resolvePath(explicit string, lookPath func(string) (string, error), candidates []string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		resolved, err := resolveExplicit(explicit, lookPath)
		if err != nil {
			return "", fmt.Errorf("环境变量 %s 指向的 Grok CLI 不可用: %w", pathEnv, err)
		}
		return resolved, nil
	}
	if path, err := lookPath("grok"); err == nil {
		return filepath.Abs(path)
	}
	for _, candidate := range candidates {
		if isExecutableFile(candidate) {
			return filepath.Clean(candidate), nil
		}
	}
	return "", fmt.Errorf("未找到 Grok CLI；请安装 Grok CLI，或通过 %s 指定可执行文件路径", pathEnv)
}

func resolveExplicit(value string, lookPath func(string) (string, error)) (string, error) {
	if !strings.ContainsRune(value, filepath.Separator) {
		path, err := lookPath(value)
		if err != nil {
			return "", err
		}
		return filepath.Abs(path)
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	if !isExecutableFile(path) {
		return "", fmt.Errorf("%s 不是可执行文件", path)
	}
	return path, nil
}

func darwinCandidates(home string) []string {
	return []string{
		"/opt/homebrew/bin/grok",
		"/usr/local/bin/grok",
		filepath.Join(home, ".local", "bin", "grok"),
		filepath.Join(home, ".volta", "bin", "grok"),
		filepath.Join(home, ".npm-global", "bin", "grok"),
		filepath.Join(home, "bin", "grok"),
	}
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0
}
