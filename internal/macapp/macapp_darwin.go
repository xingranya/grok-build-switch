//go:build darwin && cgo

// Package macapp 提供 macOS 原生应用生命周期。
package macapp

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

int macapp_prepare(const char *web_url);
void macapp_run(void);
void macapp_request_exit(void);
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

func init() {
	// AppKit 必须在进程主线程运行；在 init 中锁定可确保 main 延续在该线程。
	runtime.LockOSThread()
}

// Prepare 创建 Dock 应用、应用菜单以及重新打开网页所需的回调。
func Prepare(webURL string) error {
	cWebURL := C.CString(webURL)
	defer C.free(unsafe.Pointer(cWebURL))
	switch result := C.macapp_prepare(cWebURL); result {
	case 1:
		return nil
	case -1:
		return fmt.Errorf("AppKit 未在 macOS 主线程初始化")
	case -2:
		return fmt.Errorf("AppKit 初始化异常，详情请查看应用日志")
	default:
		return fmt.Errorf("无法初始化 macOS Dock 应用")
	}
}

// Run 在主线程运行 AppKit 事件循环，直到用户请求退出。
func Run() {
	C.macapp_run()
}

// RequestExit 安全地通知 AppKit 结束事件循环。
func RequestExit() {
	C.macapp_request_exit()
}
