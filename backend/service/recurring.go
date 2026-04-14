package service

import (
	"fmt"
	"time"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// RecurringRecordAdder 记录添加接口（周期记账需要 Add 和 AddWithAccount）
type RecurringRecordAdder interface {
	Add(recordType, category, note, date string, amount float64) (int64, error)
	AddWithAccount(recordType, category, note, date string, amount float64, accountID int64) (int64, error)
}

// RecurringService 周期记账业务逻辑层
type RecurringService struct {
	recordService RecurringRecordAdder
}

// NewRecurringService 创建周期记账服务实例
func NewRecurringService() *RecurringService {
	return &RecurringService{}
}

// SetRecordService 注入 RecordService
func (s *RecurringService) SetRecordService(svc RecurringRecordAdder) {
	s.recordService = svc
}

// Create 创建周期记账规则
func (s *RecurringService) Create(recType, category, note, frequency string, amount float64, weekDay, monthDay, yearMonth int, accountID int64) (*models.RecurringRecord, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("金额必须大于0")
	}
	if recType != "income" && recType != "expense" {
		return nil, fmt.Errorf("类型必须是 income 或 expense")
	}
	if !isValidFrequency(frequency) {
		return nil, fmt.Errorf("无效的周期类型: %s（支持 daily/weekly/monthly/yearly）", frequency)
	}

	nextDate := calculateNextDate(frequency, weekDay, monthDay, yearMonth, time.Now())

	r := &models.RecurringRecord{
		Type:      recType,
		Amount:    amount,
		Category:  category,
		Note:      note,
		AccountID: accountID,
		Frequency: frequency,
		WeekDay:   weekDay,
		MonthDay:  monthDay,
		YearMonth: yearMonth,
		NextDate:  nextDate,
		Enabled:   true,
	}

	_, err := dbdata.ORM.Insert(r)
	if err != nil {
		return nil, fmt.Errorf("创建周期记账失败: %w", err)
	}
	return r, nil
}

// GetAll 获取所有周期记账规则
func (s *RecurringService) GetAll() ([]models.RecurringRecord, error) {
	records := make([]models.RecurringRecord, 0)
	if err := dbdata.ORM.OrderBy("created_at DESC").Find(&records); err != nil {
		return nil, err
	}
	return records, nil
}

// Update 更新周期记账规则
func (s *RecurringService) Update(id int64, recType, category, note, frequency string, amount float64, weekDay, monthDay, yearMonth int, accountID int64, enabled bool) error {
	r := new(models.RecurringRecord)
	found, err := dbdata.ORM.ID(id).Get(r)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("周期记账规则不存在")
	}

	if amount <= 0 {
		return fmt.Errorf("金额必须大于0")
	}
	if recType != "income" && recType != "expense" {
		return fmt.Errorf("类型必须是 income 或 expense")
	}
	if !isValidFrequency(frequency) {
		return fmt.Errorf("无效的周期类型: %s", frequency)
	}

	r.Type = recType
	r.Amount = amount
	r.Category = category
	r.Note = note
	r.AccountID = accountID
	r.Frequency = frequency
	r.WeekDay = weekDay
	r.MonthDay = monthDay
	r.YearMonth = yearMonth
	r.Enabled = enabled
	r.NextDate = calculateNextDate(frequency, weekDay, monthDay, yearMonth, time.Now())

	_, err = dbdata.ORM.ID(id).Update(r)
	return err
}

// ToggleEnabled 切换启用状态
func (s *RecurringService) ToggleEnabled(id int64) error {
	r := new(models.RecurringRecord)
	found, err := dbdata.ORM.ID(id).Get(r)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("周期记账规则不存在")
	}

	r.Enabled = !r.Enabled
	_, err = dbdata.ORM.ID(id).Cols("enabled", "updated_at").Update(r)
	return err
}

// Delete 删除周期记账规则
func (s *RecurringService) Delete(id int64) error {
	_, err := dbdata.ORM.ID(id).Delete(new(models.RecurringRecord))
	return err
}

// ExecutePending 执行到期的周期记账（应用启动时调用）
func (s *RecurringService) ExecutePending() ([]models.Record, error) {
	today := time.Now().Format("2006-01-02")
	var created []models.Record

	// 查找已启用且到期需要执行的规则
	records := make([]models.RecurringRecord, 0)
	if err := dbdata.ORM.Where("enabled = ? AND next_date <= ?", true, today).
		Find(&records); err != nil {
		return nil, fmt.Errorf("查询到期规则失败: %w", err)
	}

	for i := range records {
		r := &records[i]
		if s.recordService == nil {
			continue
		}

		// 生成需要补记的所有日期
		dates := generateDatesBetween(*r, today)

		for _, date := range dates {
			var recordID int64
			var err error
			if r.AccountID > 0 {
				recordID, err = s.recordService.AddWithAccount(r.Type, r.Category, r.Note, date, r.Amount, r.AccountID)
			} else {
				recordID, err = s.recordService.Add(r.Type, r.Category, r.Note, date, r.Amount)
			}
			if err != nil {
				fmt.Printf("执行周期记账失败 [规则ID=%d, 日期=%s]: %v\n", r.ID, date, err)
				continue
			}

			created = append(created, models.Record{
				ID:        recordID,
				Type:      r.Type,
				Amount:    r.Amount,
				Category:  r.Category,
				Note:      r.Note,
				Date:      date,
				CreatedAt: time.Now(),
			})
		}

		// 更新规则的下次执行日期和上次执行日期
		r.LastRunDate = today
		r.NextDate = calculateNextDate(r.Frequency, r.WeekDay, r.MonthDay, r.YearMonth, time.Now())
		dbdata.ORM.ID(r.ID).Cols("last_run_date", "next_date", "updated_at").Update(r)
	}

	return created, nil
}

