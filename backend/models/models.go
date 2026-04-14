package models

import "time"

// Record 记账记录
type Record struct {
	ID        int64     `json:"id"          xorm:"pk autoincr 'id'"`
	Type      string    `json:"type"        xorm:"not null index 'type'"`            // income / expense
	Amount    float64   `json:"amount"      xorm:"not null 'amount'"`
	Category  string    `json:"category"    xorm:"not null 'category'"`
	Note      string    `json:"note"        xorm:"default('') 'note'"`
	Date      string    `json:"date"        xorm:"not null index 'date'"`            // YYYY-MM-DD
	AccountID int64     `json:"account_id"  xorm:"index 'account_id'"`               // 账户ID（0=未分类）
	CreatedAt time.Time `json:"created_at"  xorm:"created 'created_at'"`
	Tags      []Tag     `json:"tags"        xorm:"-"`                                 // 标签（关联查询，不映射数据库列）
}

// TableName 指定表名
func (Record) TableName() string {
	return "records"
}

// Category 类别
type Category struct {
	ID       int64  `json:"id"        xorm:"pk autoincr 'id'"`
	Name     string `json:"name"      xorm:"unique not null 'name'"`
	Icon     string `json:"icon"      xorm:"not null default('📦') 'icon'"`
	Type     string `json:"type"      xorm:"not null index 'type'"` // income / expense
	IsSystem bool   `json:"is_system" xorm:"'is_system'"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// MonthlyStats 月度收支统计（纯内存结构，不映射表）
type MonthlyStats struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
	Month        string  `json:"month"`
}

// YearlyStats 年度收支统计（纯内存结构）
type YearlyStats struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
	Year         string  `json:"year"`
}

// MonthlyTrend 月度趋势（纯内存结构）
type MonthlyTrend struct {
	Month        string  `json:"month"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
}

// MonthlyCategoryStats 按类别月度统计（纯内存结构）
type MonthlyCategoryStats struct {
	Category     string  `json:"category"`
	CategoryIcon string  `json:"category_icon"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

// DailyStats 日度统计（纯内存结构）
type DailyStats struct {
	Date         string  `json:"date"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
}

