# 风险评估

## S.U.P.E.R 架构健康摘要

| 原则 | 状态 | 关键发现 | 适配优先级 |
|:---|:---|:---|:---|
| **S** 单一职责 | 部分 | 核心包大体按领域拆分，但 `server.go`、`tray.go` 和 `ui/app.js` 偏大 | 中 |
| **U** 单向依赖 | 通过 | `main -> adapter -> core` 方向基本清晰，未发现包循环依赖 | 低 |
| **P** 端口优先 | 部分 | HTTP JSON、Profile 和 Settings 已形成契约；系统命令与具体 Store 难以注入测试 | 高 |
| **E** 环境无关 | 违背 | 自启动仅支持 Windows，构建仅产出 EXE，Finder PATH 未处理，托盘资源为 ICO | 最高 |
| **R** 可替换组件 | 违背 | 自启动文件用 `!windows` 合并全部平台，构建/图标/命令发现没有独立 macOS 适配层 | 高 |

**整体健康度**：1/5 项完全健康。核心业务可复用，但桌面系统集成层需要重构后才能形成可靠 macOS 交付物。

### 主要违背热点

1. `internal/autostart/autostart_other.go`：Darwin 与其他平台共用“不支持”实现，直接阻塞设置保存和托盘自启动。
2. `internal/tray/tray.go`：固定 ICO、裸 `grok` 命令、忽略自启动错误，平台行为与菜单状态耦合。
3. `internal/server/server.go`：重复执行裸 `grok login`，Finder 启动环境下高概率无法继承终端 PATH。
4. `build.ps1` 与缺失的 macOS 资源：无法产出符合 macOS 目录结构的 `.app`，也没有 DMG/签名/架构验证。
5. 测试治理：平台适配层、设置保存和托盘交互无自动化保护。

## 风险矩阵

| 风险 | 影响 | 可能性 | 等级 | 缓解措施 |
|:---|:---|:---|:---|:---|
| macOS 开启自启动必然失败 | 用户无法保存相关设置，菜单状态失真 | 确定 | P0 | 实现 Darwin LaunchAgent，并补幂等/转义测试 |
| Finder 启动后找不到 `grok` | 官方账号登录按钮失败 | 高 | P0 | 统一命令发现，检查 PATH 和常见安装位置，返回可执行的错误信息 |
| ICO 在菜单栏显示异常或空白 | 应用缺少可识别状态栏图标 | 高 | P0 | 为 Darwin 嵌入 PNG/template 图标，保留 Windows ICO |
| 只生成裸二进制 | 无法像正常 macOS 应用安装和启动 | 确定 | P0 | 构建标准 `.app`，写入 `Info.plist`、资源与 `LSUIElement` |
| Apple Silicon 产物无法运行于 Intel | 公开发布覆盖不足 | 中 | P1 | 明确架构范围；公开发行时构建双架构并合并 Universal 2 |
| 未签名/未公证被 Gatekeeper 拦截 | 下载用户首次启动失败或警告 | 高 | P1 | 本机版使用 ad-hoc 签名；公开版预留 Developer ID 签名和公证参数 |
| macOS 系统命令缺少测试端口 | 改动难以在 CI 验证 | 高 | P1 | 把 plist 生成与命令查找设计为纯函数，系统执行保持薄适配层 |
| 数据目录改为原生路径导致历史数据丢失 | Windows/macOS 或旧版数据不兼容 | 中 | P1 | 本次继续使用 `~/.grok_switch`，不做隐式迁移 |
| 当前机器缺少 Go 1.26 | 本地无法编译和跑测试 | 确定 | P0 前置 | 安装匹配工具链，或先用 `macos-14` CI 建立可重复构建 |
| `.git/config` 只读 | 无法自动设置 fork 远程或提交推送 | 确定 | 流程风险 | 保留已创建的远端 fork；实施完成后由有写权限环境整理远程和提交 |

## 高等级风险说明

### 自启动状态可能先写入、后失败

Web 设置保存时先持久化 `settings.json`，再调用 `autostart.Sync`。macOS 当前必然返回不支持错误，于是 API 返回失败，但磁盘中的 `autostart=true` 已经保留。应用下次启动又会重复同步失败。托盘路径更严重：它保存设置后忽略 `Sync` 错误并显示“已开启”。实施时需要让平台操作和设置持久化形成一致结果，至少在失败时回滚设置或先验证平台操作。

### `.app` 与裸二进制的运行环境不同

从终端执行时，`exec.Command("grok", "login")` 能使用 shell PATH；从 Finder 或登录项启动 `.app` 时通常只有系统级 PATH，不一定包含 `/opt/homebrew/bin`、`/usr/local/bin` 或用户工具目录。macOS 适配不能只做到“能编译”，还必须从 `.app` 启动后验证官方登录入口。

### 发行范围决定构建设计

本机 Apple Silicon 自用版可以只构建 `arm64` 并使用 ad-hoc 签名。面向 Release 的公开版本应至少生成 Universal 2 或分别提供 `arm64`/`amd64`，并处理 Developer ID 签名、公证和 stapling。两者在 CI、凭据和验收成本上差异较大，必须在计划阶段确认。

## 技术债

- `server.go` 和 `ui/app.js` 过大，但与本次平台适配没有直接因果，不应借机大规模重构。
- `notify` 将三类系统能力集中在一个包，后续可按命令执行端口拆分；本次只处理影响 macOS 稳定性的部分。
- `tray` 直接持有多个具体 Store 和 Switcher，单元测试困难；本次优先抽离可纯测的平台逻辑。
- API 使用 Go 结构体和运行时代码作为契约，没有 OpenAPI/JSON Schema；本次保持字段兼容，不引入新协议层。

## 测试风险

- 已有核心测试无法覆盖 LaunchAgent、`.app` 目录、Finder PATH、菜单栏图标和 Gatekeeper 行为。
- 需要新增 Darwin 可运行的单元测试：plist 内容、参数转义、启用/禁用幂等性、命令候选优先级。
- 需要新增构建产物静态检查：`Info.plist` 键、二进制架构、资源存在、`LSUIElement=true`。
- 最终仍需要一次真实 macOS 手工冒烟：双击 `.app`、菜单栏图标、打开面板、切换配置、`grok login`、启用/关闭登录项、退出后重登。
- 当前环境没有 Go，可先完成纯文件与脚本审查，但不能把静态检查当作编译通过。

## 项目治理风险

- 当前仓库没有持久化指令或记忆入口，本次分析遵循会话内项目指令；后续如需新增 `AGENTS.md`，必须保留用户提供的规则并先确认其作为共享规范的权威性。
- GitHub Issues 在 fork 中关闭，Projects scope 也不可用，因此选择 `LOCAL_ONLY`，避免制造一半在 GitHub、一半在本地的双重真源。
- `.codegraph/` 是本轮按项目要求初始化的索引，应明确是否纳入版本控制；默认不把数据库当成产品源码提交。

## 兼容性关注

- 保持 `GROK_HOME`、`GROK_CONFIG`、`~/.grok_switch`、JSON 字段和 HTTP API 不变。
- 保持 Windows 构建与注册表自启动，不用 macOS 改动替换现有实现。
- macOS LaunchAgent 的 label、plist 路径和应用标识一旦发布就应稳定，避免升级后遗留重复登录项。
- 应用名称可以显示为 `Grok Build Switch`，内部数据目录暂时继续使用 `grok_switch`，避免迁移风险。
