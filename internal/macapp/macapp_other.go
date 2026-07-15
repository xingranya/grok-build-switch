//go:build !darwin

// Package macapp 提供 macOS 原生应用生命周期。
package macapp

// Prepare 在非 macOS 系统上不执行操作。
func Prepare(string) error {
	return nil
}

// Run 在非 macOS 系统上不执行操作。
func Run() {}

// RequestExit 在非 macOS 系统上不执行操作。
func RequestExit() {}
