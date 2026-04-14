// Package service 业务逻辑层
// 负责处理业务规则、参数校验、协调 repository 层
// 每个服务对应一个业务领域，对外暴露接口便于测试和解耦
package service

import (
	"fmt"
	"time"
)

// 全局服务实例
// 这些实例在应用启动时初始化，供 app.go 调用
var (
	Records        *RecordService
	Categories     *CategoryService
	Budgets        *BudgetService
	ExportImport   *ExportImportService
	SmartRecognize *SmartRecognizeService
	Tags           *TagService
	Reminders      *ReminderService
	Recurring      *RecurringService
	Accounts       *AccountService
	Transfers      *TransferService
	Backups        *BackupService
	Ledgers        *LedgerService
	Tasks          *TaskService
)

// InitServices 初始化所有服务实例
// 在应用启动时调用，确保所有服务可用
func InitServices() {
	Records = NewRecordService()
	Categories = NewCategoryService()
	Budgets = NewBudgetService()
	ExportImport = NewExportImportService()
	SmartRecognize = NewSmartRecognizeService()
	Tags = NewTagService()
	Reminders = NewReminderService()
	Recurring = NewRecurringService()
	Accounts = NewAccountService()
	Transfers = NewTransferService()
	Backups = NewBackupService()
	Ledgers = NewLedgerService()
	Tasks = NewTaskService()
}

// ==================== 通用工具函数 ====================

// GetCurrentMonth 获取当前月份 YYYY-MM
func GetCurrentMonth() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), int(now.Month()))
}

// GetCurrentDate 获取当前日期 YYYY-MM-DD
func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// FormatMonthName 格式化月份为中文名
func FormatMonthName(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return t.Format("2006年1月")
}

// ParseDate 解析日期字符串，验证格式
func ParseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}

// ParseMonth 解析月份字符串，验证格式
func ParseMonth(month string) (time.Time, error) {
	return time.Parse("2006-01", month)
}

// ValidateAmount 验证金额是否有效
func ValidateAmount(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("金额不能为负数")
	}
	if amount > 1000000000 {
		return fmt.Errorf("金额超出限制")
	}
	return nil
}

// ValidateCategoryType 验证记录类型
func ValidateCategoryType(t string) error {
	if t != "expense" && t != "income" {
		return fmt.Errorf("无效的记录类型: %s，必须是 expense 或 income", t)
	}
	return nil
}
