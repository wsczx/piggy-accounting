package service

import (
	"fmt"
	"strconv"
	"strings"

	"xorm.io/xorm"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// RecordService 记账记录业务逻辑层
type RecordService struct{}

// NewRecordService 创建记录服务实例
func NewRecordService() *RecordService {
	return &RecordService{}
}

// Add 添加记账记录
func (s *RecordService) Add(recordType, category, note, date string, amount float64) (int64, error) {
	if err := s.validateRecord(recordType, amount); err != nil {
		return 0, err
	}
	if err := ValidateAmount(amount); err != nil {
		return 0, err
	}

	rec := &models.Record{
		Type:     recordType,
		Amount:   amount,
		Category: category,
		Note:     note,
		Date:     date,
	}
	_, err := dbdata.ORM.Insert(rec)
	if err != nil {
		return 0, fmt.Errorf("添加记录失败: %w", err)
	}
	return rec.ID, nil
}

// AddWithAccount 添加记账记录（含账户ID）
func (s *RecordService) AddWithAccount(recordType, category, note, date string, amount float64, accountID int64) (int64, error) {
	if err := s.validateRecord(recordType, amount); err != nil {
		return 0, err
	}
	if err := ValidateAmount(amount); err != nil {
		return 0, err
	}

	rec := &models.Record{
		Type:      recordType,
		Amount:    amount,
		Category:  category,
		Note:      note,
		Date:      date,
		AccountID: accountID,
	}
	_, err := dbdata.ORM.Insert(rec)
	if err != nil {
		return 0, fmt.Errorf("添加记录失败: %w", err)
	}
	return rec.ID, nil
}

// Create 通过 Record 结构体添加记录（供智能识别等内部使用）
func (s *RecordService) Create(record *models.Record) (int64, error) {
	if err := s.validateRecord(record.Type, record.Amount); err != nil {
		return 0, err
	}
	_, err := dbdata.ORM.Insert(record)
	if err != nil {
		return 0, fmt.Errorf("添加记录失败: %w", err)
	}
	return record.ID, nil
}

// GetByDateRange 按日期范围获取记录
func (s *RecordService) GetByDateRange(startDate, endDate string) ([]models.Record, error) {
	records := make([]models.Record, 0)
	session := dbdata.NewSession()
	defer session.Close()

	if startDate != "" && endDate != "" {
		session = session.Where("date >= ? AND date <= ?", startDate, endDate)
	}
	session = session.OrderBy("date DESC, created_at DESC")

	if err := session.Find(&records); err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}
	return records, nil
}

// SearchRecords 搜索记录（支持关键词、类别、类型筛选 + 分页）
func (s *RecordService) SearchRecords(req models.SearchRequest) (models.SearchResult, error) {
	// 参数默认值处理
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	result := models.SearchResult{}

	applyFilters := func(session *xorm.Session) *xorm.Session {
		if req.StartDate != "" && req.EndDate != "" {
			session = session.Where("date >= ? AND date <= ?", req.StartDate, req.EndDate)
		}
		if req.Type != "" {
			session = session.Where("type = ?", req.Type)
		}
		if req.CategoryIDs != "" {
			idStrs := strings.Split(req.CategoryIDs, ",")
			var ids []int64
			for _, str := range idStrs {
				str = strings.TrimSpace(str)
				if str == "" {
					continue
				}
				if id, err := strconv.ParseInt(str, 10, 64); err == nil {
					ids = append(ids, id)
				}
			}
			if len(ids) > 0 {
				var catNames []string
				dbdata.ORM.Table("categories").Cols("name").In("id", ids).Find(&catNames)
				if len(catNames) > 0 {
					session = session.In("category", catNames)
				}
			}
		} else if req.Category != "" {
			session = session.Where("category = ?", req.Category)
		}
		if req.Keyword != "" {
			like := "%" + req.Keyword + "%"
			session = session.Where("(category LIKE ? OR note LIKE ?)", like, like)
		}
		if req.AccountID > 0 {
			session = session.Where("account_id = ?", req.AccountID)
		}
		if req.TagID > 0 {
			session = session.Join("INNER", "record_tag", "records.id = record_tag.record_id").
				Where("record_tag.tag_id = ?", req.TagID)
		}
		return session
	}

	countSession := dbdata.ORM.NewSession()
	defer countSession.Close()
	total, err := applyFilters(countSession).Count(new(models.Record))
	if err != nil {
		return result, fmt.Errorf("查询记录数失败: %w", err)
	}
	result.Total = total

	dataSession := dbdata.ORM.NewSession()
	defer dataSession.Close()
	records := make([]models.Record, 0)
	if err := applyFilters(dataSession).OrderBy("date DESC, created_at DESC").
		Limit(int(req.Limit), int((req.Page-1)*req.Limit)).
		Find(&records); err != nil {
		return result, fmt.Errorf("查询记录失败: %w", err)
	}
	result.Records = records
	result.Page = req.Page
	result.Limit = req.Limit

	// 批量填充每条记录的标签
	if len(records) > 0 {
		s.fillRecordTags(records)
	}

	return result, nil
}

