package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goRuntime "runtime"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"piggy-accounting/backend/base"
	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"
	"piggy-accounting/backend/service"
)

// App 应用结构体 - 仅保留需要 context 的基础设施方法
// 所有业务 API 已通过 main.go 直接绑定 service 结构体暴露给前端
type App struct {
	ctx         context.Context
	hasTestData bool
}

// NewApp 创建新应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化日志系统
	initLogger()

	base.Info("应用启动", "platform", goRuntime.GOOS)

	// 初始化数据库
	hasTestData, err := dbdata.Init()
	if err != nil {
		base.Error("数据库初始化失败", "error", err)
	}
	a.hasTestData = hasTestData
	base.Info("数据库初始化完成", "hasTestData", hasTestData)

	// 初始化默认提醒设置（提醒逻辑在 service 层，dbdata 不能反向引用）
	if err := service.Reminders.InitDefaultReminders(); err != nil {
		base.Error("初始化提醒设置失败", "error", err)
	}

	// 注入服务依赖（解决循环依赖）
	service.Reminders.SetBudgetServiceForReminder(service.Budgets)
	service.SmartRecognize.SetRecordServiceForRecognize(service.Records.Create)
	service.Recurring.SetRecordService(service.Records)
	service.Transfers.SetRecordService(service.Records)

	// 执行到期的周期记账
	created, err := service.Recurring.ExecutePending()
	if err != nil {
		base.Error("执行周期记账失败", "error", err)
	} else if len(created) > 0 {
		base.Info("周期记账自动执行", "count", len(created))
	}

	// 启动提醒定时任务
	go a.startReminderTicker()
}

// initLogger 初始化 base 日志系统
func initLogger() {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".piggy-accounting", "logs")
	os.MkdirAll(logDir, 0755)

	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, today+".log")

	cfg := base.LogConfig{
		Level:      "info",
		Output:     "file",
		File:       logPath,
		Format:     "text",
		Color:      false, // 文件输出禁用颜色
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
		AddSource:  true,
	}
	base.InitLog(cfg)
	base.Info("日志系统初始化完成", "logPath", logPath)
}

// shutdown 应用关闭时调用
func (a *App) shutdown(_ context.Context) {
	base.Info("应用关闭")
	dbdata.Close()
	base.Close()
}

// startReminderTicker 启动提醒定时检查（每分钟检查一次）
func (a *App) startReminderTicker() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// 各类提醒的"今日已触发"标记
	lastDailyDate := ""
	lastWeeklyDate := ""
	lastTaskDate := ""
	lastBudgetAlertDate := ""

	// 启动时做一次静默检查：预算预警弹窗 + 待办任务弹窗（不发 Webhook，不标记已提醒）
	go func() {
		time.Sleep(3 * time.Second)
		a.runSilentStartupCheck()
	}()

	for range ticker.C {
		a.runReminderChecks(&lastDailyDate, &lastWeeklyDate, &lastTaskDate, &lastBudgetAlertDate)
	}
}

// runSilentStartupCheck 启动时静默检查（只推前端弹窗，不发 Webhook，不标记已提醒）
func (a *App) runSilentStartupCheck() {
	// 1. 预算预警弹窗
	alerts, err := service.Reminders.CheckBudgetAlerts()
	if err != nil {
		base.Error("启动时预算预警检查失败", "error", err)
	} else if len(alerts) > 0 {
		settings, _ := service.Reminders.GetReminderSettings()
		if settings == nil || settings.PopupEnabled {
			for _, alert := range alerts {
				wailsRuntime.EventsEmit(a.ctx, "budget_alert", map[string]interface{}{
					"type":       alert.Type,
					"percentage": alert.Percentage,
					"message":    alert.Message,
				})
			}
			base.Info("启动时发现预算预警", "count", len(alerts))
		}
	}

	// 2. 待办任务弹窗
	tasks, err := service.Reminders.GetTasksNeedReminder()
	if err != nil {
		base.Error("启动时任务提醒检查失败", "error", err)
		return
	}
	if len(tasks) == 0 {
		return
	}

	base.Info("启动时发现待到期任务", "count", len(tasks))
	settings, _ := service.Reminders.GetReminderSettings()
	if settings == nil || settings.PopupEnabled {
		wailsRuntime.EventsEmit(a.ctx, "task_reminder", map[string]interface{}{
			"count": len(tasks),
			"tasks": tasks,
		})
	}
}

