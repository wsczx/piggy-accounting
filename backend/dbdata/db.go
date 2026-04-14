package dbdata

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"piggy-accounting/backend/models"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

// ORM 全局 xorm 引擎
var ORM *xorm.Engine

// dbMutex 保护 ORM 的 Close/Init 切换过程，防止并发访问 panic
var dbMutex sync.RWMutex

// NewSession 创建新的数据库 session
func NewSession() *xorm.Session {
	return ORM.NewSession()
}

// Init 初始化数据库（根据 ledger_config.json 决定打开哪个账本）
// 如果配置文件不存在，自动创建默认账本（"默认账本.db"）
// 返回 hasTestData 表示本次是否写入了测试数据（仅首次创建新账本时为 true）
func Init() (bool, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	return initInternal()
}

// initInternal 实际初始化逻辑（由 Init 调用，已持写锁）
func initInternal() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("获取用户目录失败: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".piggy-accounting")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return false, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 确保 ledger_config.json 存在且指向有效的账本文件
	hasTestData, err := ensureDefaultLedger(dataDir)
	if err != nil {
		return false, fmt.Errorf("初始化账本配置失败: %w", err)
	}

	// 获取当前活跃账本的数据库路径
	config, err := LoadLedgerConfig()
	if err != nil {
		return false, fmt.Errorf("加载账本配置失败: %w", err)
	}
	dbPath, err := getActiveDBPath(config, dataDir)
	if err != nil {
		return false, err
	}

	ORM, err = xorm.NewEngine("sqlite", dbPath)
	if err != nil {
		return false, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 设置日志
	ormLogger := log.NewSimpleLogger(os.Stdout)
	ormLogger.ShowSQL(true)
	ORM.SetLogger(log.NewLoggerAdapter(ormLogger))

	// 启用 WAL 模式（Write-Ahead Logging），提升并发读写性能
	if _, err := ORM.Exec("PRAGMA journal_mode=WAL"); err != nil {
		fmt.Printf("[WARN] 启用 WAL 模式失败: %v\n", err)
	}

	ORM.SetMaxIdleConns(5)
	ORM.SetMaxOpenConns(10)

	// 同步表结构（AutoMigrate）
	if err := ORM.Sync2(
		new(models.Record),
		new(models.Category),
		new(models.Budget),
		new(models.Tag),
		new(models.RecordTag),
		new(models.Reminder),
		new(models.RecurringRecord),
		new(models.Account),
		new(models.Transfer),
		new(models.Task),
		new(models.ReminderSettings),
		new(models.TaskReminder),
	); err != nil {
		return false, fmt.Errorf("同步表结构失败: %w", err)
	}

	// 初始化系统类别
	if err := seedCategories(); err != nil {
		return false, fmt.Errorf("初始化类别失败: %w", err)
	}

	// 初始化默认账户
	if err := seedDefaultAccount(); err != nil {
		return false, fmt.Errorf("初始化默认账户失败: %w", err)
	}

	return hasTestData, nil
}

// ensureDefaultLedger 确保 ledger_config.json 存在，且指向有效的默认账本
// 如果没有配置文件，创建默认配置和默认账本数据库
// 返回 hasTestData 表示是否写入了测试数据（仅新创建账本时为 true）
func ensureDefaultLedger(dataDir string) (hasTestData bool, err error) {
	ledgerDir := filepath.Join(dataDir, "ledgers")
	if err := os.MkdirAll(ledgerDir, 0755); err != nil {
		return false, err
	}

	config, err := LoadLedgerConfig()
	if err != nil {
		return false, err
	}

	// 检查当前活跃账本对应的数据库文件是否存在
	if config.CurrentLedger != "" {
		filename := resolveLedgerFilename(config.CurrentLedger, config)
		dbPath := filepath.Join(ledgerDir, filename)
		if _, err := os.Stat(dbPath); err == nil {
			return false, nil // 配置和文件都存在，直接返回
		}
	}

	// 配置不存在或指向的文件不存在 → 创建默认账本
	defaultName := "默认账本"
	defaultFilename := defaultName + ".db"
	defaultPath := filepath.Join(ledgerDir, defaultFilename)

	// 如果默认账本文件也不存在，创建空数据库并初始化表结构 + 测试数据
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		eng, err := xorm.NewEngine("sqlite", defaultPath)
		if err != nil {
			return false, fmt.Errorf("创建默认账本数据库失败: %w", err)
		}
		if err := eng.Sync2(
			new(models.Record),
			new(models.Category),
			new(models.Budget),
			new(models.Tag),
			new(models.RecordTag),
			new(models.Reminder),
			new(models.RecurringRecord),
			new(models.Account),
			new(models.Transfer),
			new(models.Task),
			new(models.ReminderSettings),
			new(models.TaskReminder),
		); err != nil {
			eng.Close()
			os.Remove(defaultPath)
			return false, fmt.Errorf("初始化默认账本表结构失败: %w", err)
		}
		// 写入测试数据（类别 + 默认账户 + 示例记录）
		if err := seedTestData(eng); err != nil {
			fmt.Printf("[WARN] 写入测试数据失败: %v\n", err)
		}
		eng.Close()
		hasTestData = true
	}

	// 保存配置
	config.CurrentLedger = defaultName
	if config.NameMap == nil {
		config.NameMap = make(map[string]string)
	}
	config.NameMap[defaultFilename] = defaultName
	if err := SaveLedgerConfig(config); err != nil {
		return false, err
	}

	return hasTestData, nil
}