// GetFrequencyLabel 获取周期类型的中文标签
func GetFrequencyLabel(frequency string) string {
	switch frequency {
	case "daily":
		return "每天"
	case "weekly":
		return "每周"
	case "monthly":
		return "每月"
	case "yearly":
		return "每年"
	default:
		return frequency
	}
}

// GetFrequencyDetail 获取周期详情描述
func GetFrequencyDetail(r models.RecurringRecord) string {
	switch r.Frequency {
	case "daily":
		return "每天执行"
	case "weekly":
		weekdays := []string{"", "周一", "周二", "周三", "周四", "周五", "周六", "周日"}
		if r.WeekDay >= 1 && r.WeekDay <= 7 {
			return "每" + weekdays[r.WeekDay]
		}
		return "每周执行"
	case "monthly":
		if r.MonthDay > 0 {
			return fmt.Sprintf("每月%d日", r.MonthDay)
		}
		return "每月执行"
	case "yearly":
		if r.YearMonth > 0 && r.MonthDay > 0 {
			return fmt.Sprintf("每年%d月%d日", r.YearMonth, r.MonthDay)
		} else if r.YearMonth > 0 {
			return fmt.Sprintf("每年%d月", r.YearMonth)
		}
		return "每年执行"
	default:
		return r.Frequency
	}
}

// ==================== 内部辅助函数 ====================

func isValidFrequency(frequency string) bool {
	switch frequency {
	case "daily", "weekly", "monthly", "yearly":
		return true
	}
	return false
}

func calculateNextDate(frequency string, weekDay, monthDay, yearMonth int, from time.Time) string {
	switch frequency {
	case "daily":
		return from.AddDate(0, 0, 1).Format("2006-01-02")
	case "weekly":
		return nextWeekday(from, weekDay)
	case "monthly":
		return nextMonthDay(from, monthDay)
	case "yearly":
		return nextYearMonthDay(from, yearMonth, monthDay)
	default:
		return from.AddDate(0, 0, 1).Format("2006-01-02")
	}
}

func nextWeekday(from time.Time, weekday int) string {
	if weekday < 1 || weekday > 7 {
		weekday = 1
	}
	goWeekday := time.Weekday(weekday % 7)
	daysUntil := int(goWeekday - from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil).Format("2006-01-02")
}

func nextMonthDay(from time.Time, day int) string {
	if day < 1 || day > 31 {
		day = 1
	}
	next := from.AddDate(0, 1, 0)
	maxDay := daysInMonth(next.Year(), next.Month())
	if day > maxDay {
		day = maxDay
	}
	return time.Date(next.Year(), next.Month(), day, 0, 0, 0, 0, from.Location()).Format("2006-01-02")
}

func nextYearMonthDay(from time.Time, month, day int) string {
	if month < 1 || month > 12 {
		month = 1
	}
	if day < 1 || day > 31 {
		day = 1
	}
	next := from.AddDate(1, 0, 0)
	maxDay := daysInMonth(next.Year(), time.Month(month))
	if day > maxDay {
		day = maxDay
	}
	return time.Date(next.Year(), time.Month(month), day, 0, 0, 0, 0, from.Location()).Format("2006-01-02")
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func generateDatesBetween(r models.RecurringRecord, endDate string) []string {
	var dates []string
	end, _ := time.Parse("2006-01-02", endDate)

	start, _ := time.Parse("2006-01-02", r.NextDate)
	if r.LastRunDate != "" {
		last, _ := time.Parse("2006-01-02", r.LastRunDate)
		start = calculateDateFrom(last, r.Frequency, r.WeekDay, r.MonthDay, r.YearMonth)
	}

	for !start.After(end) {
		dates = append(dates, start.Format("2006-01-02"))
		start = calculateDateFrom(start, r.Frequency, r.WeekDay, r.MonthDay, r.YearMonth)
		if len(dates) > 365 {
			break
		}
	}
	return dates
}

func calculateDateFrom(from time.Time, frequency string, weekDay, monthDay, yearMonth int) time.Time {
	switch frequency {
	case "daily":
		return from.AddDate(0, 0, 1)
	case "weekly":
		if weekDay < 1 || weekDay > 7 {
			weekDay = 1
		}
		goWeekday := time.Weekday(weekDay % 7)
		daysUntil := int(goWeekday - from.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7
		}
		return from.AddDate(0, 0, daysUntil)
	case "monthly":
		next := from.AddDate(0, 1, 0)
		day := monthDay
		if day < 1 || day > 31 {
			day = 1
		}
		maxDay := daysInMonth(next.Year(), next.Month())
		if day > maxDay {
			day = maxDay
		}
		return time.Date(next.Year(), next.Month(), day, 0, 0, 0, 0, from.Location())
	case "yearly":
		next := from.AddDate(1, 0, 0)
		month := yearMonth
		day := monthDay
		if month < 1 || month > 12 {
			month = 1
		}
		if day < 1 || day > 31 {
			day = 1
		}
		maxDay := daysInMonth(next.Year(), time.Month(month))
		if day > maxDay {
			day = maxDay
		}
		return time.Date(next.Year(), time.Month(month), day, 0, 0, 0, 0, from.Location())
	default:
		return from.AddDate(0, 0, 1)
	}
}
