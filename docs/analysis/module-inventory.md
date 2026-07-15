# 模块清单

评分含义：`通过` 表示当前满足原则，`部分` 表示边界存在但仍有耦合，`违背` 表示本次适配必须处理。

| 模块 | 职责 | 主要依赖 | 规模 | 复杂度 | S.U.P.E.R 评分 |
|:---|:---|:---|---:|:---|:---|
| `main` | 组合并启动全部组件 | 全部顶层服务 | 155 行 | 中 | S通过 U通过 P部分 E部分 R部分 |
| `autostart` | 管理登录时启动 | Windows 注册表 | 87 行 | 中 | S通过 U通过 P部分 E违背 R部分 |
| `config` | 解析、预览和重写 Grok TOML | `profiles` | 1127 行（含测试） | 高 | S部分 U通过 P通过 E通过 R部分 |
| `crash` | 记录日志并恢复 goroutine panic | 标准库 | 83 行 | 低 | S通过 U通过 P部分 E通过 R通过 |
| `grokauth` | 导入、保存、刷新和代理 OAuth 凭据 | `netproxy` | 725 行（含测试） | 高 | S部分 U部分 P部分 E通过 R部分 |
| `grokpool` | 多账号存储、巡检、轮换和调度 | `grokauth`、`netproxy` | 1579 行（含测试） | 高 | S部分 U部分 P通过 E通过 R部分 |
| `netproxy` | 校验代理 URL 并构造 HTTP Transport | 标准库 | 72 行（含测试） | 低 | S通过 U通过 P通过 E通过 R通过 |
| `notify` | 通知、打开路径和剪贴板适配 | OS 命令 | 98 行 | 中 | S部分 U通过 P部分 E部分 R部分 |
| `paths` | 解析并创建应用数据路径 | 标准库 | 54 行 | 低 | S通过 U通过 P通过 E部分 R通过 |
| `profiles` | Profile 模型、归一化与持久化 | 标准库 | 493 行（含测试） | 中 | S通过 U通过 P通过 E通过 R通过 |
| `server` | 暴露 HTTP API 和嵌入静态资源 | 多个核心模块 | 1662 行（含测试） | 高 | S违背 U部分 P部分 E部分 R部分 |
| `settings` | 应用设置持久化与归一化 | 标准库 | 155 行 | 低 | S通过 U通过 P通过 E通过 R通过 |
| `switcher` | 备份、切换和校验当前配置 | `config`、`profiles` | 276 行 | 中 | S通过 U通过 P部分 E通过 R部分 |
| `tray` | 菜单栏展示、交互与系统命令 | 多个适配模块 | 326 行 | 高 | S部分 U部分 P部分 E违背 R违背 |
| `ui` | 浏览器端管理供应商、账号池和设置 | HTTP JSON API | 2949 行 | 高 | S违背 U通过 P部分 E通过 R部分 |
| 构建与发布 | 生成桌面交付物和文档站 | PowerShell、Actions | 2 个入口 | 中 | S部分 U通过 P部分 E违背 R违背 |

## 模块详情

### `main`

- **路径**：`main.go`
- **公共入口**：`main`、`waitForSignal`、`shutdown`
- **评估**：组合根职责清晰，依赖方向由入口指向内部模块。当前直接传递具体存储类型，替换成本一般。平台问题主要来自它无条件同步自启动，以及 `.app` 启动环境的可执行文件路径。
- **适配要点**：保持组合结构，只接入 macOS 自启动实现与统一的 Grok CLI 启动端口。

### `autostart`

- **路径**：`internal/autostart/`
- **公共 API**：`Enable`、`Disable`、`IsEnabled`、`Sync`
- **评估**：API 形状稳定且调用方向清楚，但 `!windows` 实现把所有非 Windows 平台都视为不支持，macOS 开启自启动必然失败。
- **适配要点**：拆分为 `windows`、`darwin`、`other` 三个构建目标；macOS 使用用户级 LaunchAgent，原子写入 plist，并补路径转义与幂等测试。

### `config`

- **路径**：`internal/config/`
- **公共 API**：Profile 导入、应用、预览、官方认证切换、隐私保护和一致性检查。
- **评估**：输入输出使用 TOML 字节与 `profiles.Profile`，平台假设很少；单文件 621 行且兼具解析与文本保留式重写，维护复杂度较高，但不属于 macOS 适配热点。
- **适配要点**：保持行为不变，用现有测试防止配置回归。

### `crash`

- **路径**：`internal/crash/crash.go`
- **公共 API**：`Setup`、`Logf`、`Guard`、`RecoverMainThread`
- **评估**：职责单一且跨平台。注释只描述 Windows GUI 背景，但实现可继续用于无终端的 macOS `.app`。
- **适配要点**：仅更新平台说明，不修改日志格式和目录，避免破坏排障兼容性。

### `grokauth`

