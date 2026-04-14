package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"
)

// BackupService 备份业务逻辑层
type BackupService struct{}

// NewBackupService 创建备份服务实例
func NewBackupService() *BackupService {
	return &BackupService{}
}

// CreateBackup 创建备份
func (s *BackupService) CreateBackup(auto bool) (*models.BackupInfo, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return nil, fmt.Errorf("获取备份目录失败: %w", err)
	}

	dbPath, err := dbdata.GetCurrentDBPath()
	if err != nil {
		return nil, fmt.Errorf("获取数据库路径失败: %w", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("数据库文件不存在")
	}

	now := time.Now()
	ledgerName := getCurrentLedgerName()

	filename := generateBackupFilename(ledgerName, auto, now)
	destPath := filepath.Join(backupDir, filename)

	baseDest := destPath
	for i := 2; ; i++ {
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(baseDest)
		noExt := strings.TrimSuffix(baseDest, ext)
		destPath = fmt.Sprintf("%s_%d%s", noExt, i, ext)
	}

	src, err := os.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("创建备份文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("复制数据失败: %w", err)
	}

	info, _ := os.Stat(destPath)

	displayName := buildBackupDisplayName(auto, ledgerName, now)

	return &models.BackupInfo{
		Filename:    filename,
		DisplayName: displayName,
		FileSize:    info.Size(),
		CreatedAt:   now.Format("2006-01-02 15:04:05"),
		Auto:        auto,
		LedgerName:  ledgerName,
	}, nil
}

// RestoreBackup 从备份恢复
func (s *BackupService) RestoreBackup(filename string) error {
	backupDir, err := getBackupDir()
	if err != nil {
		return err
	}

	srcPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", filename)
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("非法文件名")
	}

	dbPath, err := dbdata.GetCurrentDBPath()
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filename, ".db") {
		return fmt.Errorf("非法文件类型")
	}

	if err := dbdata.Close(); err != nil {
		fmt.Printf("[WARN] 关闭数据库连接失败: %v\n", err)
	}

	dataDir, _ := getDataDir()
	now := time.Now()
	timeStr := now.Format("01-02_1504")
	preRestoreFile := filepath.Join(dataDir, "backups", fmt.Sprintf("恢复前_%s.db", timeStr))
	src, err := os.Open(dbPath)
	if err == nil {
		dst, err := os.Create(preRestoreFile)
		if err == nil {
			io.Copy(dst, src)
			dst.Close()
		}
		src.Close()
	}

	src, err = os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开备份文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("写入数据库失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("恢复数据失败: %w", err)
	}

	_, err = dbdata.Init()
	return err
}

// DeleteBackup 删除备份
func (s *BackupService) DeleteBackup(filename string) error {
	backupDir, err := getBackupDir()
	if err != nil {
		return err
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("非法文件名")
	}

	filePath := filepath.Join(backupDir, filename)
	return os.Remove(filePath)
}

// ListBackups 列出所有备份
func (s *BackupService) ListBackups() ([]models.BackupInfo, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return []models.BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return []models.BackupInfo{}, nil
	}

	var backups []models.BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		auto, displayName, ledgerName := parseBackupFileName(name)

		backups = append(backups, models.BackupInfo{
			Filename:    name,
			DisplayName: displayName,
			FileSize:    info.Size(),
			CreatedAt:   info.ModTime().Format("2006-01-02 15:04:05"),
			Auto:        auto,
			LedgerName:  ledgerName,
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// ExportBackup 导出备份到前端
func (s *BackupService) ExportBackup(filename string) ([]byte, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return nil, err
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return nil, fmt.Errorf("非法文件名")
	}

	filePath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("备份文件不存在: %s", filename)
	}

	return os.ReadFile(filePath)
}

// ExportBackupToDir 将备份导出到指定目录
func (s *BackupService) ExportBackupToDir(filename, destDir string) (string, error) {
	backupDir, err := getBackupDir()
	if err != nil {
		return "", err
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return "", fmt.Errorf("非法文件名")
	}

	srcPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("备份文件不存在: %s", filename)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("创建目标目录失败: %w", err)
	}

	destPath := filepath.Join(destDir, filename)
	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("打开备份文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("导出失败: %w", err)
	}

	return destPath, nil
}

