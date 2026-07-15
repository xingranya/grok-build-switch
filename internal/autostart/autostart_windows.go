//go:build windows

package autostart

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
	name   = "grok_switch"
)

func Enable(exePath string, silent bool) error {
	arguments := []string{}
	if silent {
		arguments = append(arguments, "--silent")
	}
	return EnableWithArguments(exePath, arguments)
}

// EnableWithArguments 使用指定参数创建 Windows 登录启动项。
func EnableWithArguments(exePath string, arguments []string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	value := fmt.Sprintf("%q", exePath)
	for _, argument := range arguments {
		value += " " + fmt.Sprintf("%q", argument)
	}
	return k.SetStringValue(name, value)
}

func Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	err = k.DeleteValue(name)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

func IsEnabled() (bool, string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, "", err
	}
	defer k.Close()
	value, _, err := k.GetStringValue(name)
	if err == registry.ErrNotExist {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return value != "", value, nil
}

func Sync(enabled bool, exePath string, silent bool) error {
	if enabled {
		return Enable(exePath, silent)
	}
	return Disable()
}

// SyncWithArguments 使用指定参数同步 Windows 登录启动项。
func SyncWithArguments(enabled bool, exePath string, arguments []string) error {
	if enabled {
		return EnableWithArguments(exePath, arguments)
	}
	return Disable()
}
