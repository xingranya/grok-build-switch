//go:build darwin

package autostart

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderLaunchAgentEscapesPathAndAddsSilentArgument(t *testing.T) {
	data, err := renderLaunchAgent("/Applications/Grok & Switch.app/Contents/MacOS/grok_switch", true)
	if err != nil {
		t.Fatal(err)
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		if _, err := decoder.Token(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("plist 不是有效 XML: %v", err)
		}
	}
	text := string(data)
	if !strings.Contains(text, "Grok &amp; Switch.app") {
		t.Fatalf("路径没有正确转义:\n%s", text)
	}
	arguments, err := parseProgramArguments(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(arguments) != 2 || arguments[1] != "--silent" {
		t.Fatalf("启动参数不正确: %#v", arguments)
	}
}

func TestEnableDisableAndIsEnabledAreIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	executable := filepath.Join(home, "Grok & Switch.app", "Contents", "MacOS", "grok_switch")
	if err := os.MkdirAll(filepath.Dir(executable), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Enable(executable, true); err != nil {
		t.Fatal(err)
	}
	path, err := launchAgentPath()
	if err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Enable(executable, true); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("重复启用生成了不同的 LaunchAgent")
	}

	enabled, command, err := IsEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if !enabled || !strings.Contains(command, executable) || !strings.HasSuffix(command, "--silent") {
		t.Fatalf("登录项状态不正确: enabled=%v command=%q", enabled, command)
	}

	if err := Disable(); err != nil {
		t.Fatal(err)
	}
	if err := Disable(); err != nil {
		t.Fatal(err)
	}
	enabled, _, err = IsEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if enabled {
		t.Fatal("删除后登录项仍显示为启用")
	}
}

func TestEnableRejectsMissingExecutable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := Enable(filepath.Join(t.TempDir(), "missing"), true); err == nil {
		t.Fatal("不存在的应用程序路径应返回错误")
	}
}
