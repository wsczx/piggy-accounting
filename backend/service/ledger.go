package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"

	"xorm.io/xorm"
)

// LedgerService 多账本业务逻辑层
type LedgerService struct{}

// NewLedgerService 创建账本服务实例
func NewLedgerService() *LedgerService {
	return &LedgerService{}
}

// GetAllLedgers 获取所有账本
func (s *LedgerService) GetAllLedgers() ([]models.LedgerInfo, error) {
	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return nil, err
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(ledgerDir)
	if err != nil {
		return nil, err
	}

	var ledgers []models.LedgerInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 从文件名反推账本名（优先使用 NameMap 精确还原）
		ledgerName := resolveLedgerName(entry.Name(), config)

		// 查询记录数量
		dbPath := filepath.Join(ledgerDir, entry.Name())
		recordCount := getRecordCountFromDB(dbPath)

		ledgers = append(ledgers, models.LedgerInfo{
			Name:        ledgerName,
			IsActive:    ledgerName == config.CurrentLedger,
			RecordCount: recordCount,
			CreatedAt:   info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	// 按名称排序，当前活跃的排前面
	sort.Slice(ledgers, func(i, j int) bool {
		if ledgers[i].IsActive != ledgers[j].IsActive {
			return ledgers[i].IsActive
		}
		return ledgers[i].Name < ledgers[j].Name
	})

	return ledgers, nil
}

// CreateLedger 创建新账本
func (s *LedgerService) CreateLedger(name string) (*models.LedgerInfo, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("账本名称不能为空")
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return nil, err
	}

	// 检查是否已存在
	filename := dbdata.LedgerNameToFilename(name)
	dbPath := filepath.Join(ledgerDir, filename)
	if _, err := os.Stat(dbPath); err == nil {
		return nil, fmt.Errorf("账本已存在")
	}

	// 创建数据库
	eng, err := xorm.NewEngine("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	// 初始化表结构
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
		os.Remove(dbPath)
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	// 初始化默认数据
	if err := seedCategoriesInEngine(eng); err != nil {
		eng.Close()
		os.Remove(dbPath)
		return nil, fmt.Errorf("初始化类别失败: %w", err)
	}
	if err := seedDefaultAccountInEngine(eng); err != nil {
		eng.Close()
		os.Remove(dbPath)
		return nil, fmt.Errorf("初始化默认账户失败: %w", err)
	}
	if err := Reminders.InitDefaultRemindersInEngine(eng); err != nil {
		eng.Close()
		os.Remove(dbPath)
		return nil, fmt.Errorf("初始化提醒设置失败: %w", err)
	}

	eng.Close()

	// 注册账本显示名到配置
	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return nil, err
	}
	config.NameMap[filename] = name
	if err := dbdata.SaveLedgerConfig(config); err != nil {
		fmt.Printf("[WARN] 注册账本名称失败: %v\n", err)
	}

	return &models.LedgerInfo{
		Name:        name,
		IsActive:    false,
		RecordCount: 0,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// SwitchLedger 切换当前账本
func (s *LedgerService) SwitchLedger(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("账本名称不能为空")
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return err
	}

	// 查找账本文件
	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return err
	}

	filename := dbdata.ResolveLedgerFilename(name, config)
	dbPath := filepath.Join(ledgerDir, filename)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("账本不存在: %s", name)
	}

	// 关闭当前数据库连接
	if err := dbdata.Close(); err != nil {
		fmt.Printf("[WARN] 关闭数据库连接失败: %v\n", err)
	}

	// 更新配置
	config.CurrentLedger = name
	if err := dbdata.SaveLedgerConfig(config); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// 重新初始化数据库
	_, err = dbdata.Init()
	return err
}

// DeleteLedger 删除账本
func (s *LedgerService) DeleteLedger(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("账本名称不能为空")
	}

	// 不允许删除当前活跃账本
	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return err
	}
	if config.CurrentLedger == name {
		return fmt.Errorf("不能删除当前使用的账本，请先切换到其他账本")
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return err
	}

	filename := dbdata.ResolveLedgerFilename(name, config)
	dbPath := filepath.Join(ledgerDir, filename)

	// 删除数据库文件及其辅助文件
	for _, filePath := range []string{dbPath, dbPath + "-shm", dbPath + "-wal"} {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除文件失败 %s: %w", filepath.Base(filePath), err)
		}
	}

	// 从 NameMap 中移除
	delete(config.NameMap, filename)
	return dbdata.SaveLedgerConfig(config)
}

