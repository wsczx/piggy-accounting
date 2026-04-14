<div align="center">

# 🐷 猪猪记账 (Piggy Accounting)

[![Wails](https://img.shields.io/badge/Wails-v2.12-ff4d6d?logo=wails)](https://wails.io)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vue.js)](https://vuejs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**一款精美的跨平台个人记账桌面应用，让财务管理轻松又愉快！**

</div>

---

## 📖 简介

猪猪记账是一款基于 [Wails](https://wails.io) 框架开发的跨平台桌面记账应用，采用 Go + Vue 3 + TypeScript + SQLite 技术栈，支持 macOS 和 Windows。

应用采用粉紫渐变品牌设计，无边框窗口 + 圆角玻璃质感，macOS/Windows 双平台适配，支持浅色与深色主题切换。

## ✨ 功能特性

### 💰 核心记账
- **收支记录** — 快速记录日常收入和支出，底部弹出面板一键记账
- **分类管理** — 19 个系统预设类别 + 自定义类别，图标和颜色标识
- **标签系统** — 为记录添加标签，便于分类检索
- **智能识别** — 自然语言文本自动解析为记账记录（如"午餐 25"）
- **搜索与筛选** — 按关键词、类型、类别筛选，支持无限滚动分页

### 📊 统计分析
- **月度统计** — 本月收支概览 + 分类支出占比
- **年度统计** — 年度总览 + 月度趋势对比图 + 年度分类排行
- **结余卡片** — 首页实时显示本月结余、收支总额、预算进度（阶段式变色）、本周支出柱状图 + TOP3 分类

### 💳 账户管理
- **多账户** — 支持现金、微信、支付宝、银行卡等，自动计算账户余额
- **转账记录** — 记录账户间转账，自动生成对应的支出/收入记录

### 📅 自动化
- **周期记账** — 支持每日/每周/每月/每年周期自动记账，启动时自动执行到期记录
- **智能提醒** — 预算预警、每日记账提醒、每周汇总、任务到期提醒，支持 Webhook 推送

### 📋 任务管理
- **待办任务** — 主页快速面板 + 独立任务页面，支持创建、完成、编辑、删除

### 📚 多账本
- **多账本支持** — 创建多个独立账本，一键切换，支持账本导入导出

### 💾 数据安全
- **备份恢复** — 自动/手动备份，一键恢复，支持导出到桌面
- **数据导入** — 标准 CSV 导入 + 微信账单导入 + 支付宝账单导入
- **数据导出** — CSV 格式导出
- **数据管理** — 一键清空当前账本数据（保留系统类别和默认账户）

### 🎨 界面设计
- **深色模式** — 浅色/深色主题自由切换
- **无边框窗口** — 自定义标题栏 + macOS 交通灯按钮
- **品牌图标** — 粉紫渐变背景 + 🐷 emoji
- **动画交互** — 流畅的面板弹出、进度条动画

## 🏗️ 技术架构

### 后端 (Go)

采用分层架构，职责清晰分离：

```
backend/
├── app.go              # Wails 生命周期 + 基础设施方法
├── main.go             # 入口 + 服务绑定（13 个 Service + App）
├── base/               # 日志系统
├── models/             # 数据模型（Record, Category, Budget 等）
├── dbdata/             # 纯数据访问层
│   ├── db.go           # 数据库初始化 + 系统类别/默认账户种子
│   ├── db_testdata.go  # 首次安装测试数据（覆盖多月份）
│   ├── db_orm.go       # 通用 CRUD 泛型封装
│   └── ledger.go       # 账本基础设施（配置管理/路径解析）
└── service/            # 业务逻辑层（14 个模块）
    ├── service.go      # 服务入口 + 通用工具函数
    ├── record.go       # 记录 CRUD + 搜索分页 + 统计查询
    ├── category.go     # 类别 CRUD + 图标查询
    ├── budget.go       # 预算 CRUD + 预算信息查询
    ├── account.go      # 账户管理 + 余额计算
    ├── tag.go          # 标签 CRUD + RecordTag 关联
    ├── task.go         # 待办任务 CRUD
    ├── transfer.go     # 转账 CRUD（自动生成收支记录）
    ├── recurring.go    # 周期记账 CRUD + 自动执行
    ├── reminder.go     # 提醒业务逻辑 + Webhook 通知
    ├── smart_recognize.go  # 智能文本解析
    ├── export_import.go    # CSV/微信/支付宝导入导出
    ├── backup.go       # 备份/恢复/清空
    └── ledger.go       # 多账本 CRUD + 切换 + 迁移
```

- **ORM**: [xorm](https://xorm.io) v1.3.11
- **数据库**: SQLite（纯 Go 驱动 `modernc.org/sqlite`，无需 CGO）
- **存储**: `~/.piggy-accounting/ledgers/` 目录，每个账本一个 `.db` 文件

### 前端 (Vue 3)

```
frontend/src/
├── App.vue                 # 应用入口 + 全局通知系统
├── main.ts                 # 前端入口
├── style.css               # CSS 变量主题系统 + 布局骨架 + 全局动画
├── router/
│   └── index.ts            # 路由（首页/统计/任务/账单/我的）
├── stores/
│   ├── theme.ts            # 主题管理（亮色/暗色）
│   └── accounting.ts       # 核心业务数据管理
├── views/
│   ├── HomeView.vue        # 主页（结余卡片 + 预算进度 + 本周柱状图 + 记录列表 + 任务面板）
│   ├── StatisticsView.vue  # 统计分析页 + 预算管理
│   ├── RecordsView.vue     # 全部账单页
│   ├── TasksView.vue       # 任务管理页
│   └── ProfileView.vue     # 个人页（功能入口 + 14 个弹窗调度）
├── components/
│   ├── BalanceCard.vue     # 结余卡片（收支 + 预算进度 + 本周柱状图 + TOP3）
│   ├── RecordPanel.vue     # 底部弹出记账面板
│   ├── RecordList.vue      # 记录列表（搜索 + 筛选 + 无限滚动）
│   ├── QuickActions.vue    # 快捷入口
│   ├── TabBar.vue          # 底部导航栏
│   ├── LineChart.vue       # 折线图组件
│   ├── PieChart.vue        # 饼图组件
│   ├── ConfirmModal.vue    # 确认弹窗
│   ├── FilterModal.vue     # 筛选弹窗
│   ├── TaskForm.vue        # 任务表单
│   └── modals/             # 15 个独立弹窗组件
│       ├── AccountModal.vue        # 账户管理
│       ├── BackupModal.vue         # 备份恢复
│       ├── BudgetModal.vue         # 预算设置
│       ├── CategoryModal.vue       # 类别管理
│       ├── DataManageModal.vue     # 数据管理（清空）
│       ├── ExportModal.vue         # 数据导出
│       ├── HelpFeedbackModal.vue   # 帮助与反馈
│       ├── ImportModal.vue         # 数据导入
│       ├── LedgerModal.vue         # 账本管理
│       ├── RecurringModal.vue      # 周期记账
│       ├── ReminderModal.vue       # 提醒设置
│       ├── SecuritySettingsModal.vue # 安全设置
│       ├── TagModal.vue            # 标签管理
│       ├── TransferModal.vue       # 转账记录
│       └── AboutModal.vue          # 关于
└── utils/
    ├── formatters.ts       # 格式化工具（金额/日期）
    ├── category.ts         # 类别共享工具（图标/颜色）
    ├── date.ts             # 日期共享工具
    └── logger.ts           # 日志工具（前端 + 后端桥接）
```

- **状态管理**: [Pinia](https://pinia.vuejs.org)
- **路由**: [Vue Router](https://router.vuejs.org)
- **CSS**: 自定义 CSS 变量主题系统 + Tailwind CSS v4
- **布局**: 纯 CSS 布局骨架（`.app-container` > `.app-header` + `.app-main` + `.app-tabbar`）

## 🚀 快速开始

### 环境要求

- [Go](https://golang.org/dl/) 1.25+
- [Node.js](https://nodejs.org/) 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.12+

### 安装与运行

```bash
# 克隆仓库
git clone https://github.com/yourusername/piggy-accounting.git
cd piggy-accounting

# 开发模式运行
wails dev

# 构建应用
wails build
```

### 首次启动

首次安装时会自动创建 `~/.piggy-accounting/` 目录，生成 `默认账本.db` 并写入测试数据（覆盖 1-4 月共 70 条记录），方便体验各功能。

### 下载预编译版本

前往 [Releases](https://github.com/yourusername/piggy-accounting/releases) 页面下载对应平台的安装包。

## 📁 数据目录

```
~/.piggy-accounting/
├── ledgers/                  # 账本数据库
│   ├── 默认账本.db
│   └── ...
├── backups/                  # 备份文件
│   ├── manual_YYYYMMDD_HHmmss.db
│   └── auto_YYYYMMDD_HHmmss.db
├── ledger_config.json        # 账本配置（当前账本 + 文件名映射）
└── logs/                     # 日志文件
    └── YYYY-MM-DD.log
```

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request！

## 📜 开源协议

本项目采用 [MIT 协议](LICENSE) 开源。

## 🙏 致谢

- [Wails](https://wails.io) — Go + Web 桌面应用框架
- [Vue.js](https://vuejs.org) — 渐进式 JavaScript 框架
- [xorm](https://xorm.io) — Go ORM 库
- [modernc.org/sqlite](https://modernc.org/sqlite) — 纯 Go SQLite 驱动

---

<div align="center">

**Made with ❤️ by [孤鸿](mailto:wsc@wsczx.com)**

</div>