// Budget 预算
type Budget struct {
	ID        int64     `json:"id"          xorm:"pk autoincr 'id'"`
	Type      string    `json:"type"        xorm:"not null index 'type'"`              // "monthly" / "yearly"
	Year      int       `json:"year"        xorm:"not null index 'year'"`              // 2026
	Month     int       `json:"month"       xorm:"not null default(0) index 'month'"` // 1-12（月度有值，年度为0）
	Amount    float64   `json:"amount"      xorm:"not null 'amount'"`                  // 预算金额
	CreatedAt time.Time `json:"created_at"  xorm:"created 'created_at'"`
	UpdatedAt time.Time `json:"updated_at"  xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (Budget) TableName() string {
	return "budgets"
}

// BudgetInfo 预算概览（纯内存结构）
type BudgetInfo struct {
	BudgetType  string  `json:"budget_type"`  // "monthly" / "yearly"
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	BudgetAmount float64 `json:"budget_amount"` // 预算金额
	Spent       float64 `json:"spent"`          // 已花费
	Remaining   float64 `json:"remaining"`      // 剩余
	Percentage  float64 `json:"percentage"`     // 使用百分比（0-100）
}

// Tag 标签
type Tag struct {
	ID        int64     `json:"id"         xorm:"pk autoincr 'id'"`
	Name      string    `json:"name"       xorm:"unique not null 'name'"`
	Color     string    `json:"color"      xorm:"not null default('#6366f1') 'color'"`
	CreatedAt time.Time `json:"created_at" xorm:"created 'created_at'"`
	UpdatedAt time.Time `json:"updated_at" xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// RecordTag 记录-标签关联表
type RecordTag struct {
	ID       int64 `json:"id"        xorm:"pk autoincr 'id'"`
	RecordID int64 `json:"record_id" xorm:"not null index 'record_id'"`
	TagID    int64 `json:"tag_id"    xorm:"not null index 'tag_id'"`
}

// TableName 指定表名
func (RecordTag) TableName() string {
	return "record_tag"
}

// Reminder 提醒功能配置（每种提醒功能一行种子记录，控制开关和参数）
type Reminder struct {
	ID           int64  `json:"id"            xorm:"pk autoincr 'id'"`
	Type         string `json:"type"          xorm:"not null 'type'"`             // "budget_warning" / "daily_reminder" / "weekly_summary" / "task_reminder"
	BudgetType   string `json:"budget_type"    xorm:"default('') 'budget_type'"`  // "monthly" / "yearly"（仅 budget_warning）
	Threshold    int    `json:"threshold"      xorm:"not null default(80) 'threshold'"` // 预算百分比阈值（仅 budget_warning）
	ReminderTime string `json:"reminder_time"  xorm:"default('') 'reminder_time'"`  // 提醒时间 HH:MM（仅 daily_reminder）
	Enabled      bool   `json:"enabled"        xorm:"not null default(true) 'enabled'"`
	Message      string `json:"message"        xorm:"default('') 'message'"`       // 提醒消息文案
}

// TableName 指定表名
func (Reminder) TableName() string {
	return "reminders"
}

// ReminderSettings 提醒全局设置
type ReminderSettings struct {
	ID               int64  `json:"id"                 xorm:"pk autoincr 'id'"`
	WebhookURL       string `json:"webhook_url"        xorm:"default('') 'webhook_url'"`       // 全局 Webhook 地址
	WebhookEnabled   bool   `json:"webhook_enabled"    xorm:"not null default(false) 'webhook_enabled'"`
	PopupEnabled            bool   `json:"popup_enabled"             xorm:"not null default(true) 'popup_enabled'"`               // 弹窗提醒开关
	SystemNotificationEnabled bool `json:"system_notification_enabled" xorm:"not null default(false) 'system_notification_enabled'"` // 系统通知开关
	TaskReminderDays        int    `json:"task_reminder_days"        xorm:"not null default(1) 'task_reminder_days'"`              // 任务提前提醒天数
	UpdatedAt        time.Time `json:"updated_at"      xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (ReminderSettings) TableName() string {
	return "reminder_settings"
}

// TaskReminder 任务提醒记录（用于追踪已发送的提醒）
type TaskReminder struct {
	ID        int64     `json:"id"          xorm:"pk autoincr 'id'"`
	TaskID    int64     `json:"task_id"     xorm:"not null index 'task_id'"`
	RemindDate string   `json:"remind_date" xorm:"not null 'remind_date'"` // YYYY-MM-DD 提醒日期
	SentAt    time.Time `json:"sent_at"     xorm:"created 'sent_at'"`
}

// TableName 指定表名
func (TaskReminder) TableName() string {
	return "task_reminders"
}

// BudgetAlert 预算提醒结果
type BudgetAlert struct {
	Type       string  `json:"type"`        // "monthly" / "yearly"
	Percentage float64 `json:"percentage"`  // 已使用百分比
	Budget     float64 `json:"budget"`      // 预算金额
	Spent      float64 `json:"spent"`       // 已花费
	Remaining  float64 `json:"remaining"`   // 剩余
	Message    string  `json:"message"`     // 提醒消息
}

// WeeklySummary 周汇总
type WeeklySummary struct {
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	TotalIncome float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance     float64 `json:"balance"`
	TopCategory string  `json:"top_category"`
	Message     string  `json:"message"`
}

// RecurringRecord 周期记账规则
type RecurringRecord struct {
	ID          int64     `json:"id"           xorm:"pk autoincr 'id'"`
	Type        string    `json:"type"         xorm:"not null index 'type'"`             // income / expense
	Amount      float64   `json:"amount"       xorm:"not null 'amount'"`
	Category    string    `json:"category"     xorm:"not null 'category'"`
	Note        string    `json:"note"         xorm:"default('') 'note'"`
	AccountID   int64     `json:"account_id"   xorm:"default(0) 'account_id'"`            // 账户ID（0=未指定）
	Frequency   string    `json:"frequency"    xorm:"not null index 'frequency'"`        // daily / weekly / monthly / yearly
	WeekDay     int       `json:"week_day"     xorm:"default(0) 'week_day'"`              // 1-7（仅 weekly）
	MonthDay    int       `json:"month_day"    xorm:"default(0) 'month_day'"`             // 1-31（仅 monthly）
	YearMonth   int       `json:"year_month"   xorm:"default(0) 'year_month'"`            // 1-12（仅 yearly）
	NextDate    string    `json:"next_date"    xorm:"not null index 'next_date'"`         // YYYY-MM-DD 下次执行日期
	Enabled     bool      `json:"enabled"      xorm:"not null default(true) 'enabled'"`
	LastRunDate string    `json:"last_run_date" xorm:"default('') 'last_run_date'"`       // YYYY-MM-DD 上次执行日期
	CreatedAt   time.Time `json:"created_at"   xorm:"created 'created_at'"`
	UpdatedAt   time.Time `json:"updated_at"   xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (RecurringRecord) TableName() string {
	return "recurring_records"
}

// Account 账户
type Account struct {
	ID        int64     `json:"id"          xorm:"pk autoincr 'id'"`
	Name      string    `json:"name"        xorm:"unique not null 'name'"`
	Icon      string    `json:"icon"        xorm:"not null default('💳') 'icon'"`
	Balance   float64   `json:"balance"     xorm:"not null default(0) 'balance'"`     // 初始余额（手动设置）
	IsDefault bool      `json:"is_default"  xorm:"not null default(false) 'is_default'"` // 默认账户
	SortOrder int       `json:"sort_order"  xorm:"not null default(0) 'sort_order'"`
	CreatedAt time.Time `json:"created_at"  xorm:"created 'created_at'"`
	UpdatedAt time.Time `json:"updated_at"  xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}

// Transfer 转账记录
type Transfer struct {
	ID          int64     `json:"id"           xorm:"pk autoincr 'id'"`
	FromAccount int64     `json:"from_account" xorm:"not null index 'from_account'"` // 源账户ID
	ToAccount   int64     `json:"to_account"   xorm:"not null index 'to_account'"`   // 目标账户ID
	Amount      float64   `json:"amount"       xorm:"not null 'amount'"`              // 转账金额
	Note        string    `json:"note"         xorm:"default('') 'note'"`             // 备注
	Date        string    `json:"date"         xorm:"not null index 'date'"`           // YYYY-MM-DD
	RecordOutID int64     `json:"record_out_id" xorm:"'record_out_id'"`               // 关联的转出记录ID（支出）
	RecordInID  int64     `json:"record_in_id"  xorm:"'record_in_id'"`               // 关联的转入记录ID（收入）
	CreatedAt   time.Time `json:"created_at"   xorm:"created 'created_at'"`
}

// TableName 指定表名
func (Transfer) TableName() string {
	return "transfers"
}

// AccountInfo 账户信息（含实时余额，纯内存结构）
type AccountInfo struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Icon       string  `json:"icon"`
	Balance    float64 `json:"balance"`    // 初始余额
	RealBalance float64 `json:"real_balance"` // 实时余额（初始 + 收入 - 支出 + 转入 - 转出）
	IsDefault  bool    `json:"is_default"`
}

// SearchRequest 搜索请求参数
type SearchRequest struct {
	StartDate   string `json:"start_date"`    // YYYY-MM-DD
	EndDate     string `json:"end_date"`      // YYYY-MM-DD
	Type        string `json:"type"`          // income / expense（空=全部）
	Category    string `json:"category"`      // 类别名（空=全部，兼容单类筛选）
	CategoryIDs string `json:"category_ids"`  // 类别ID逗号分隔（如 "1,3,5"，支持多选）
	Keyword     string `json:"keyword"`       // 关键词（匹配类别+备注）
	AccountID   int64  `json:"account_id"`    // 账户ID（0=全部）
	TagID       int64  `json:"tag_id"`        // 标签ID（0=全部）
	Page        int64  `json:"page"`          // 页码（从1开始）
	Limit       int64  `json:"limit"`         // 每页条数
}

// SearchResult 搜索响应
type SearchResult struct {
	Records []Record `json:"records"`
	Total   int64    `json:"total"`
	Page    int64    `json:"page"`
	Limit   int64    `json:"limit"`
}

// BackupInfo 备份信息（纯内存结构，用于前端展示）
type BackupInfo struct {
	Filename    string `json:"filename"`     // 实际文件名（系统用）
	DisplayName string `json:"display_name"` // 显示名称（前端展示）
	FileSize    int64  `json:"file_size"`    // 文件大小（字节）
	CreatedAt   string `json:"created_at"`   // 备份时间
	Auto        bool   `json:"auto"`        // 是否自动备份
	LedgerName  string `json:"ledger_name"`  // 所属账本名
}

// LedgerInfo 账本信息（纯内存结构）
type LedgerInfo struct {
	Name      string `json:"name"`       // 账本名称
	IsActive  bool   `json:"is_active"`  // 是否当前活跃
	RecordCount int64 `json:"record_count"` // 记录数量
	CreatedAt string `json:"created_at"` // 创建时间
}

// ParsedRecord 智能识别解析结果（纯内存结构）
type ParsedRecord struct {
	Type     string   `json:"type"`
	Amount   float64  `json:"amount"`
	Category string   `json:"category"`
	Note     string   `json:"note"`
	Date     string   `json:"date"`
	Tags     []string `json:"tags"`
}

// ExportRecord CSV 导出的一条记录（纯内存结构）
type ExportRecord struct {
	Date     string  `json:"date"`
	Type     string  `json:"type"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Note     string  `json:"note"`
}

// ImportResult 导入结果（纯内存结构）
type ImportResult struct {
	SuccessCount int      `json:"successCount"`
	SkipCount    int      `json:"skipCount"`
	ErrorCount   int      `json:"errorCount"`
	Errors       []string `json:"errors"`
}

// WebhookPayload Webhook 请求体（纯内存结构）
type WebhookPayload struct {
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Task 待办任务
type Task struct {
	ID        int64     `json:"id"          xorm:"pk autoincr 'id'"`
	Title     string    `json:"title"       xorm:"not null 'title'"`
	DueDate   string    `json:"due_date"    xorm:"not null index 'due_date'"`  // YYYY-MM-DD
	Amount    float64   `json:"amount"      xorm:"default(0) 'amount'"`        // 关联金额（可选）
	Completed bool      `json:"completed"   xorm:"not null default(false) 'completed'"`
	CreatedAt time.Time `json:"created_at"  xorm:"created 'created_at'"`
	UpdatedAt time.Time `json:"updated_at"  xorm:"updated 'updated_at'"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}