// getActiveDBPath 根据配置获取当前活跃账本的数据库路径
func getActiveDBPath(config *LedgerConfig, dataDir string) (string, error) {
	ledgerDir := filepath.Join(dataDir, "ledgers")
	filename := resolveLedgerFilename(config.CurrentLedger, config)
	dbPath := filepath.Join(ledgerDir, filename)

	if _, err := os.Stat(dbPath); err != nil {
		return "", fmt.Errorf("账本数据库文件不存在: %s", dbPath)
	}
	return dbPath, nil
}

// Close 关闭数据库连接
func Close() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if ORM != nil {
		return ORM.Close()
	}
	return nil
}

// GetCurrentDBPath 获取当前活跃账本的数据库路径（供 backup 等使用）
func GetCurrentDBPath() (string, error) {
	config, err := LoadLedgerConfig()
	if err != nil {
		return "", err
	}
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	return getActiveDBPath(config, dataDir)
}

// seedCategories 初始化系统默认类别
func seedCategories() error {
	count, err := ORM.Where("is_system = ?", true).Count(new(models.Category))
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []models.Category{
		{Name: "餐饮", Icon: "🍜", Type: "expense", IsSystem: true},
		{Name: "交通", Icon: "🚗", Type: "expense", IsSystem: true},
		{Name: "购物", Icon: "🛒", Type: "expense", IsSystem: true},
		{Name: "娱乐", Icon: "🎮", Type: "expense", IsSystem: true},
		{Name: "住房", Icon: "🏠", Type: "expense", IsSystem: true},
		{Name: "医疗", Icon: "💊", Type: "expense", IsSystem: true},
		{Name: "教育", Icon: "📚", Type: "expense", IsSystem: true},
		{Name: "通讯", Icon: "📱", Type: "expense", IsSystem: true},
		{Name: "服饰", Icon: "👔", Type: "expense", IsSystem: true},
		{Name: "宠物", Icon: "🐱", Type: "expense", IsSystem: true},
		{Name: "运动", Icon: "⚽", Type: "expense", IsSystem: true},
		{Name: "其他支出", Icon: "📦", Type: "expense", IsSystem: true},
		{Name: "工资", Icon: "💰", Type: "income", IsSystem: true},
		{Name: "奖金", Icon: "🎁", Type: "income", IsSystem: true},
		{Name: "投资", Icon: "📈", Type: "income", IsSystem: true},
		{Name: "兼职", Icon: "💼", Type: "income", IsSystem: true},
		{Name: "红包", Icon: "🧧", Type: "income", IsSystem: true},
		{Name: "退款", Icon: "💳", Type: "income", IsSystem: true},
		{Name: "其他收入", Icon: "💎", Type: "income", IsSystem: true},
	}

	session := ORM.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	for i := range categories {
		if _, err := session.Insert(&categories[i]); err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}

// seedDefaultAccount 初始化默认账户（如果没有任何账户）
func seedDefaultAccount() error {
	count, err := ORM.Count(new(models.Account))
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	defaultAccount := &models.Account{
		Name:      "现金",
		Icon:      "💵",
		Balance:   0,
		IsDefault: true,
		SortOrder: 0,
	}
	_, err = ORM.Insert(defaultAccount)
	return err
}