// fillRecordTags 批量查询记录关联的标签并填充到 Records[i].Tags
func (s *RecordService) fillRecordTags(records []models.Record) {
	// 收集所有记录ID
	recordIDs := make([]int64, 0, len(records))
	for _, r := range records {
		recordIDs = append(recordIDs, r.ID)
	}

	// 查询所有 record_tag 关联
	type recordTagRow struct {
		RecordID int64 `xorm:"record_id"`
		TagID    int64 `xorm:"tag_id"`
	}
	rows := make([]recordTagRow, 0)
	dbdata.ORM.Table("record_tag").In("record_id", recordIDs).Find(&rows)

	// 构建 recordID -> tagIDs 映射
	recTagMap := make(map[int64][]int64)
	for _, row := range rows {
		recTagMap[row.RecordID] = append(recTagMap[row.RecordID], row.TagID)
	}

	// 查询所有涉及的标签
	allTagIDs := make(map[int64]bool)
	for _, ids := range recTagMap {
		for _, id := range ids {
			allTagIDs[id] = true
		}
	}
	if len(allTagIDs) == 0 {
		return
	}
	tagIDList := make([]int64, 0, len(allTagIDs))
	for id := range allTagIDs {
		tagIDList = append(tagIDList, id)
	}
	allTags := make([]models.Tag, 0)
	dbdata.ORM.In("id", tagIDList).Find(&allTags)
	tagMap := make(map[int64]models.Tag)
	for _, t := range allTags {
		tagMap[t.ID] = t
	}

	// 填充到每条记录
	for i := range records {
		tagIDs, ok := recTagMap[records[i].ID]
		if !ok {
			continue
		}
		tags := make([]models.Tag, 0, len(tagIDs))
		for _, tid := range tagIDs {
			if t, ok := tagMap[tid]; ok {
				tags = append(tags, t)
			}
		}
		records[i].Tags = tags
	}
}

// GetByID 根据ID获取单条记录
func (s *RecordService) GetByID(id int64) (*models.Record, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的记录ID")
	}
	record := new(models.Record)
	found, err := dbdata.ORM.ID(id).Get(record)
	if err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("记录不存在")
	}
	return record, nil
}