// runReminderChecks 定时检查（每分钟）—— 所有自动通知统一在每日提醒时间点触发
//
// 设计原则（简洁清晰）：
//   1. 每日记账提醒 → 精确到用户设定时间（如 16:00），当天首次到达该时间时触发
//   2. 预算预警 / 待办任务 / 周汇总 → 全部跟随步骤1一起触发，不各自判断时间
//   3. 启动时（3秒后）只做静默弹窗检查，不发 Webhook、不标记已处理
func (a *App) runReminderChecks(lastDailyDate, lastWeeklyDate, lastTaskDate, lastBudgetAlertDate *string) {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	// --- 步骤1：每日提醒是否到了触发时间？---

	reminder, err := service.Reminders.GetDailyReminder()
	if err == nil && reminder != nil && reminder.Enabled && len(reminder.ReminderTime) >= 5 {
		triggerTime := reminder.ReminderTime[:5] // 如 "16:00"
		if *lastDailyDate != today && currentTime >= triggerTime {
			// ✅ 到了！今日首次到达设定时间
			*lastDailyDate = today

			base.Info("触发每日提醒", "time", triggerTime)

			settings, _ := service.Reminders.GetReminderSettings()

			// 每日提醒本身：前端弹窗 + Webhook
			if settings == nil || settings.PopupEnabled {
				wailsRuntime.EventsEmit(a.ctx, "daily_reminder", map[string]interface{}{
					"time":    triggerTime,
					"message": reminder.Message,
				})
			}
			if err2 := service.Reminders.SendDailyReminderWebhook(reminder.Message); err2 != nil {
				base.Error("发送每日提醒 Webhook 失败", "error", err2)
			}

			// --- 步骤2：跟随每日提醒，一次性推送其他所有通知 ---
			a.triggerBudgetAlertNotify(today, lastBudgetAlertDate) // 预算预警
			a.triggerTaskReminders(today, lastTaskDate)            // 待办任务
			a.triggerWeeklySummary(now, today, lastWeeklyDate)     // 周汇总
		}
	} else if (reminder == nil || !reminder.Enabled) && *lastDailyDate != today && currentTime >= "09:00" && currentTime <= "22:00" {
		// 没开每日提醒时兜底：白天任意时刻只触发一次其他通知
		*lastDailyDate = today
		base.Info("触发兜底通知（无每日提醒）", "time", currentTime)
		a.triggerBudgetAlertNotify(today, lastBudgetAlertDate)
		a.triggerTaskReminders(today, lastTaskDate)
		a.triggerWeeklySummary(now, today, lastWeeklyDate)
	}

	// 其余 tick（未到达时间 / 已触发过）：什么都不做
}

// triggerBudgetAlertNotify 预算预警推送（仅在每日提醒触发的步骤2中调用）
func (a *App) triggerBudgetAlertNotify(today string, lastBudgetAlertDate *string) {
	if *lastBudgetAlertDate == today {
		return
	}

	alerts, err := service.Reminders.CheckBudgetAlerts()
	if err != nil || len(alerts) == 0 {
		return
	}
	*lastBudgetAlertDate = today

	base.Info("预算预警自动通知", "count", len(alerts))
	settings, _ := service.Reminders.GetReminderSettings()

	if settings == nil || settings.PopupEnabled {
		for _, alert := range alerts {
			wailsRuntime.EventsEmit(a.ctx, "budget_alert", map[string]interface{}{
				"type":       alert.Type,
				"percentage": alert.Percentage,
				"message":    alert.Message,
			})
		}
	}
	for _, alert := range alerts {
		if err := service.Reminders.SendBudgetAlertWebhook(alert); err != nil {
			base.Error("发送预算预警 Webhook 失败", "error", err)
		}
	}
}

// triggerTaskReminders 待办任务提醒推送（仅在每日提醒触发的步骤2中调用）
func (a *App) triggerTaskReminders(today string, lastTaskDate *string) {
	if *lastTaskDate == today {
		return
	}

	tasks, err := service.Reminders.CheckTaskReminders()
	if err != nil || len(tasks) == 0 {
		return
	}
	*lastTaskDate = today

	base.Info("待办任务提醒", "count", len(tasks))
	settings, _ := service.Reminders.GetReminderSettings()

	if settings == nil || settings.PopupEnabled {
		wailsRuntime.EventsEmit(a.ctx, "task_reminder", map[string]interface{}{
			"count": len(tasks),
			"tasks": tasks,
		})
	}
	if err := service.Reminders.SendTaskReminderWebhook(tasks); err != nil {
		base.Error("发送任务提醒 Webhook 失败", "error", err)
	}
}

