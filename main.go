package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"grok_switch/internal/appsettings"
	"grok_switch/internal/crash"
	"grok_switch/internal/grokauth"
	"grok_switch/internal/grokpool"
	"grok_switch/internal/macapp"
	"grok_switch/internal/notify"
	"grok_switch/internal/paths"
	"grok_switch/internal/profiles"
	"grok_switch/internal/server"
	"grok_switch/internal/settings"
	"grok_switch/internal/switcher"
	"grok_switch/internal/tray"
)

//go:embed ui/index.html ui/app.js ui/style.css icon.svg assets/icon.ico
var assets embed.FS

func main() {
	defer crash.RecoverMainThread()

	silent := flag.Bool("silent", false, "启动时不打开浏览器")
	noTray := flag.Bool("no-tray", false, "不启用系统托盘（兼容参数）")
	trayEnabled := flag.Bool("tray", runtime.GOOS != "darwin", "启用系统托盘；macOS 默认关闭")
	flag.Parse()
	useTray := shouldUseTray(*trayEnabled, *noTray)

	resolved, err := paths.Resolve()
	if err != nil {
		fatal(err)
	}
	if err := resolved.Ensure(); err != nil {
		fatal(err)
	}
	// Open the crash log as early as possible so startup failures and any
	// later stderr writes / panics are captured instead of vanishing.
	crash.Setup(resolved.LogFile)

	exePath, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	exePath, _ = filepath.Abs(exePath)

	profileStore := profiles.NewStore(resolved.ProfilesFile)
	settingsStore := settings.NewStore(resolved.SettingsFile)
	appSettings := appsettings.New(settingsStore, exePath)
	grokAuthStore := grokauth.NewStore(resolved.GrokAuthFile)
	grokPool, err := grokpool.NewManager(resolved.GrokPoolDir)
	if err != nil {
		fatal(err)
	}
	if err := grokAuthStore.SetProxyURL(grokPool.Status().Settings.ProxyURL); err != nil {
		fatal(err)
	}
	if singleStatus, statusErr := grokAuthStore.Status(); statusErr == nil && singleStatus.Configured {
		if raw, readErr := os.ReadFile(resolved.GrokAuthFile); readErr == nil {
			if _, migrateErr := grokPool.Ensure([]grokpool.ImportFile{{Name: "legacy-grok-auth.json", Content: string(raw)}}); migrateErr != nil {
				crash.Logf("migrate legacy Grok auth into pool: %v", migrateErr)
			}
		}
	}
	grokPool.Start()
	defer grokPool.Close()
	sw := &switcher.Switcher{
		ConfigPath: resolved.GrokConfig,
		BackupsDir: resolved.BackupsDir,
		Profiles:   profileStore,
	}
	if err := sw.EnsureDefaultProfile(); err != nil {
		crash.Logf("default profile import skipped: %v", err)
	}

	currentSettings, err := settingsStore.Get()
	if err != nil {
		fatal(err)
	}
	if err := appSettings.SyncCurrent(); err != nil {
		crash.Logf("autostart sync failed: %v", err)
	}

	appServer := &server.Server{
		Paths:       resolved,
		Profiles:    profileStore,
		Settings:    settingsStore,
		AppSettings: appSettings,
		GrokAuth:    grokAuthStore,
		GrokPool:    grokPool,
		Switcher:    sw,
		Assets:      assets,
	}
	quitCh := make(chan struct{})
	var quitOnce sync.Once
	appServer.OnQuit = func() {
		quitOnce.Do(func() { close(quitCh) })
	}
	httpServer, port, err := appServer.Listen(currentSettings.Port)
	if err != nil {
		fatal(err)
	}
	// Route net/http's internal panic/error reports into the crash log too.
	if crashFile := resolved.LogFile; crashFile != "" {
		if f, ferr := os.OpenFile(crashFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); ferr == nil {
			httpServer.ErrorLog = log.New(f, "http: ", log.LstdFlags)
		}
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	crash.Logf("http server listening: %s", url)
	if runtime.GOOS == "darwin" && !useTray {
		if err := macapp.Prepare(url); err != nil {
			fatal(err)
		}
	}

	trayApp := &tray.Tray{
		Profiles:    profileStore,
		Settings:    settingsStore,
		AppSettings: appSettings,
		Switcher:    sw,
		URL:         url,
		DataDir:     resolved.DataDir,
		LogFile:     resolved.LogFile,
		AuthFile:    filepath.Join(resolved.GrokHome, "auth.json"),
		Assets:      assets,
	}
	if useTray {
		appServer.SetOnChanged(trayApp.Refresh)
	}

	if !*silent && currentSettings.AutoOpenBrowser {
		_ = tray.OpenBrowser(url)
	}

	if !useTray {
		var exitReason string
		if runtime.GOOS == "darwin" {
			exitReason = runMacApplication(quitCh)
		} else {
			exitReason = waitForExit(quitCh)
		}
		crash.Logf("application exit requested by: %s", exitReason)
		shutdown(httpServer)
		crash.Flush()
		return
	}
	trayApp.Run()
	crash.Logf("application exit requested by: system tray")
	shutdown(httpServer)
	crash.Flush()
}

func waitForExit(quit <-chan struct{}) string {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(ch)
	select {
	case receivedSignal := <-ch:
		return "signal " + receivedSignal.String()
	case <-quit:
		return "web interface"
	}
}

func runMacApplication(quit <-chan struct{}) string {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	exitReason := make(chan string, 1)
	monitorDone := make(chan struct{})
	go crash.Guard("macOS exit monitor", func() {
		select {
		case receivedSignal := <-signalCh:
			exitReason <- "signal " + receivedSignal.String()
		case <-quit:
			exitReason <- "web interface"
		case <-monitorDone:
			return
		}
		macapp.RequestExit()
	})

	macapp.Run()
	close(monitorDone)
	select {
	case reason := <-exitReason:
		return reason
	default:
		return "Dock or Command-Q"
	}
}

func shouldUseTray(trayEnabled, noTray bool) bool {
	return trayEnabled && !noTray
}

func shutdown(srv interface{ Shutdown(context.Context) error }) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func fatal(err error) {
	crash.Logf("fatal startup error: %v", err)
	crash.Flush()
	notify.Alert("Grok Build Switch 启动失败", err.Error())
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