- **路径**：`internal/grokauth/`
- **公共 API**：凭据导入、状态、Token、刷新、删除和请求鉴权。
- **评估**：包含解析、持久化、网络刷新和鉴权多个职责，但已通过 `netproxy` 隔离代理构造；非 Windows 文件权限逻辑适用于 macOS。
- **适配要点**：不重构核心；重点验证从 Finder 启动后仍能触发 Grok CLI 登录流程。

### `grokpool`

- **路径**：`internal/grokpool/`
- **公共 API**：Manager 生命周期、导入、巡检、设置、账号操作、Token 轮换和 Transport。
- **评估**：包内按 types/store/manager/inspect 分文件，职责已有拆分；Manager 仍同时承担调度、存储和选择策略。平台依赖很少且测试较完整。
- **适配要点**：保持现有数据位置和文件权限，避免 macOS 版本与 Windows 版本形成两套账号池格式。

### `netproxy`

- **路径**：`internal/netproxy/`
- **公共 API**：`BuildTransport`
- **评估**：纯逻辑、契约明确、测试完整，是当前最符合 S.U.P.E.R 的模块之一。
- **适配要点**：无需修改。

### `notify`

- **路径**：`internal/notify/notify.go`
- **公共 API**：`Info`、`OpenPath`、`CopyText`
- **评估**：已经为 macOS 使用 `osascript`、`open` 和 `pbcopy`，基本能力齐全；三个系统能力集中在一个包，且命令执行缺少可注入端口，因此不易单测。
- **适配要点**：保留现有 API；为通知参数使用安全传参方式，并增加静态或命令构造测试。

### `paths`

- **路径**：`internal/paths/paths.go`
- **公共 API**：`Resolve`、`Paths.Ensure`
- **评估**：使用用户主目录与 `filepath`，可直接在 macOS 运行。默认 `~/.grok_switch` 不符合原生 `Application Support` 惯例，但它保证跨平台数据兼容。
- **适配要点**：本次不迁移数据目录；如未来迁移，必须提供一次性迁移和回退策略。

### `profiles`

- **路径**：`internal/profiles/`
- **公共 API**：Profile 模型、归一化、CRUD 和激活状态。
- **评估**：模型与存储职责分文件，使用原子写入，非 Windows 权限处理适合 macOS。
- **适配要点**：无需修改，以既有测试保护兼容性。

### `server`

- **路径**：`internal/server/`
- **公共 API**：`Server.Listen`、22 个 HTTP 路由。
- **评估**：JSON API 是清晰的 UI 端口，但 `server.go` 约 794 行，混合路由、设置、模型探测、配置编辑和静态资源。`handleOfficialActivate` 直接执行 `grok login`，与托盘实现重复，且 Finder 环境下可能因 PATH 缺失失败。
- **适配要点**：提取统一 Grok CLI 启动器供 server/tray 复用；设置更新继续通过 `autostart` 端口，不把 LaunchAgent 细节泄漏到 HTTP 层。

### `settings`

- **路径**：`internal/settings/settings.go`
- **公共 API**：`Default`、Store 的读取、更新和端口持久化。
- **评估**：契约明确、原子写入、平台差异局部化。缺少测试，但逻辑简单。
- **适配要点**：沿用 `autostart` 与 `silent_autostart` 字段，避免修改前端 API。

### `switcher`

- **路径**：`internal/switcher/switcher.go`
- **公共 API**：导入、激活、备份、恢复、直接编辑和一致性检查。
- **评估**：依赖方向为 switcher -> config/profiles，没有反向调用；接口是具体 Store，替换性一般。
- **适配要点**：无需平台修改。

### `tray`

- **路径**：`internal/tray/tray.go`
- **公共 API**：`Tray.Run`、`Tray.Refresh`、`OpenBrowser`、`StartGrokLogin`
- **评估**：菜单构造、事件生命周期、业务调用和系统命令集中在同一文件。浏览器打开已有 Darwin 分支，但图标固定读取 ICO，Grok CLI 使用裸命令名，自启动错误被忽略。
- **适配要点**：使用 macOS 可识别的 PNG/template 图标；复用统一 CLI 启动器；自启动切换失败时不得先保存成功状态或显示成功通知。

### `ui`

- **路径**：`ui/`
- **公共端口**：`/api/*` JSON API。
- **评估**：浏览器技术栈与平台无关，但 `app.js` 1639 行，状态管理与所有功能交互集中。自启动字段可直接复用。
- **适配要点**：只调整面向用户的 macOS 安装/状态文案，不做无关前端重构。

### 构建与发布

- **路径**：`build.ps1`、`.github/workflows/pages.yml`、资源文件。
- **评估**：应用构建完全绑定 Windows；现有 Actions 只发布 MkDocs。缺少应用标识、版本、菜单栏模式、签名和架构矩阵契约。
- **适配要点**：新增可复现的 macOS 构建脚本、`Info.plist` 模板和产物校验；发布策略取决于目标是本机自用、公开未公证下载还是签名公证发行。