// RenameLedger 重命名账本
func (s *LedgerService) RenameLedger(oldName, newName string) error {
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("账本名称不能为空")
	}

	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return err
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return err
	}

	oldFilename := dbdata.ResolveLedgerFilename(oldName, config)
	oldPath := filepath.Join(ledgerDir, oldFilename)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("账本不存在: %s", oldName)
	}

	newFilename := dbdata.LedgerNameToFilename(newName)
	newPath := filepath.Join(ledgerDir, newFilename)

	// 检查新文件名是否和别的文件冲突（排除自身）
	if newFilename != oldFilename {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("账本名称已存在: %s", newName)
		}
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名失败: %w", err)
	}

	// 更新 NameMap
	delete(config.NameMap, oldFilename)
	config.NameMap[newFilename] = newName

	if config.CurrentLedger == oldName {
		config.CurrentLedger = newName
	}
	return dbdata.SaveLedgerConfig(config)
}

// ExportLedger 导出账本文件
func (s *LedgerService) ExportLedger(name string) ([]byte, error) {
	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return nil, err
	}

	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return nil, err
	}

	filename := dbdata.ResolveLedgerFilename(name, config)
	dbPath := filepath.Join(ledgerDir, filename)

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("账本不存在: %s", name)
	}

	return os.ReadFile(dbPath)
}

// ImportLedger 从文件导入账本
func (s *LedgerService) ImportLedger(name string, data []byte) (*models.LedgerInfo, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("账本名称不能为空")
	}

	ledgerDir, err := dbdata.GetLedgerDir()
	if err != nil {
		return nil, err
	}

	filename := dbdata.LedgerNameToFilename(name)
	dbPath := filepath.Join(ledgerDir, filename)
	if _, err := os.Stat(dbPath); err == nil {
		return nil, fmt.Errorf("账本已存在: %s", name)
	}

	if err := os.WriteFile(dbPath, data, 0644); err != nil {
		return nil, fmt.Errorf("写入失败: %w", err)
	}

	// 注册账本显示名
	config, err := dbdata.LoadLedgerConfig()
	if err != nil {
		return nil, err
	}
	config.NameMap[filename] = name
	if err := dbdata.SaveLedgerConfig(config); err != nil {
		fmt.Printf("[WARN] 注册导入账本名称失败: %v\n", err)
	}

	info, _ := os.Stat(dbPath)
	return &models.LedgerInfo{
		Name:        name,
		IsActive:    false,
		RecordCount: 0,
		CreatedAt:   info.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// ==================== 内部辅助函数 ====================

// resolveLedgerName 从文件名反推账本显示名
func resolveLedgerName(filename string, config *dbdata.LedgerConfig) string {
	if name, ok := config.NameMap[filename]; ok {
		return name
	}
	// 没有映射记录时，去掉 .db 后缀作为显示名
	return strings.TrimSuffix(filename, ".db")
}

// getRecordCountFromDB 打开临时引擎查询记录数
func getRecordCountFromDB(dbPath string) int64 {
	eng, err := xorm.NewEngine("sqlite", dbPath)
	if err != nil {
		return 0
	}
	defer eng.Close()

	count, err := eng.Count(new(models.Record))
	if err != nil {
		return 0
	}
	return count
}

// seedCategoriesInEngine 在指定引擎上初始化默认类别
func seedCategoriesInEngine(eng *xorm.Engine) error {
	count, err := eng.Where("is_system = ?", true).Count(new(models.Category))
	if err != nil || count > 0 {
		return err
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

	session := eng.NewSession()
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

// seedDefaultAccountInEngine 在指定引擎上初始化默认账户
func seedDefaultAccountInEngine(eng *xorm.Engine) error {
	count, err := eng.Count(new(models.Account))
	if err != nil || count > 0 {
		return err
	}

	defaultAccount := &models.Account{
		Name:      "现金",
		Icon:      "💵",
		Balance:   0,
		IsDefault: true,
		SortOrder: 0,
	}
	_, err = eng.Insert(defaultAccount)
	return err
}
