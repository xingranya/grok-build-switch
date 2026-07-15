//go:build !windows && !darwin

package autostart

import (
	"fmt"
	"runtime"
)

// Enable 在当前平台不支持登录时启动时返回明确错误。
func Enable(exePath string, silent bool) error {
	return fmt.Errorf("当前平台 %s 暂不支持开机自启", runtime.GOOS)
}

// Disable 在不支持登录时启动的平台保持幂等。
func Disable() error {
	return nil
}

// IsEnabled 返回当前平台未启用登录时启动。
func IsEnabled() (bool, string, error) {
	return false, "", nil
}

// Sync 将期望状态同步到当前平台支持的登录项。
func Sync(enabled bool, exePath string, silent bool) error {
	if enabled {
		return Enable(exePath, silent)
	}
	return Disable()
}
