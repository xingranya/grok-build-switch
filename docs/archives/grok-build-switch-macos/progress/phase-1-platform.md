# Phase 1：平台适配基础

**目标**：消除阻塞 macOS 正常运行的系统集成缺口。  
**状态**：已完成

## 任务

- [x] **T1.1：实现 macOS LaunchAgent 自启动**
  - 优先级：P0
  - 工作量：M
  - 测试要求：新增 plist、启停、幂等与路径测试
  - 记忆影响：在进度记录中固定 LaunchAgent label 和路径
  - 验收：Darwin 独立实现通过测试；Windows 文件不变；其他平台明确返回不支持
  - 备注：已完成；固定 label 为 `com.grokbuildswitch.app`，plist 为 `~/Library/LaunchAgents/com.grokbuildswitch.app.plist`；Darwin 测试通过
- [x] **T1.2：保证自启动设置与系统状态一致**
  - 优先级：P0
  - 工作量：M
  - 测试要求：覆盖失败回滚或“系统成功后持久化”的行为
  - 记忆影响：无
  - 验收：API 与托盘不再吞掉自启动失败，也不会留下错误勾选状态
  - 备注：已完成；`internal/appsettings` 统一系统同步、持久化和失败回滚；完整 Go 测试通过
- [x] **T1.3：统一 Grok CLI 发现与登录启动**
  - 优先级：P0
  - 工作量：M
  - 测试要求：覆盖环境变量、PATH、常见 macOS 路径和错误信息
  - 记忆影响：记录 Finder PATH 限制
  - 验收：server 与 tray 不再重复执行裸 `grok` 命令
  - 备注：已完成；解析顺序为 `GROK_CLI`、PATH、macOS 常见目录；完整 Go 测试通过
- [x] **T1.4：提供 macOS 菜单栏资源**
  - 优先级：P0
  - 工作量：S
  - 测试要求：资源存在性和平台选择静态检查
  - 记忆影响：无
  - 验收：macOS 使用 PNG/template 图标；Windows 保持 ICO
  - 备注：已完成；生成 64px 透明 template PNG，Darwin 使用 `SetTemplateIcon`，图标测试与全仓测试通过

## 阶段备注

- 保持现有 `~/.grok_switch` 数据位置，不迁移到 `Application Support`。
- 当前执行规则不授权子代理，理论并行 Lane 改为顺序执行。
- T1.4 的系统 Quick Look 转换被沙箱拒绝，改用临时 Go SVG 栅格化器；任务漂移达到 replan 阈值。Phase 1 已无剩余任务，因此不重拆本阶段，并把经验应用到 T2.1：提交预生成 ICNS，构建脚本不依赖 SVG 转换工具。

## 阶段完成清单

- [x] 上述任务全部完成
- [x] `MASTER.md` 阶段计数已更新
- [x] `MASTER.md` 当前状态已切换到 Phase 2
