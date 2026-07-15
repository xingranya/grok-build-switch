# 记忆面归档说明

本次任务没有可写的项目原生记忆，也未获用户授权创建仓库内回退记忆文件。因此没有写入持久记忆；稳定技术决策保存在本归档的分析、计划和进度文档中。

关键决策：

- macOS 应用标识和 LaunchAgent label 固定为 `com.grokbuildswitch.app`。
- 用户数据继续保存在 `~/.grok_switch`，不迁移到 `Application Support`。
- Finder 环境下的 Grok CLI 按 `GROK_CLI`、PATH、macOS 常见目录依次解析。
- 当前交付范围为 Apple Silicon arm64、ad-hoc 签名，不包含 Universal 2 和 Apple 公证。
