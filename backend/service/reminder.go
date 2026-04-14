package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"

	"xorm.io/xorm"
)

// BudgetInfoGetter 预算信息获取接口（供预算预警使用）
type BudgetInfoGetter interface {
	GetBudgetInfo(budgetType string, year, month int) (*models.BudgetInfo, error)
}

// HTTPClient HTTP 客户端结构体（便于测试）
type HTTPClient struct {
	Client *http.Client
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ReminderService 提醒业务逻辑层
type ReminderService struct {
	budgetService BudgetInfoGetter
}

// NewReminderService 创建提醒服务实例
func NewReminderService() *ReminderService {
	return &ReminderService{}
}

// SetBudgetServiceForReminder 注入 BudgetService（供预算预警使用）
func (s *ReminderService) SetBudgetServiceForReminder(getter BudgetInfoGetter) {
	s.budgetService = getter
}

// InitDefaultReminders 初始化默认提醒设置
func (s *ReminderService) InitDefaultReminders() error {
	return initDefaultReminders()
}

// GetAllReminders 获取所有提醒设置
func (s *ReminderService) GetAllReminders() ([]models.Reminder, error) {
	return getAllReminders()
}

// GetReminderSettings 获取全局提醒设置
func (s *ReminderService) GetReminderSettings() (*models.ReminderSettings, error) {
	return getReminderSettings()
}

// UpdateReminderSettings 更新全局提醒设置
func (s *ReminderService) UpdateReminderSettings(webhookURL string, webhookEnabled, popupEnabled, systemNotificationEnabled bool, taskReminderDays int) error {
	return updateReminderSettings(webhookURL, webhookEnabled, popupEnabled, systemNotificationEnabled, taskReminderDays)
}

// UpdateReminder 更新提醒设置（开关、阈值、时间）
func (s *ReminderService) UpdateReminder(id int64, enabled bool, threshold int, message string, reminderTime string, taskReminderDays int) error {
	return updateReminder(id, enabled, threshold, message, reminderTime)
}

// CheckBudgetAlerts 检查预算预警
func (s *ReminderService) CheckBudgetAlerts() ([]models.BudgetAlert, error) {
	return checkBudgetAlerts(s.budgetService)
}

// GetDailyReminder 获取每日提醒
func (s *ReminderService) GetDailyReminder() (*models.Reminder, error) {
	return getDailyReminder()
}

// GetWeeklySummary 获取周汇总提醒
func (s *ReminderService) GetWeeklySummary() (*models.WeeklySummary, error) {
	return getWeeklySummary()
}

// TestWebhookURL 测试指定 Webhook 地址
func (s *ReminderService) TestWebhookURL(webhookURL string, payload *models.WebhookPayload) error {
	client := &http.Client{Timeout: 10 * time.Second}
	return sendToURL(client, webhookURL, payload)
}

// SendWebhookNotification 发送 Webhook 通知（使用数据库中保存的设置）
func (s *ReminderService) SendWebhookNotification(payload *models.WebhookPayload) error {
	settings, err := getReminderSettings()
	if err != nil {
		return err
	}

	if !settings.WebhookEnabled || settings.WebhookURL == "" {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return sendToURL(client, settings.WebhookURL, payload)
}

// SendBudgetAlertWebhook 发送预算预警 Webhook
func (s *ReminderService) SendBudgetAlertWebhook(alert models.BudgetAlert) error {
	return s.SendWebhookNotification(&models.WebhookPayload{
		Type:    "budget_alert",
		Title:   "预算预警",
		Message: alert.Message,
		Data: map[string]interface{}{
			"budget_type": alert.Type,
			"percentage":  alert.Percentage,
			"budget":      alert.Budget,
			"spent":       alert.Spent,
			"remaining":   alert.Remaining,
		},
	})
}

// SendDailyReminderWebhook 发送每日记账提醒 Webhook
func (s *ReminderService) SendDailyReminderWebhook(message string) error {
	return s.SendWebhookNotification(&models.WebhookPayload{
		Type:    "daily_reminder",
		Title:   "记账提醒",
		Message: message,
		Data:    map[string]interface{}{},
	})
}

// SendTaskReminderWebhook 发送任务提醒 Webhook
func (s *ReminderService) SendTaskReminderWebhook(tasks []models.Task) error {
	taskList := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		taskList = append(taskList, map[string]interface{}{
			"id":       t.ID,
			"title":    t.Title,
			"due_date": t.DueDate,
			"amount":   t.Amount,
		})
	}

	return s.SendWebhookNotification(&models.WebhookPayload{
		Type:    "task_reminder",
		Title:   "待办任务提醒",
		Message: fmt.Sprintf("您有 %d 个待办任务即将到期", len(tasks)),
		Data: map[string]interface{}{
			"task_count": len(tasks),
			"tasks":      taskList,
		},
	})
}

// SendWeeklySummaryWebhook 发送周汇总 Webhook
func (s *ReminderService) SendWeeklySummaryWebhook(payload *models.WebhookPayload) error {
	return s.SendWebhookNotification(payload)
}

// GetTasksNeedReminder 获取需要提醒的任务（不标记已提醒，用于静默检查）
func (s *ReminderService) GetTasksNeedReminder() ([]models.Task, error) {
	return getTasksNeedReminder()
}

// CheckTaskReminders 检查并返回任务提醒
func (s *ReminderService) CheckTaskReminders() ([]models.Task, error) {
	return checkTaskReminders()
}

// InitDefaultRemindersInEngine 在指定引擎上初始化默认提醒（供账本创建使用）
func (s *ReminderService) InitDefaultRemindersInEngine(eng *xorm.Engine) error {
	return initDefaultRemindersInEngine(eng)
}

// ==================== 以下为内部函数（不导出） ====================

func initDefaultReminders() error {
	count, err := dbdata.ORM.Count(&models.Reminder{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	reminders := []models.Reminder{
		{Type: "budget_warning", BudgetType: "monthly", Threshold: 80, Enabled: true, Message: "本月预算已使用超过80%"},
		{Type: "budget_warning", BudgetType: "yearly", Threshold: 80, Enabled: true, Message: "本年度预算已使用超过80%"},
		{Type: "daily_reminder", Enabled: false, Message: "记得记账哦", ReminderTime: "20:00"},
		{Type: "weekly_summary", Enabled: true, Message: "本周消费汇总"},
		{Type: "task_reminder", Enabled: true, Message: "您有待办任务即将到期"},
	}

	for _, r := range reminders {
		if _, err := dbdata.ORM.Insert(&r); err != nil {
			return err
		}
	}

	// 初始化全局设置
	var settingsCount int64
	settingsCount, err = dbdata.ORM.Count(&models.ReminderSettings{})
	if err != nil {
		return err
	}
	if settingsCount == 0 {
		settings := &models.ReminderSettings{
			WebhookURL:       "",
			WebhookEnabled:   false,
			PopupEnabled:     true,
			TaskReminderDays: 1,
		}
		if _, err := dbdata.ORM.Insert(settings); err != nil {
			return err
		}
	}

	return nil
}

func getAllReminders() ([]models.Reminder, error) {
	reminders := make([]models.Reminder, 0)
	err := dbdata.ORM.Find(&reminders)
	return reminders, err
}

func getReminderSettings() (*models.ReminderSettings, error) {
	settings := &models.ReminderSettings{}
	has, err := dbdata.ORM.Get(settings)
	if err != nil {
		return nil, err
	}
	if !has {
		// 创建默认设置
		settings = &models.ReminderSettings{
			WebhookURL:       "",
			WebhookEnabled:   false,
			PopupEnabled:     true,
			TaskReminderDays: 1,
		}
		if _, err := dbdata.ORM.Insert(settings); err != nil {
			return nil, err
		}
	}
	return settings, nil
}

func updateReminderSettings(webhookURL string, webhookEnabled, popupEnabled, systemNotificationEnabled bool, taskReminderDays int) error {
	settings, err := getReminderSettings()
	if err != nil {
		return err
	}

	settings.WebhookURL = webhookURL
	settings.WebhookEnabled = webhookEnabled
	settings.PopupEnabled = popupEnabled
	settings.SystemNotificationEnabled = systemNotificationEnabled
	settings.TaskReminderDays = taskReminderDays

	_, err = dbdata.ORM.ID(settings.ID).Cols("webhook_url", "webhook_enabled", "popup_enabled", "system_notification_enabled", "task_reminder_days").Update(settings)
	return err
}

func updateReminder(id int64, enabled bool, threshold int, message string, reminderTime string) error {
	reminder := &models.Reminder{}
	has, err := dbdata.ORM.ID(id).Get(reminder)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("提醒设置不存在")
	}

	reminder.Enabled = enabled

	if reminder.Type == "budget_warning" && threshold > 0 {
		reminder.Threshold = threshold
	}
	if message != "" {
		reminder.Message = message
	}
	if reminder.Type == "daily_reminder" {
		reminder.ReminderTime = reminderTime
	}

	_, err = dbdata.ORM.ID(id).Cols("enabled", "threshold", "message", "reminder_time").Update(reminder)
	return err
}

func checkBudgetAlerts(budgetSvc BudgetInfoGetter) ([]models.BudgetAlert, error) {
	alerts := make([]models.BudgetAlert, 0)

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	reminders, err := getAllReminders()
	if err != nil {
		return nil, err
	}

	for _, r := range reminders {
		if !r.Enabled || r.Type != "budget_warning" {
			continue
		}

		var budgetInfo *models.BudgetInfo
		if budgetSvc != nil {
			if r.BudgetType == "monthly" {
				budgetInfo, err = budgetSvc.GetBudgetInfo("monthly", year, month)
			} else {
				budgetInfo, err = budgetSvc.GetBudgetInfo("yearly", year, 0)
			}
		}
		if err != nil {
			continue
		}
		if budgetInfo == nil {
			continue // 没有设置预算，跳过
		}

		if budgetInfo.Percentage >= float64(r.Threshold) {
			alert := models.BudgetAlert{
				Type:       r.BudgetType,
				Percentage: budgetInfo.Percentage,
				Budget:     budgetInfo.BudgetAmount,
				Spent:      budgetInfo.Spent,
				Remaining:  budgetInfo.Remaining,
				Message:    generateAlertMessage(r, budgetInfo),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func generateAlertMessage(reminder models.Reminder, info *models.BudgetInfo) string {
	if reminder.Message != "" {
		return fmt.Sprintf("%s（已使用 %.1f%%）", reminder.Message, info.Percentage)
	}

	var period string
	if reminder.BudgetType == "monthly" {
		period = "本月"
	} else {
		period = "本年"
	}

	return fmt.Sprintf("%s预算已使用 %.1f%%，剩余 ¥%.2f", period, info.Percentage, info.Remaining)
}

func getDailyReminder() (*models.Reminder, error) {
	reminder := &models.Reminder{}
	has, err := dbdata.ORM.Where("type = ?", "daily_reminder").Get(reminder)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return reminder, nil
}

func getWeeklySummary() (*models.WeeklySummary, error) {
	reminder := &models.Reminder{}
	has, err := dbdata.ORM.Where("type = ?", "weekly_summary").Get(reminder)
	if err != nil {
		return nil, err
	}
	if !has || !reminder.Enabled {
		return nil, nil
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1)
	endOfWeek := now

	startDate := startOfWeek.Format("2006-01-02")
	endDate := endOfWeek.Format("2006-01-02")

	var totalIncome, totalExpense float64
	type Result struct {
		Type   string  `xorm:"type"`
		Amount float64 `xorm:"amount"`
	}
	results := make([]Result, 0)
	err = dbdata.ORM.Table("record").
		Select("type, SUM(amount) as amount").
		Where("date >= ? AND date <= ?", startDate, endDate).
		GroupBy("type").
		Find(&results)
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		if r.Type == "income" {
			totalIncome = r.Amount
		} else {
			totalExpense = r.Amount
		}
	}

	type CategoryResult struct {
		Category string  `xorm:"category"`
		Amount   float64 `xorm:"amount"`
	}
	catResults := make([]CategoryResult, 0)
	err = dbdata.ORM.Table("record").
		Select("category, SUM(amount) as amount").
		Where("date >= ? AND date <= ? AND type = ?", startDate, endDate, "expense").
		GroupBy("category").
		Desc("amount").
		Limit(1).
		Find(&catResults)

	topCategory := ""
	if len(catResults) > 0 {
		topCategory = catResults[0].Category
	}

	return &models.WeeklySummary{
		StartDate:    startDate,
		EndDate:      endDate,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Balance:      totalIncome - totalExpense,
		TopCategory:  topCategory,
		Message:      reminder.Message,
	}, nil
}

// ==================== Webhook 通知（内部） ====================

type webhookType int

const (
	webhookTypeGeneric  webhookType = iota
	webhookTypeWecom
	webhookTypeDingtalk
	webhookTypeFeishu
)

func detectWebhookType(url string) webhookType {
	switch {
	case strings.Contains(url, "qyapi.weixin.qq.com") || strings.Contains(url, "qyapi.weixin.qq.com/cgi-bin/webhook"):
		return webhookTypeWecom
	case strings.Contains(url, "oapi.dingtalk.com"):
		return webhookTypeDingtalk
	case strings.Contains(url, "open.feishu.cn") || strings.Contains(url, "open.larksuite.com"):
		return webhookTypeFeishu
	default:
		return webhookTypeGeneric
	}
}

func buildWecomPayload(payload *models.WebhookPayload) interface{} {
	content := fmt.Sprintf("**%s**\n%s", payload.Title, payload.Message)
	if payload.Timestamp > 0 {
		t := time.Unix(payload.Timestamp, 0)
		content += fmt.Sprintf("\n> %s", t.Format("2006-01-02 15:04:05"))
	}
	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}
}

func buildDingtalkPayload(payload *models.WebhookPayload) interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n%s", payload.Title, payload.Message),
		},
	}
}

func buildFeishuPayload(payload *models.WebhookPayload) interface{} {
	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("%s\n%s", payload.Title, payload.Message),
		},
	}
}

