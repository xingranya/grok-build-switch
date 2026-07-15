# 里程碑

| 序号 | 里程碑 | 阶段 | 完成标准 | 状态 |
|:---|:---|:---|:---|:---|
| M1 | macOS 系统能力接通 | Phase 1 | LaunchAgent、CLI 定位、状态一致性和菜单栏资源均实现并有测试 | 已完成 |
| M2 | 可安装产物生成 | Phase 2 | 单命令和 GitHub macOS runner 均可生成 arm64 `.app` 与 DMG，完成 ad-hoc 签名并更新文档 | 已完成 |
| M3 | 真实验证与 fork 交付 | Phase 3 | Go 测试、vet、构建和产物检查通过，代码同步至个人 fork | 已完成 |

## 范围外项目

- Intel `amd64` 或 Universal 2 产物。
- Apple Developer ID 正式签名、公证和 stapling。
- Mac App Store 沙盒化或发布。
- 与 macOS 适配无关的 server/UI 大规模重构。
- 向上游 `1parado/grok-build-switch` 创建 Pull Request。
