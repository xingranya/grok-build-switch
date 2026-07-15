package main

import "testing"

func TestShouldUseTrayRespectsDefaultAndLegacyOverride(t *testing.T) {
	tests := []struct {
		name        string
		trayEnabled bool
		noTray      bool
		want        bool
	}{
		{name: "默认后台模式", trayEnabled: false, want: false},
		{name: "显式启用托盘", trayEnabled: true, want: true},
		{name: "兼容参数强制关闭", trayEnabled: true, noTray: true, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldUseTray(test.trayEnabled, test.noTray); got != test.want {
				t.Fatalf("shouldUseTray() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestShouldUseMacAppSkipsDesktopLifecycleInHeadlessMode(t *testing.T) {
	if shouldUseMacApp(false, true) {
		t.Fatal("headless 模式不应创建桌面应用")
	}
	if shouldUseMacApp(true, false) {
		t.Fatal("托盘模式不应重复创建桌面应用")
	}
}
