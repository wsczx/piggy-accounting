# 开发环境搭建

## 前置要求

- [Go](https://golang.org/dl/) 1.25+
- [Node.js](https://nodejs.org/) 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.12+

### 安装 Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

验证安装：

```bash
wails version
```

## 克隆项目

```bash
git clone https://github.com/yourusername/piggy-accounting.git
cd piggy-accounting
```

## 开发模式

```bash
wails dev
```

这将启动：
- Vite 开发服务器（热重载）
- Wails 后端服务
- 桌面应用窗口

## 构建

### 开发构建

```bash
wails build
```

### 生产构建

```bash
wails build -platform darwin    # macOS
wails build -platform windows   # Windows
wails build -platform linux     # Linux
```

### 跨平台构建

```bash
# macOS 通用二进制（Intel + Apple Silicon）
wails build -platform darwin/universal

# Windows AMD64
wails build -platform windows/amd64

# Linux AMD64
wails build -platform linux/amd64
```

## 项目结构

```
piggy-accounting/
├── backend/              # Go 后端
│   ├── database/         # 数据库操作
│   ├── models/           # 数据模型
│   └── services/         # 业务逻辑
├── frontend/             # Vue 前端
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── stores/       # Pinia 状态管理
│   │   ├── views/        # 页面
│   │   └── utils/        # 工具函数
│   └── wailsjs/          # Wails 绑定代码
├── build/                # 构建资源
└── docs/                 # 文档
```

## 数据库

项目使用 SQLite 作为数据库，通过 XORM 进行 ORM 操作。

数据库文件位置：
- macOS: `~/.piggy-accounting/ledgers/`
- Windows: `%USERPROFILE%\.piggy-accounting\ledgers\`
- Linux: `~/.piggy-accounting/ledgers/`

## 调试

### 前端调试

开发模式下，可以访问 http://localhost:34115 在浏览器中调试前端代码。

### 后端调试

使用 VS Code 或 GoLand 配置调试器，启动 `wails dev` 后附加到进程。

### 日志

日志文件位置：
- macOS: `~/.piggy-accounting/logs/`
- Windows: `%USERPROFILE%\.piggy-accounting\logs\`
- Linux: `~/.piggy-accounting/logs/`

## 代码规范

### Go

```bash
# 格式化
go fmt ./...

# 检查
go vet ./...

# 测试
go test ./...
```

### TypeScript/Vue

```bash
cd frontend

# 检查
npm run lint

# 类型检查
npm run type-check
```

## 提交规范

提交信息格式：

```
<type>: <subject>

<body>

<footer>
```

类型：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

示例：

```
feat: 添加周期记账功能

支持日/周/月/年四种周期类型，自动执行到期记账任务。

Closes #123
```
