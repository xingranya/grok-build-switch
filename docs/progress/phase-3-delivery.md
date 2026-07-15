# Phase 3：验证与 fork 交付

**目标**：在真实 Apple Silicon macOS 环境完成验证并同步个人 fork。  
**状态**：进行中

## 任务

- [x] **T3.1：建立工具链并执行完整验证**
  - 优先级：P0
  - 工作量：M
  - 测试要求：`go test ./...`、`go vet ./...`、macOS 构建与产物静态检查
  - 记忆影响：记录实际工具链和验证命令
  - 验收：全部自动化验证通过；不启动非必要 GUI 进程
  - 备注：已完成；Go test/vet、Windows amd64 编译、MkDocs strict、arm64 `.app`、ad-hoc 签名和 Actions DMG 验证通过
- [ ] **T3.2：将适配结果同步到个人 fork**
  - 优先级：P1
  - 工作量：S
  - 测试要求：核对远端提交树、分支和关键文件；不适用业务测试
  - 记忆影响：记录只读 `.git` 下的交付方式
  - 验收：个人 fork 存在 macOS 适配提交；没有向上游创建 PR
  - 备注：进行中；已通过 `/tmp` 可写克隆推送首个适配提交，正在同步最终验证状态

## 阶段备注

- 个人 fork：`xingranya/grok-build-switch`。
- 上游：`1parado/grok-build-switch`。

## 阶段完成清单

- [ ] 上述任务全部完成
- [ ] `MASTER.md` 全部阶段计数已更新
- [ ] 进入归档阶段
