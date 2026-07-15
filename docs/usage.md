# 使用教程

## 安装Grok Build

[Grok Build 安装链接](https://grok.com/build) 

如果有supergrok账号请在终端直接grok login，不需要使用grok build switch，该项目适合随时切换上游供应商的情况。

## Grok Build Switch

### 1.下载并运行

- macOS Apple Silicon：解压 `Grok-Build-Switch-macos-arm64-cli-v0.4.4.zip`，保留完整文件夹并双击 `启动.command`。
- Windows x64：下载并运行 `grok_switch.exe`。

macOS 可执行文件使用 ad-hoc 签名。首次打开如被拦截，请按住 Control 点击 `.command` 文件，再选择“打开”。

![Windows 托盘状态](assets/images/grok_1.png)

###  2.打开本地 Web 面板

![打开本地 Web 面板](assets/images/grok_2.png)

###  3.添加供应商 Profile 

![添加供应商](assets/images/grok_3.png)

###  4.拉取并选择模型

![模型配置](assets/images/grok_4.png)

![模型高级字段配置](assets/images/grok_5.png)

###  5.保存并启用供应商

![模型高级字段配置](assets/images/grok_6.png)

###  6.检查 Grok CLI 配置是否生效

在终端输入grok，启动grok然后进行问答

![模型高级字段配置](assets/images/grok_7.png)

###  7.使用 Windows 托盘快速切换
###  8.备份和还原配置

macOS 不创建 Dock 或菜单栏图标。双击 `启动.command` 可启动服务或重新打开管理页面；运行期间保留终端窗口并按需最小化。双击 `停止.command` 或在“设置 → 服务”中点击“停止服务”即可结束服务和终端窗口。

## macOS 登录时启动

在设置页开启“开机自启”后，应用会写入：

```text
~/Library/LaunchAgents/com.grokbuildswitch.app.plist
```

登录项会在下次登录时静默启动应用。关闭后会删除该文件。

## macOS 找不到 Grok CLI

从 Finder 启动的应用不会继承终端的完整 PATH。程序会自动检查 Homebrew、Volta 和常见用户安装目录；仍找不到时，可在启动应用前设置 `GROK_CLI` 为完整可执行文件路径。