// ClearLedgerData 清空当前账本的所有用户数据
// 保留系统类别（is_system=true）、默认账户（is_default=true）、提醒设置
// 清空：记账记录、预算、转账、周期记账、标签、记录-标签关联、待办任务、任务提醒
func (s *BackupService) ClearLedgerData() (int64, error) {
	// 统计清空前记录数
	recordCount, _ := dbdata.ORM.Count(new(models.Record))

	session := dbdata.ORM.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return 0, fmt.Errorf("开始事务失败: %w", err)
	}

	// 按外键依赖顺序删除
	tables := []interface{}{
		new(models.RecordTag),
		new(models.TaskReminder),
		new(models.Record),
		new(models.Transfer),
		new(models.Budget),
		new(models.RecurringRecord),
		new(models.Tag),
		new(models.Task),
		new(models.Reminder),
		// 不删除：Category（保留系统类别）、Account（保留默认账户）、ReminderSettings
	}

	for _, table := range tables {
		if _, err := session.Where("1=1").Delete(table); err != nil {
			session.Rollback()
			return 0, fmt.Errorf("清空表数据失败: %w", err)
		}
	}

	// 清除非默认账户（保留默认的"现金"账户）
	if _, err := session.Where("is_default = ?", false).Delete(new(models.Account)); err != nil {
		session.Rollback()
		return 0, fmt.Errorf("清空非默认账户失败: %w", err)
	}

	// 重置默认账户余额
	if _, err := session.Where("is_default = ?", true).Cols("balance").Update(&models.Account{Balance: 0}); err != nil {
		session.Rollback()
		return 0, fmt.Errorf("重置默认账户余额失败: %w", err)
	}

	// 清除非系统类别（用户自定义类别）
	if _, err := session.Where("is_system = ?", false).Delete(new(models.Category)); err != nil {
		session.Rollback()
		return 0, fmt.Errorf("清空自定义类别失败: %w", err)
	}

	if err := session.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return recordCount, nil
}

// CleanupOldBackups 清理超限的旧备份
func (s *BackupService) CleanupOldBackups() {
	backupDir, err := getBackupDir()
	if err != nil {
		return
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	type fileEntry struct {
		name         string
		modTime      time.Time
		isPrerestore bool
	}
	var files []fileEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEntry{
			name:         entry.Name(),
			modTime:      info.ModTime(),
			isPrerestore: strings.HasPrefix(entry.Name(), "prerestore_"),
		})
	}

	var normalFiles []fileEntry
	for _, f := range files {
		if !f.isPrerestore {
			normalFiles = append(normalFiles, f)
		}
	}

	sort.Slice(normalFiles, func(i, j int) bool {
		return normalFiles[i].modTime.Before(normalFiles[j].modTime)
	})

	excess := len(normalFiles) - maxBackupCount
	for i := 0; i < excess; i++ {
		path := filepath.Join(backupDir, normalFiles[i].name)
		if err := os.Remove(path); err == nil {
			fmt.Printf("[Backup] 清理旧备份: %s\n", normalFiles[i].name)
		}
	}
}

// ==================== 内部辅助函数 ====================

const (
	backupPrefixManual = "手动_"
	backupPrefixAuto   = "auto_"
	maxBackupCount     = 30
)

func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".piggy-accounting"), nil
}

func getBackupDir() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(dataDir, "backups")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func getCurrentLedgerName() string {
	config, _ := dbdata.LoadLedgerConfig()
	name := config.CurrentLedger
	if name == "" {
		name = "默认账本"
	}
	return name
}

func generateBackupFilename(ledgerName string, auto bool, t time.Time) string {
	prefix := backupPrefixManual
	if auto {
		prefix = backupPrefixAuto
	}
	timestamp := t.Format("01-02_1504")
	safeName := safeFilenameForBackup(ledgerName)
	if safeName == "" || safeName == "default" {
		safeName = "默认账本"
	}
	return fmt.Sprintf("%s%s_%s.db", prefix, safeName, timestamp)
}

func safeFilenameForBackup(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	result := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '-'
		default:
			return r
		}
	}, trimmed)
	runes := []rune(result)
	if len(runes) > 10 {
		result = string(runes[:10])
	}
	for len(result) > 0 && (result[len(result)-1] == '-' || result[len(result)-1] == '_') {
		result = result[:len(result)-1]
	}
	if result == "" {
		return "default"
	}
	return result
}

func buildBackupDisplayName(auto bool, ledgerName string, t time.Time) string {
	tag := "手动"
	if auto {
		tag = "自动"
	}
	timeStr := t.Format("01-02 15:04")
	return fmt.Sprintf("%s · %s · %s", tag, ledgerName, timeStr)
}

func parseBackupFileName(filename string) (auto bool, displayName, ledgerName string) {
	stripped := strings.TrimSuffix(filename, ".db")

	var prefix, rest string
	if strings.HasPrefix(stripped, backupPrefixAuto) {
		auto = true
		prefix = backupPrefixAuto
		rest = strings.TrimPrefix(stripped, prefix)
	} else if strings.HasPrefix(stripped, backupPrefixManual) {
		auto = false
		prefix = backupPrefixManual
		rest = strings.TrimPrefix(stripped, prefix)
	} else if strings.HasPrefix(stripped, "恢复前_") {
		timeStr := strings.TrimPrefix(stripped, "恢复前_")
		return false, fmt.Sprintf("恢复前 · %s", timeStr), "恢复前"
	} else {
		auto = strings.HasPrefix(filename, "auto_")
		ledgerName = "未知账本"
		displayName = filename
		return
	}

	lastUnderscore := strings.LastIndex(rest, "_")
	if lastUnderscore <= 0 {
		ledgerName = rest
		displayName = filename
		return
	}

	ledgerName = rest[:lastUnderscore]

	tag := "自动"
	if !auto {
		tag = "手动"
	}
	displayName = fmt.Sprintf("%s · %s", tag, ledgerName)
	return
}