func sendToURL(httpClient *http.Client, webhookURL string, payload *models.WebhookPayload) error {
	if webhookURL == "" {
		return fmt.Errorf("Webhook 地址为空")
	}

	payload.Timestamp = time.Now().Unix()

	var bodyData interface{}
	wt := detectWebhookType(webhookURL)
	switch wt {
	case webhookTypeWecom:
		bodyData = buildWecomPayload(payload)
	case webhookTypeDingtalk:
		bodyData = buildDingtalkPayload(payload)
	case webhookTypeFeishu:
		bodyData = buildFeishuPayload(payload)
	default:
		bodyData = payload
	}

	jsonData, err := json.Marshal(bodyData)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("无效的 Webhook 地址: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PiggyAccounting/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("网络请求失败，请检查地址是否可访问: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("服务器返回错误状态码 %d，请检查 Webhook 地址是否正确", resp.StatusCode)
	}

	if wt == webhookTypeWecom {
		var result struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr == nil {
			if result.ErrCode != 0 {
				return fmt.Errorf("企业微信返回错误: %s (errcode=%d)", result.ErrMsg, result.ErrCode)
			}
		}
	}

	return nil
}

// ==================== 任务提醒 ====================

func getTasksNeedReminder() ([]models.Task, error) {
	// 检查任务提醒是否开启
	reminder := &models.Reminder{}
	has, err := dbdata.ORM.Where("type = ?", "task_reminder").Get(reminder)
	if err != nil {
		return nil, err
	}
	if !has || !reminder.Enabled {
		return nil, nil
	}

	// 从全局设置获取提前提醒天数
	settings, err := getReminderSettings()
	if err != nil {
		return nil, err
	}
	days := settings.TaskReminderDays
	if days <= 0 {
		days = 1
	}

	today := time.Now().Format("2006-01-02")
	futureDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	tasks := make([]models.Task, 0)
	err = dbdata.ORM.Where("completed = ? AND due_date >= ? AND due_date <= ?", false, today, futureDate).
		OrderBy("due_date ASC").
		Find(&tasks)
	if err != nil {
		return nil, err
	}

	result := make([]models.Task, 0)
	for _, task := range tasks {
		if !hasTaskRemindedToday(task.ID) {
			result = append(result, task)
		}
	}

	return result, nil
}

