//go:build darwin

package autostart

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	launchAgentLabel = "com.grokbuildswitch.app"
	plistFileName    = launchAgentLabel + ".plist"
)

type plistValue struct {
	kind   string
	text   string
	items  []string
	truthy bool
}

type plistEntry struct {
	key   string
	value plistValue
}

// Enable 创建用户级 LaunchAgent，使应用在下次登录时静默启动。
func Enable(exePath string, silent bool) error {
	arguments := []string{}
	if silent {
		arguments = append(arguments, "--silent")
	}
	return EnableWithArguments(exePath, arguments)
}

// EnableWithArguments 创建包含指定启动参数的用户级 LaunchAgent。
func EnableWithArguments(exePath string, arguments []string) error {
	resolvedPath, err := filepath.Abs(strings.TrimSpace(exePath))
	if err != nil || strings.TrimSpace(exePath) == "" {
		return fmt.Errorf("应用程序路径无效")
	}
	if info, statErr := os.Stat(resolvedPath); statErr != nil {
		return fmt.Errorf("应用程序不存在: %w", statErr)
	} else if info.IsDir() {
		return fmt.Errorf("应用程序路径不能是目录")
	}

	data, err := renderLaunchAgentWithArguments(resolvedPath, arguments)
	if err != nil {
		return err
	}
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	return atomicWrite(path, data, 0o644)
}

// Disable 删除用户级 LaunchAgent；重复调用也视为成功。
func Disable() error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsEnabled 返回 LaunchAgent 是否存在以及其中保存的启动命令。
func IsEnabled() (bool, string, error) {
	path, err := launchAgentPath()
	if err != nil {
		return false, "", err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	arguments, err := parseProgramArguments(data)
	if err != nil {
		return false, "", fmt.Errorf("读取登录项失败: %w", err)
	}
	return true, strings.Join(arguments, " "), nil
}

// Sync 将期望状态同步到 macOS 用户级 LaunchAgent。
func Sync(enabled bool, exePath string, silent bool) error {
	if enabled {
		return Enable(exePath, silent)
	}
	return Disable()
}

// SyncWithArguments 使用指定参数同步 macOS 用户级 LaunchAgent。
func SyncWithArguments(enabled bool, exePath string, arguments []string) error {
	if enabled {
		return EnableWithArguments(exePath, arguments)
	}
	return Disable()
}

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistFileName), nil
}

func renderLaunchAgent(exePath string, silent bool) ([]byte, error) {
	arguments := []string{exePath}
	if silent {
		arguments = append(arguments, "--silent")
	}
	return renderLaunchAgentArguments(arguments)
}

func renderLaunchAgentWithArguments(exePath string, arguments []string) ([]byte, error) {
	programArguments := make([]string, 1, len(arguments)+1)
	programArguments[0] = exePath
	programArguments = append(programArguments, arguments...)
	return renderLaunchAgentArguments(programArguments)
}

func renderLaunchAgentArguments(arguments []string) ([]byte, error) {
	entries := []plistEntry{
		{key: "Label", value: plistValue{kind: "string", text: launchAgentLabel}},
		{key: "ProgramArguments", value: plistValue{kind: "array", items: arguments}},
		{key: "RunAtLoad", value: plistValue{kind: "bool", truthy: true}},
		{key: "KeepAlive", value: plistValue{kind: "bool", truthy: false}},
		{key: "ProcessType", value: plistValue{kind: "string", text: "Background"}},
		{key: "LimitLoadToSessionType", value: plistValue{kind: "string", text: "Aqua"}},
	}

	var output bytes.Buffer
	output.WriteString(xml.Header)
	output.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	encoder := xml.NewEncoder(&output)
	encoder.Indent("", "  ")
	if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "plist"}, Attr: []xml.Attr{{Name: xml.Name{Local: "version"}, Value: "1.0"}}}); err != nil {
		return nil, err
	}
	if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "dict"}}); err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if err := encodeElement(encoder, "key", entry.key); err != nil {
			return nil, err
		}
		switch entry.value.kind {
		case "string":
			if err := encodeElement(encoder, "string", entry.value.text); err != nil {
				return nil, err
			}
		case "array":
			if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "array"}}); err != nil {
				return nil, err
			}
			for _, item := range entry.value.items {
				if err := encodeElement(encoder, "string", item); err != nil {
					return nil, err
				}
			}
			if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "array"}}); err != nil {
				return nil, err
			}
		case "bool":
			name := "false"
			if entry.value.truthy {
				name = "true"
			}
			if err := encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: name}}); err != nil {
				return nil, err
			}
			if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: name}}); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("不支持的 plist 值类型: %s", entry.value.kind)
		}
	}
	if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "dict"}}); err != nil {
		return nil, err
	}
	if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "plist"}}); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	output.WriteByte('\n')
	return output.Bytes(), nil
}

func encodeElement(encoder *xml.Encoder, name, value string) error {
	start := xml.StartElement{Name: xml.Name{Local: name}}
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	if err := encoder.EncodeToken(xml.CharData(value)); err != nil {
		return err
	}
	return encoder.EncodeToken(start.End())
}

func parseProgramArguments(data []byte) ([]string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	currentKey := ""
	inProgramArguments := false
	var arguments []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			if end, endOK := token.(xml.EndElement); endOK && end.Name.Local == "array" && inProgramArguments {
				break
			}
			continue
		}
		switch start.Name.Local {
		case "key":
			if err := decoder.DecodeElement(&currentKey, &start); err != nil {
				return nil, err
			}
		case "array":
			inProgramArguments = currentKey == "ProgramArguments"
		case "string":
			if !inProgramArguments {
				continue
			}
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return nil, err
			}
			arguments = append(arguments, value)
		}
	}
	if len(arguments) == 0 {
		return nil, fmt.Errorf("缺少 ProgramArguments")
	}
	return arguments, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
