package appsettings

import (
	"errors"
	"strings"
	"testing"

	"grok_switch/internal/settings"
)

type fakeStore struct {
	current   settings.Settings
	updateErr error
	events    *[]string
}

func (s *fakeStore) Get() (settings.Settings, error) {
	return s.current, nil
}

func (s *fakeStore) Update(next settings.Settings) (settings.Settings, error) {
	if s.events != nil {
		*s.events = append(*s.events, "update")
	}
	if s.updateErr != nil {
		return settings.Settings{}, s.updateErr
	}
	s.current = next
	return next, nil
}

func TestUpdateSyncsBeforePersisting(t *testing.T) {
	events := []string{}
	store := &fakeStore{current: settings.Default(), events: &events}
	service := NewWithSync(store, "/Applications/Grok Build Switch.app/Contents/MacOS/grok_switch", func(enabled bool, _ string, silent bool) error {
		if !enabled || !silent {
			t.Fatalf("同步参数不正确: enabled=%v silent=%v", enabled, silent)
		}
		events = append(events, "sync")
		return nil
	})

	next := settings.Default()
	next.Autostart = true
	updated, err := service.Update(next)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Autostart || !updated.SilentAutostart {
		t.Fatalf("设置没有归一化: %#v", updated)
	}
	if strings.Join(events, ",") != "sync,update" {
		t.Fatalf("执行顺序不正确: %#v", events)
	}
}

func TestUpdateDoesNotPersistWhenSyncFails(t *testing.T) {
	events := []string{}
	store := &fakeStore{current: settings.Default(), events: &events}
	service := NewWithSync(store, "/tmp/grok_switch", func(bool, string, bool) error {
		events = append(events, "sync")
		return errors.New("sync failed")
	})

	next := settings.Default()
	next.Autostart = true
	if _, err := service.Update(next); err == nil || !strings.Contains(err.Error(), "同步系统登录项失败") {
		t.Fatalf("应返回同步错误，实际为 %v", err)
	}
	if strings.Join(events, ",") != "sync" {
		t.Fatalf("同步失败后不应写入设置: %#v", events)
	}
	if store.current.Autostart {
		t.Fatal("同步失败后设置被错误修改")
	}
}

func TestUpdateRollsBackSyncWhenPersistingFails(t *testing.T) {
	events := []string{}
	current := settings.Default()
	store := &fakeStore{current: current, updateErr: errors.New("disk failed"), events: &events}
	service := NewWithSync(store, "/tmp/grok_switch", func(enabled bool, _ string, _ bool) error {
		if enabled {
			events = append(events, "sync:on")
		} else {
			events = append(events, "sync:off")
		}
		return nil
	})

	next := current
	next.Autostart = true
	if _, err := service.Update(next); err == nil || !strings.Contains(err.Error(), "系统登录项已恢复") {
		t.Fatalf("应返回已回滚错误，实际为 %v", err)
	}
	if strings.Join(events, ",") != "sync:on,update,sync:off" {
		t.Fatalf("回滚顺序不正确: %#v", events)
	}
}

func TestSyncCurrentUsesPersistedState(t *testing.T) {
	current := settings.Default()
	current.Autostart = true
	current.SilentAutostart = true
	store := &fakeStore{current: current}
	called := false
	service := NewWithSync(store, "/tmp/grok_switch", func(enabled bool, path string, silent bool) error {
		called = enabled && silent && path == "/tmp/grok_switch"
		return nil
	})
	if err := service.SyncCurrent(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("没有按磁盘状态同步登录项")
	}
}