// triggerWeeklySummary 周消费汇总推送（仅在每周一 + 步骤2 中调用）
func (a *App) triggerWeeklySummary(now time.Time, today string, lastWeeklyDate *string) {
	if now.Weekday() != time.Monday {
		return
	}
	if *lastWeeklyDate == today {
		return
	}

	summary, err := service.Reminders.GetWeeklySummary()
	if err != nil || summary == nil {
		return
	}
	*lastWeeklyDate = today

	base.Info("周消费汇总", "spent", summary.TotalExpense, "income", summary.TotalIncome)
	settings, _ := service.Reminders.GetReminderSettings()

	if settings == nil || settings.PopupEnabled {
		wailsRuntime.EventsEmit(a.ctx, "weekly_summary", map[string]interface{}{
			"message":      summary.Message,
			"total_income": summary.TotalIncome,
			"total_expense": summary.TotalExpense,
			"balance":      summary.Balance,
			"top_category": summary.TopCategory,
			"start_date":   summary.StartDate,
			"end_date":     summary.EndDate,
		})
	}
	payload := &models.WebhookPayload{
		Type:    "weekly_summary",
		Title:   "📊 周消费汇总",
		Message: fmt.Sprintf("本周支出 ¥%.1f | 收入 ¥%.1f | 结余 ¥%.1f\n最高支出类别: %s",
			summary.TotalExpense, summary.TotalIncome, summary.Balance, summary.TopCategory),
		Data: map[string]interface{}{
			"total_income":  summary.TotalIncome,
			"total_expense": summary.TotalExpense,
			"balance":       summary.Balance,
			"top_category":  summary.TopCategory,
		},
	}
	if err := service.Reminders.SendWeeklySummaryWebhook(payload); err != nil {
		base.Error("发送周汇总 Webhook 失败", "error", err)
	}
}

// ==================== 前端基础设施 API（需要 context） ====================

// LogFrontend 前端日志接口 —— 前端可调用此方法写入后端日志
func (a *App) LogFrontend(level, message string) {
	switch level {
	case "ERROR":
		base.Error("[Frontend] " + message)
	case "WARN":
		base.Warn("[Frontend] " + message)
	default:
		base.Info("[Frontend] " + message)
	}
}

// GetPlatform 返回当前运行平台（darwin / windows / linux）
func (a *App) GetPlatform() string {
	return goRuntime.GOOS
}

// HasTestData 返回本次启动是否写入了测试数据（仅首次创建新账本时为 true）
func (a *App) HasTestData() bool {
	return a.hasTestData
}

// GetCurrentMonth 获取当前月份（YYYY-MM 格式）
func (a *App) GetCurrentMonth() string {
	return service.GetCurrentMonth()
}

// GetCurrentDate 获取当前日期（YYYY-MM-DD 格式）
func (a *App) GetCurrentDate() string {
	return service.GetCurrentDate()
}

// SelectDirectory 弹出原生目录选择对话框
func (a *App) SelectDirectory(title string) (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
	})
}

// SaveExportToDir 保存导出文件到指定目录（需要 os 包写文件）
func (a *App) SaveExportToDir(filename, dir string, data []byte) (string, error) {
	base.Info("保存导出文件", "filename", filename, "dir", dir, "size", len(data))
	destPath := filepath.Join(dir, filename)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		base.Error("保存导出文件失败", "error", err)
		return "", err
	}
	base.Info("保存导出文件成功", "path", destPath)
	return destPath, nil
}

// ExportBackupToDesktop 导出备份到桌面（需要 os.UserHomeDir）
func (a *App) ExportBackupToDesktop(filename string) (string, error) {
	base.Info("导出备份到桌面", "filename", filename)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(homeDir, "Desktop")
	result, err := service.Backups.ExportBackupToDir(filename, destDir)
	if err != nil {
		base.Error("导出备份到桌面失败", "error", err)
		return "", err
	}
	base.Info("导出备份到桌面成功", "path", result)
	return result, nil
}

// TestWebhook 测试 Webhook（需要构造 payload）
func (a *App) TestWebhook(webhookURL string) error {
	base.Info("测试 Webhook", "url", webhookURL)
	if webhookURL == "" {
		return fmt.Errorf("Webhook 地址不能为空")
	}
	return service.Reminders.TestWebhookURL(webhookURL, &models.WebhookPayload{
		Type:    "test",
		Title:   "猪猪记账 - 测试通知",
		Message: "这是一条测试消息，Webhook 配置成功！",
		Data: map[string]interface{}{
			"test":    true,
			"app":     "猪猪记账",
			"version": "1.0.0",
		},
	})
}
