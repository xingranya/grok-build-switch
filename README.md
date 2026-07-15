# Grok Build Switch

Windows 托盘与 macOS 本地 Web 工具：用供应商（Profile）管理 Grok CLI 的 `~/.grok/config.toml`。

一键切换上游 `base_url`、默认模型、联网搜索模型、subagents 与各 `[model.*]` 定义。

## 功能

- 供应商增删改查：名称、Base URL、API Key、上游格式、默认 / 联网 / Subagents 模型、已启用模型列表
- 供应商默认使用 `high` 推理强度；每个模型自动写入 `supports_reasoning_effort = true` 和 `low/medium/high` 支持列表
- 一键启用：写入 `[endpoints]`、`[models]`、`[subagents].default_model` 与 `[model.*]`，其它段尽量保留
- 切换 / 保存 config 前自动备份；设置页可还原备份、直接编辑 `config.toml`
- 首次运行可从当前 `config.toml` 导入 Default 供应商
- 导入 CPA `xai-*.json` 或 Grok CLI `auth.json`，由内嵌代理提供稳定的本地 URL/key 并自动刷新 token
- Grok 多账号池：批量导入、定时自动巡检、健康分类、坏号自动隔离、健康号轮换与单账号回退
- Web UI 仅监听 `127.0.0.1`（默认端口 `17878`，被占用时自动递增）
- 可设置 Windows 开机自启或 macOS 登录时启动
- Windows 托盘菜单：快速切换、打开面板、复制地址、打开数据/日志目录
- macOS 默认直接运行本地 Web 面板，不初始化菜单栏组件

## 系统要求

