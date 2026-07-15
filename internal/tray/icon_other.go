//go:build !darwin

package tray

import (
	"embed"

	"fyne.io/systray"
)

func setTrayIcon(assets embed.FS) {
	icon, err := assets.ReadFile("assets/icon.ico")
	if err != nil || len(icon) == 0 {
		icon = iconData
	}
	systray.SetIcon(icon)
}
