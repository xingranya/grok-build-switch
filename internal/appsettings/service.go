package appsettings

import (
	"fmt"

	"grok_switch/internal/autostart"
	"grok_switch/internal/settings"
)

// Store 定义应用设置服务需要的最小持久化端口。
type Store interface {
	Get() (settings.Settings, error)
	Update(settings.Settings) (settings.Settings, error)
}

// SyncFunc 定义登录项同步端口，便于隔离平台实现并进行测试。
type SyncFunc func(enabled bool, exePath string, silent bool) error

// Service 保证应用设置和系统登录项以一致状态更新。
type Service struct {
	store   Store
	exePath string
	sync    SyncFunc
}

// New 使用当前平台的登录项实现创建设置服务。
func New(store Store, exePath string) *Service {
	return NewWithSync(store, exePath, autostart.Sync)
}

// NewWithSync 使用指定登录项端口创建可测试的设置服务。
func NewWithSync(store Store, exePath string, sync SyncFunc) *Service {
	return &Service{store: store, exePath: exePath, sync: sync}
}

// SyncCurrent 将磁盘中的当前设置同步到系统登录项。
func (s *Service) SyncCurrent() error {
	if err := s.validate(); err != nil {
		return err
	}
	current, err := s.store.Get()
	if err != nil {
		return err
	}
	return s.sync(current.Autostart, s.exePath, current.SilentAutostart)
}

// Update 先更新系统登录项，再保存设置；保存失败时恢复原登录项。
func (s *Service) Update(next settings.Settings) (settings.Settings, error) {
	if err := s.validate(); err != nil {
		return settings.Settings{}, err
	}
	current, err := s.store.Get()
	if err != nil {
		return settings.Settings{}, err
	}
	target := settings.Normalize(next)
	if err := s.sync(target.Autostart, s.exePath, target.SilentAutostart); err != nil {
		return settings.Settings{}, fmt.Errorf("同步系统登录项失败: %w", err)
	}
	updated, err := s.store.Update(target)
	if err == nil {
		return updated, nil
	}
	if rollbackErr := s.sync(current.Autostart, s.exePath, current.SilentAutostart); rollbackErr != nil {
		return settings.Settings{}, fmt.Errorf("保存设置失败: %v；恢复原登录项也失败: %w", err, rollbackErr)
	}
	return settings.Settings{}, fmt.Errorf("保存设置失败，系统登录项已恢复: %w", err)
}

func (s *Service) validate() error {
	if s == nil || s.store == nil {
		return fmt.Errorf("设置服务未初始化")
	}
	if s.sync == nil {
		return fmt.Errorf("登录项同步服务未初始化")
	}
	return nil
}