func hasTaskRemindedToday(taskID int64) bool {
	today := time.Now().Format("2006-01-02")
	count, err := dbdata.ORM.Where("task_id = ? AND remind_date = ?", taskID, today).Count(&models.TaskReminder{})
	if err != nil {
		return false
	}
	return count > 0
}

func markTaskReminded(taskID int64) error {
	today := time.Now().Format("2006-01-02")
	reminder := &models.TaskReminder{
		TaskID:     taskID,
		RemindDate: today,
	}
	_, err := dbdata.ORM.Insert(reminder)
	return err
}

func checkTaskReminders() ([]models.Task, error) {
	tasks, err := getTasksNeedReminder()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		markTaskReminded(task.ID)
	}

	return tasks, nil
}

func initDefaultRemindersInEngine(eng *xorm.Engine) error {
	count, err := eng.Count(new(models.Reminder))
	if err != nil || count > 0 {
		return err
	}

	reminders := []models.Reminder{
		{Type: "budget_warning", BudgetType: "monthly", Threshold: 80, Enabled: true, Message: "本月预算已使用超过80%"},
		{Type: "budget_warning", BudgetType: "yearly", Threshold: 80, Enabled: true, Message: "本年度预算已使用超过80%"},
		{Type: "daily_reminder", Enabled: false, Message: "记得记账哦", ReminderTime: "20:00"},
		{Type: "weekly_summary", Enabled: false, Message: "本周消费汇总"},
		{Type: "task_reminder", Enabled: true, Message: "您有待办任务即将到期"},
	}

	session := eng.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	for i := range reminders {
		if _, err := session.Insert(&reminders[i]); err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}
