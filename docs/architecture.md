# 架构说明

## 技术栈

猪猪记账采用前后端分离的桌面应用架构：

- **后端**: Go 1.25 + Wails v2
- **前端**: Vue 3.4 + TypeScript + Tailwind CSS v4
- **数据库**: SQLite (modernc.org/sqlite)
- **ORM**: XORM v1.3
- **状态管理**: Pinia
- **路由**: Vue Router

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      猪猪记账应用                            │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Vue 3 UI   │  │  Pinia Store │  │ Vue Router   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         └─────────────────┴─────────────────┘               │
│                           │                                 │
│                    Wails Bridge                           │
│                           │                                 │
│  ┌────────────────────────┴────────────────────────┐       │
│  │                    Go Backend                    │       │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐        │       │
│  │  │  Services│ │ Database │ │  Models  │        │       │
│  │  └──────────┘ └──────────┘ └──────────┘        │       │
│  └─────────────────────────────────────────────────┘       │
│                           │                                 │
│                      SQLite DB                            │
└─────────────────────────────────────────────────────────────┘
```

## 后端架构

### 目录结构

```
backend/
├── database/          # 数据访问层
│   ├── database.go    # 数据库初始化
│   ├── record_repo.go # 记账记录操作
│   ├── category_repo.go # 分类操作
│   ├── account_repo.go  # 账户操作
│   ├── transfer_repo.go # 转账操作
│   ├── recurring_repo.go # 周期记账
│   ├── task_repo.go     # 待办任务
│   ├── budget_repo.go   # 预算管理
│   ├── reminder_repo.go # 提醒系统
│   ├── tag_repo.go      # 标签管理
│   ├── backup_repo.go   # 备份恢复
│   └── ledger_repo.go   # 多账本管理
├── models/            # 数据模型
│   └── models.go      # 所有模型定义
└── services/          # 服务层
    └── services.go    # 服务入口
```

### 核心模型

```go
// Record - 记账记录
type Record struct {
    ID          int64
    Type        string    // expense/income
    Amount      float64
    Category    string
    AccountID   int64
    Note        string
    Date        time.Time
    Tags        []Tag     `xorm:"-"`
}

// Account - 账户
type Account struct {
    ID          int64
    Name        string
    Icon        string
    Balance     float64   // 实时计算
    IsDefault   bool
}

// RecurringRecord - 周期记账
type RecurringRecord struct {
    ID          int64
    Type        string    // daily/weekly/monthly/yearly
    Amount      float64
    Category    string
    AccountID   int64
    Frequency   string
    WeekDay     int       // 每周几
    MonthDay    int       // 每月几号
    NextDate    time.Time
    Enabled     bool
}
```

### 数据库设计

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   records    │     │  categories  │     │   accounts   │
├──────────────┤     ├──────────────┤     ├──────────────┤
│ id (PK)      │     │ id (PK)      │     │ id (PK)      │
│ type         │────▶│ name         │     │ name         │
│ amount       │     │ icon         │     │ icon         │
│ category     │     │ type         │     │ is_default   │
│ account_id   │────▶│ sort_order   │     │ sort_order   │
│ note         │     └──────────────┘     └──────────────┘
│ date         │
│ created_at   │
└──────────────┘
       │
       │    ┌──────────────┐
       └───▶│ record_tags  │
            ├──────────────┤
            │ record_id    │
            │ tag_id       │
            └──────────────┘
```

## 前端架构

### 目录结构

```
frontend/src/
├── components/        # 公共组件
│   ├── TabBar.vue     # 底部导航
│   ├── RecordPanel.vue # 记账面板
│   ├── RecordList.vue  # 记录列表
│   └── ...
├── views/             # 页面视图
│   ├── HomeView.vue      # 首页
│   ├── StatisticsView.vue # 统计
│   ├── CategoriesView.vue # 分类
│   ├── TasksView.vue      # 任务
│   └── ProfileView.vue    # 个人页
├── stores/            # Pinia 状态管理
│   ├── theme.ts       # 主题管理
│   └── accounting.ts  # 业务数据
├── router/            # 路由配置
│   └── index.ts
└── utils/             # 工具函数
    └── formatters.ts  # 格式化工具
```

### 状态管理

```typescript
// stores/accounting.ts
export const useAccountingStore = defineStore('accounting', () => {
  // State
  const records = ref<Record[]>([])
  const categories = ref<Category[]>([])
  const accounts = ref<Account[]>([])
  const totalAssets = ref(0)
  
  // Getters
  const monthlyStats = computed(() => {...})
  
  // Actions
  const loadRecords = async () => {...}
  const addRecord = async (data: RecordData) => {...}
  
  return {
    records, categories, accounts,
    monthlyStats,
    loadRecords, addRecord
  }
})
```

### 组件通信

- **Props/Events**: 父子组件通信
- **Pinia**: 全局状态管理
- **Wails Events**: 后端推送事件（提醒通知）

## 数据流

### 记账流程

```
用户操作 → RecordPanel.vue
              ↓
         accountingStore.addRecord()
              ↓
         Wails: AddRecord()
              ↓
         backend: record_repo.go
              ↓
         SQLite INSERT
              ↓
         返回结果 → 更新 UI
```

### 提醒流程

```
App 启动 → startReminderTicker()
              ↓
         定时检查（每分钟）
              ↓
         reminder_repo.go: CheckReminders()
              ↓
         触发条件满足？
              ↓
    是 ──▶ Wails: EventsEmit()
              ↓
         App.vue: EventsOn()
              ↓
         显示通知
```

## 构建流程

```
开发阶段:
wails dev → Vite Dev Server + Go 编译 → 热重载

生产构建:
wails build → 
  1. npm run build (Vite 构建前端)
  2. go build (编译 Go 后端)
  3. 打包资源 → 可执行文件
```

## 跨平台支持

### 平台差异处理

```go
// app.go
func (a *App) GetPlatform() string {
    return runtime.GOOS  // darwin/windows/linux
}
```

### 数据目录

```go
func getDataDir() string {
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(os.Getenv("HOME"), ".piggy-accounting")
    case "windows":
        return filepath.Join(os.Getenv("USERPROFILE"), ".piggy-accounting")
    default:
        return filepath.Join(os.Getenv("HOME"), ".piggy-accounting")
    }
}
```

## 安全考虑

1. **数据存储**: 本地 SQLite，不上传云端
2. **备份安全**: 恢复前自动创建安全备份
3. **路径安全**: 备份/恢复操作进行路径检查，防止目录穿越
4. **并发安全**: 数据库操作使用互斥锁保护

## 性能优化

1. **数据库**: 使用 WAL 模式提升并发性能
2. **前端**: 虚拟滚动处理大量记录
3. **查询**: 分页加载，避免一次性加载过多数据
4. **缓存**: Pinia Store 缓存常用数据
