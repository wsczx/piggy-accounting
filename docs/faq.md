# 常见问题 (FAQ)

## 一般问题

### Q: 猪猪记账是免费的吗？

A: 是的，猪猪记账是完全免费的开源软件，采用 MIT 协议发布。

### Q: 我的数据存储在哪里？

A: 您的数据存储在本地 SQLite 数据库中：
- macOS: `~/.piggy-accounting/ledgers/`
- Windows: `%USERPROFILE%\.piggy-accounting\ledgers\`
- Linux: `~/.piggy-accounting/ledgers/`

### Q: 如何备份我的数据？

A: 您可以通过以下方式备份：
1. 打开应用，进入 个人页 > 备份
2. 点击"立即备份"创建手动备份
3. 备份文件会保存在 `~/.piggy-accounting/backups/` 目录

### Q: 支持哪些平台？

A: 目前支持：
- macOS 11+ (Intel 和 Apple Silicon)
- Windows 10/11
- Linux (Ubuntu 20.04+ 等)

## 功能问题

### Q: 如何设置周期记账？

A:
1. 进入 个人页 > 周期记账
2. 点击"新建周期记账"
3. 设置金额、分类、周期类型（日/周/月/年）
4. 保存后系统会自动按周期生成记账记录

### Q: 可以导入其他记账软件的数据吗？

A: 目前支持 CSV 格式导入。您可以将其他软件的数据导出为 CSV，然后通过 个人页 > 导入 功能导入。

### Q: 如何设置预算提醒？

A:
1. 进入 个人页 > 提醒设置
2. 找到"预算预警"
3. 设置预警阈值（如 80%）
4. 当支出达到预算的 80% 时，系统会发送提醒

### Q: Webhook 通知是什么？

A: Webhook 允许您将提醒发送到外部服务（如钉钉、企业微信、Discord 等）。在 个人页 > 提醒设置 中配置 Webhook URL 即可。

## 技术问题

### Q: 应用无法启动怎么办？

A: 请尝试以下步骤：
1. 检查系统要求是否满足
2. 重新下载安装包
3. 检查日志文件 `~/.piggy-accounting/logs/` 中的错误信息
4. 提交 Issue 并提供日志

### Q: 如何完全卸载应用？

A:
- **macOS**: 将应用拖到废纸篓，并删除 `~/.piggy-accounting/` 目录
- **Windows**: 卸载程序，并删除 `%USERPROFILE%\.piggy-accounting\` 目录
- **Linux**: 删除应用文件和 `~/.piggy-accounting/` 目录

### Q: 数据文件损坏了怎么办？

A:
1. 检查备份目录 `~/.piggy-accounting/backups/` 是否有可用备份
2. 使用最新的备份恢复数据
3. 如果没有备份，可以尝试使用 SQLite 工具修复数据库文件

## 其他问题

### Q: 如何参与项目开发？

A: 欢迎参与！请阅读 [CONTRIBUTING.md](../CONTRIBUTING.md) 了解如何贡献代码。

### Q: 发现了 Bug 怎么办？

A: 请在 [GitHub Issues](https://github.com/yourusername/piggy-accounting/issues) 提交 Bug 报告，并尽可能提供详细信息。

### Q: 有新功能建议？

A: 欢迎在 [GitHub Issues](https://github.com/yourusername/piggy-accounting/issues) 提交功能建议，或参与讨论。

---

还有其他问题？请发送邮件至: wsc@wsczx.com