// Update 更新记账记录
func (s *RecordService) Update(id int64, recordType, category, note, date string, amount float64) error {
	if id <= 0 {
		return fmt.Errorf("无效的记录ID")
	}
	if err := s.validateRecord(recordType, amount); err != nil {
		return err
	}

	rec := &models.Record{
		Type:     recordType,
		Amount:   amount,
		Category: category,
		Note:     note,
		Date:     date,
	}
	affected, err := dbdata.ORM.ID(id).Cols("type", "amount", "category", "note", "date").Update(rec)
	if err != nil {
		return fmt.Errorf("更新记录失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("记录不存在")
	}
	return nil
}

// Delete 删除记账记录
func (s *RecordService) Delete(id int64) error {
	if id <= 0 {
		return fmt.Errorf("无效的记录ID")
	}
	// 先清理关联的 record_tag
	if _, err := dbdata.ORM.Where("record_id = ?", id).Delete(new(models.RecordTag)); err != nil {
		return fmt.Errorf("清理记录标签关联失败: %w", err)
	}
	_, err := dbdata.ORM.ID(id).Delete(new(models.Record))
	return err
}

// GetMonthlyStats 获取月度收支统计
func (s *RecordService) GetMonthlyStats(month string) (models.MonthlyStats, error) {
	if _, err := ParseMonth(month); err != nil {
		return models.MonthlyStats{}, fmt.Errorf("无效的月份格式: %s", month)
	}

	var stats models.MonthlyStats
	stats.Month = month

	startDate := month + "-01"
	endDate := month + "-31"

	incomeSum, err := dbdata.ORM.Where("type = ? AND date >= ? AND date <= ?", "income", startDate, endDate).
		Sum(new(models.Record), "amount")
	if err != nil {
		return stats, fmt.Errorf("查询收入失败: %w", err)
	}
	stats.TotalIncome = incomeSum

	expenseSum, err := dbdata.ORM.Where("type = ? AND date >= ? AND date <= ?", "expense", startDate, endDate).
		Sum(new(models.Record), "amount")
	if err != nil {
		return stats, fmt.Errorf("查询支出失败: %w", err)
	}
	stats.TotalExpense = expenseSum

	stats.Balance = stats.TotalIncome - stats.TotalExpense
	return stats, nil
}

// GetMonthlyCategoryStats 获取按类别的月度统计
func (s *RecordService) GetMonthlyCategoryStats(month, recordType string) ([]models.MonthlyCategoryStats, error) {
	if _, err := ParseMonth(month); err != nil {
		return nil, fmt.Errorf("无效的月份格式: %s", month)
	}
	if err := ValidateCategoryType(recordType); err != nil {
		return nil, err
	}

	startDate := month + "-01"
	endDate := month + "-31"

	type categorySum struct {
		Category string  `xorm:"'category'"`
		Icon     string  `xorm:"'icon'"`
		Total    float64 `xorm:"'total'"`
	}

	var results []categorySum
	err := dbdata.ORM.Table("records").
		Join("LEFT", "categories", "records.category = categories.name AND records.type = categories.type").
		Select("records.category, COALESCE(categories.icon, '📦') as icon, SUM(records.amount) as total").
		Where("records.type = ? AND records.date >= ? AND records.date <= ?", recordType, startDate, endDate).
		GroupBy("records.category").
		OrderBy("total DESC").
		Find(&results)
	if err != nil {
		return nil, fmt.Errorf("查询类别统计失败: %w", err)
	}

	stats := make([]models.MonthlyCategoryStats, 0, len(results))
	var total float64
	for _, rs := range results {
		stats = append(stats, models.MonthlyCategoryStats{
			Category:     rs.Category,
			CategoryIcon: rs.Icon,
			Amount:       rs.Total,
		})
		total += rs.Total
	}
	for i := range stats {
		if total > 0 {
			stats[i].Percentage = stats[i].Amount / total * 100
		}
	}

	return stats, nil
}

// GetDailyStats 获取月度每日统计
func (s *RecordService) GetDailyStats(month string) ([]models.DailyStats, error) {
	if _, err := ParseMonth(month); err != nil {
		return nil, fmt.Errorf("无效的月份格式: %s", month)
	}

	startDate := month + "-01"
	endDate := month + "-31"

	type dailySum struct {
		Date    string  `xorm:"'date'"`
		Income  float64 `xorm:"'income'"`
		Expense float64 `xorm:"'expense'"`
	}

	var results []dailySum
	err := dbdata.ORM.Table("records").
		Select("date, COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income, "+
			"COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as expense").
		Where("date >= ? AND date <= ?", startDate, endDate).
		GroupBy("date").
		OrderBy("date").
		Find(&results)
	if err != nil {
		return nil, fmt.Errorf("查询每日统计失败: %w", err)
	}

	stats := make([]models.DailyStats, 0, len(results))
	for _, rs := range results {
		stats = append(stats, models.DailyStats{
			Date:         rs.Date,
			TotalIncome:  rs.Income,
			TotalExpense: rs.Expense,
		})
	}

	return stats, nil
}

// GetYearlyStats 获取年度收支统计
func (s *RecordService) GetYearlyStats(year string) (models.YearlyStats, error) {
	if len(year) != 4 {
		return models.YearlyStats{}, fmt.Errorf("无效的年份格式: %s", year)
	}

	var stats models.YearlyStats
	stats.Year = year

	startDate := year + "-01-01"
	endDate := year + "-12-31"

	incomeSum, err := dbdata.ORM.Where("type = ? AND date >= ? AND date <= ?", "income", startDate, endDate).
		Sum(new(models.Record), "amount")
	if err != nil {
		return stats, fmt.Errorf("查询年度收入失败: %w", err)
	}
	stats.TotalIncome = incomeSum

	expenseSum, err := dbdata.ORM.Where("type = ? AND date >= ? AND date <= ?", "expense", startDate, endDate).
		Sum(new(models.Record), "amount")
	if err != nil {
		return stats, fmt.Errorf("查询年度支出失败: %w", err)
	}
	stats.TotalExpense = expenseSum

	stats.Balance = stats.TotalIncome - stats.TotalExpense
	return stats, nil
}

// GetYearlyCategoryStats 获取按类别的年度统计
func (s *RecordService) GetYearlyCategoryStats(year, recordType string) ([]models.MonthlyCategoryStats, error) {
	if len(year) != 4 {
		return nil, fmt.Errorf("无效的年份格式: %s", year)
	}
	if err := ValidateCategoryType(recordType); err != nil {
		return nil, err
	}

	startDate := year + "-01-01"
	endDate := year + "-12-31"

	type categorySum struct {
		Category string  `xorm:"'category'"`
		Icon     string  `xorm:"'icon'"`
		Total    float64 `xorm:"'total'"`
	}

	var results []categorySum
	err := dbdata.ORM.Table("records").
		Join("LEFT", "categories", "records.category = categories.name AND records.type = categories.type").
		Select("records.category, COALESCE(categories.icon, '📦') as icon, SUM(records.amount) as total").
		Where("records.type = ? AND records.date >= ? AND records.date <= ?", recordType, startDate, endDate).
		GroupBy("records.category").
		OrderBy("total DESC").
		Find(&results)
	if err != nil {
		return nil, fmt.Errorf("查询年度类别统计失败: %w", err)
	}

	stats := make([]models.MonthlyCategoryStats, 0, len(results))
	var total float64
	for _, rs := range results {
		stats = append(stats, models.MonthlyCategoryStats{
			Category:     rs.Category,
			CategoryIcon: rs.Icon,
			Amount:       rs.Total,
		})
		total += rs.Total
	}
	for i := range stats {
		if total > 0 {
			stats[i].Percentage = stats[i].Amount / total * 100
		}
	}

	return stats, nil
}

// GetMonthlyTrend 获取年度月度趋势
func (s *RecordService) GetMonthlyTrend(year string) ([]models.MonthlyTrend, error) {
	if len(year) != 4 {
		return nil, fmt.Errorf("无效的年份格式: %s", year)
	}

	startDate := year + "-01-01"
	endDate := year + "-12-31"

	type monthSum struct {
		Month   string  `xorm:"'month'"`
		Income  float64 `xorm:"'income'"`
		Expense float64 `xorm:"'expense'"`
	}

	var results []monthSum
	err := dbdata.ORM.Table("records").
		Select("SUBSTR(date, 1, 7) as month, "+
			"COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income, "+
			"COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as expense").
		Where("date >= ? AND date <= ?", startDate, endDate).
		GroupBy("SUBSTR(date, 1, 7)").
		OrderBy("month").
		Find(&results)
	if err != nil {
		return nil, fmt.Errorf("查询月度趋势失败: %w", err)
	}

	trend := make([]models.MonthlyTrend, 0, 12)
	for _, rs := range results {
		trend = append(trend, models.MonthlyTrend{
			Month:        rs.Month,
			TotalIncome:  rs.Income,
			TotalExpense: rs.Expense,
		})
	}

	return trend, nil
}

// validateRecord 验证记录参数
func (s *RecordService) validateRecord(recordType string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("金额必须大于0")
	}
	return ValidateCategoryType(recordType)
}