| 项目 | 说明 |
|------|------|
| macOS | **macOS 12+，Apple Silicon（arm64）** |
| Windows | **Windows 10 / 11 x64** |
| 运行 | macOS 打开 `Grok Build Switch.app`；Windows 打开 `grok_switch.exe`，均无需安装 Go / Node |
| 可选 | 本机已安装 [Grok CLI](https://x.ai)，配置目录默认为 `~/.grok` |

## 安装与使用

### 在线文档

使用教程、截图说明和联系方式会整理在项目文档站：

[https://1parado.github.io/grok-build-switch/](https://1parado.github.io/grok-build-switch/)

### macOS 安装

1. 下载 `Grok-Build-Switch-macos-arm64.dmg` 并打开。
2. 将 `Grok Build Switch.app` 拖入“应用程序”。
3. 首次启动时按住 Control 点击应用并选择“打开”；如果系统仍拦截，请在“系统设置 → 隐私与安全性”中选择“仍要打开”。
4. 浏览器会自动打开 `http://127.0.0.1:17878/`；需要结束后台进程时，在“设置 → 应用”中点击“退出应用”。

当前 macOS 版本使用 ad-hoc 签名，未进行 Apple Developer ID 公证。只应运行来自本仓库构建或你自行构建的产物。

### Windows 安装

1. 打开本仓库的 [Releases](../../releases) 页面。
2. 下载 `grok_switch.exe`。
3. 双击运行，系统托盘出现图标后即可使用。



### 从源码构建

见下方 [构建](#构建)。

### 日常操作

1. **添加供应商**：顶部「添加供应商」→ 填名称、地址、API Key → 可选展开「连接与模型」拉取并启用模型 → **保存并启用**
2. **切换上游**：列表中点目标供应商的 **启用**
3. **查看生效配置**：设置 → `config.toml` 编辑区（磁盘上只有一份生效配置；各供应商档案在本地 profiles 中）
4. **说明**：切换后**不会**结束已运行的 grok 会话；**新开**的 grok 会话才会读新 config

### Grok Auth 与号池

1. 打开 **设置 → Grok Auth JSON**，可导入单个 CPA `xai-*.json` 或 `~/.grok/auth.json`；该入口与下方自动巡检使用同一个号池。
2. 需要多账号时，在 **Grok 号池自动巡检** 中一次选择多个 JSON，或用“选择目录导入”递归读取目录及子目录中的全部 `.json`；原文件不会被移动。
3. 默认导入后立即巡检，之后每 6 小时自动巡检一次；可调整为 30–1440 分钟，并设置 1–16 并发。
4. 巡检确认权限拒绝、免费额度用尽或认证失效后，该账号会退出代理可用集合；普通 429/网络异常不会被误隔离。
5. 账号卡片会显示 HTTP 状态、错误码和具体探测错误；可批量禁用或删除所有“已巡检且非健康”的异常账号，待巡检账号不会被处理。
6. 自动巡检不会自动删除账号。手动禁用、启用、单个删除和批量操作均在设置页完成。
7. 无法直连 xAI 时，可填写 HTTP/HTTPS/SOCKS5 代理，例如 `http://127.0.0.1:7890`；该设置同时用于巡检、token 刷新和号池实时转发。

### 环境变量

| 变量 | 说明 |
|------|------|
| `GROK_CONFIG` | 指定 `config.toml` 完整路径 |
| `GROK_HOME` | 指定 `.grok` 目录，默认 `~/.grok` |
| `GROK_CLI` | 指定 Grok CLI 可执行文件；用于 macOS Finder 启动环境找不到终端 PATH 时 |

## 数据与安全

### 数据存在哪

| 路径 | 内容 |
|------|------|
| `~/.grok/config.toml` | Grok CLI **当前生效**配置 |
| `~/.grok_switch/profiles.json` | 供应商档案（**含 API Key 明文**） |
| `~/.grok_switch/backups/` | config 自动备份（**含 Key**） |
| `~/.grok_switch/settings.json` | 本工具设置 |
| `~/.grok_switch/grok_auth.json` | 单账号 xAI OAuth 凭据与本地代理 key（**敏感**） |
| `~/.grok_switch/grok_pool/pool.json` | 号池展示状态与巡检/代理设置（不含 token；代理 URL 可能包含认证信息） |
| `~/.grok_switch/grok_pool/accounts/` | 号池各账号 OAuth 凭据副本（**敏感**） |
| `~/.grok_switch/grok_switch.log` | 日志 |

macOS 登录项位于 `~/Library/LaunchAgents/com.grokbuildswitch.app.plist`。应用仍使用 `~/.grok_switch`，以保持现有数据格式兼容。

## 构建

### macOS 构建

- [Go](https://go.dev/dl/) **1.26+**
- Xcode Command Line Tools
- Apple Silicon Mac；脚本也支持在 Intel macOS runner 上交叉编译 arm64

```bash
./scripts/build-macos.sh
```

产物：

- `dist/Grok Build Switch.app`
- `dist/Grok-Build-Switch-macos-arm64.dmg`

构建会运行测试、组装 `.app`、执行 ad-hoc 签名并验证 DMG。仅在执行环境禁止 `hdiutil` 时，可用下面的命令只验证 `.app`：

```bash
SKIP_DMG=1 ./scripts/build-macos.sh
```

### Windows 构建

- Go 1.26+
- 可选：`rsrc`（嵌入 EXE 图标）、ImageMagick `magick`（从 SVG 生成 ICO）

```powershell
.\build.ps1
```

会运行测试并生成 `grok_switch.exe`。


## 开发

```bash
go test ./...
go run . -no-tray   # 仅 HTTP，无托盘（调试用）
```

### 文档站本地预览

文档站使用 MkDocs Material，内容位于 `docs/`：

```bash
uvx --with mkdocs-material mkdocs serve
```

打开终端输出的本地地址即可预览。提交到 `main` 后，GitHub Actions 会自动发布到 GitHub Pages。

主要目录：

```
main.go           # 入口
internal/         # 配置读写、供应商、HTTP、托盘
ui/               # Web 前端（嵌入应用二进制）
assets/           # 图标
docs/             # MkDocs 文档站
```

[使用教程](https://1parado.github.io/grok-build-switch/)

## 反馈群

欢迎加入grok build switch 反馈群

<img src="./QQ.jpg" alt="grok build switch 反馈群二维码" width="360">

## License

[MIT](./LICENSE)

## 友链
学AI 上L站！
[L站链接](https://linux.do/)
