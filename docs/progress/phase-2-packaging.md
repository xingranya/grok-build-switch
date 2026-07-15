# Phase 2：应用打包与使用文档

**目标**：生成可双击安装和运行的 Apple Silicon `.app` 与 DMG。  
**状态**：已完成

## 任务

- [x] **T2.1：新增 macOS 应用与磁盘映像构建**
  - 优先级：P0
  - 工作量：L
  - 测试要求：shell 语法、plist、arm64 架构、应用结构、签名和 DMG 验证
  - 记忆影响：记录标准构建命令
  - 验收：单命令生成 `.app` 与 DMG，并完成 ad-hoc 签名
  - 备注：已完成实现；本地 `.app`、arm64、plist 和 ad-hoc 签名验证通过；沙箱阻止 `hdiutil`，真实 DMG 转由 T2.3 验证
- [x] **T2.2：更新 README 与用户文档**
  - 优先级：P1
  - 工作量：M
  - 测试要求：文档命令、路径和链接静态核对；纯文档无需业务测试
  - 记忆影响：无
  - 验收：安装、Gatekeeper、自启动、数据路径和源码构建说明与真实行为一致
  - 备注：已完成；README、文档首页、使用教程和仓库链接已更新；MkDocs `--strict` 构建通过
- [x] **T2.3：新增 GitHub macOS 构建产物工作流**
  - 优先级：P1
  - 工作量：S
  - 测试要求：YAML 语法、工作流触发和远端 artifact 验证
  - 记忆影响：记录沙箱内 `hdiutil` 限制
  - 验收：macOS runner 生成并上传 `.app` 与 DMG；本地可用 `SKIP_DMG=1` 验证 `.app`
  - 备注：已完成；Actions run `29434386031` 成功，远端验证 arm64、plist、ad-hoc 签名和 DMG，artifact 已下载

## 阶段备注

- 范围限定为 arm64，不生成 Universal 2。
- 使用 ad-hoc 签名，不要求 Developer ID 或公证凭据。
- T2.1 因沙箱内 `hdiutil` 失败产生 1 点漂移，达到重规划阈值；新增 T2.3，剩余段 drift_score 已重置。
- T2.3 实际工作量增加 1 级，再次达到剩余段重规划阈值；本阶段已无待办，不再拆分，直接进入 Phase 3。

## 阶段完成清单

- [x] 上述任务全部完成
- [x] `MASTER.md` 阶段计数已更新
- [x] `MASTER.md` 当前状态已切换到 Phase 3
