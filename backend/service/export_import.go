package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"
)

// ExportImportService 数据导入导出业务逻辑层
type ExportImportService struct{}

// NewExportImportService 创建导入导出服务实例
func NewExportImportService() *ExportImportService {
	return &ExportImportService{}
}

// ExportToCSV 导出记录为 CSV 格式
func (s *ExportImportService) ExportToCSV(startDate, endDate string) ([]byte, error) {
	records := make([]models.Record, 0)
	session := dbdata.ORM.NewSession()
	defer session.Close()

	if startDate != "" && endDate != "" {
		session = session.Where("date >= ? AND date <= ?", startDate, endDate)
	}
	session = session.OrderBy("date DESC, created_at DESC")

	if err := session.Find(&records); err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.UseCRLF = true

	headers := []string{"日期", "类型", "类别", "金额", "备注"}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	for _, r := range records {
		recordType := "支出"
		if r.Type == "income" {
			recordType = "收入"
		}

		row := []string{
			r.Date,
			recordType,
			r.Category,
			fmt.Sprintf("%.2f", r.Amount),
			r.Note,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("写入 CSV 数据失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("生成 CSV 失败: %w", err)
	}

	return buf.Bytes(), nil
}

// ImportFromCSV 从 CSV 导入记录
func (s *ExportImportService) ImportFromCSV(data []byte, skipExisting bool) (*models.ImportResult, error) {
	result := &models.ImportResult{
		SuccessCount: 0,
		SkipCount:    0,
		ErrorCount:   0,
		Errors:       make([]string, 0),
	}

	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 失败: %w", err)
	}

	if len(rows) == 0 {
		return result, nil
	}

	startRow := 0
	if isHeaderRow(rows[0]) {
		startRow = 1
	}

	categories, err := getAllCategories()
	if err != nil {
		return nil, fmt.Errorf("获取类别失败: %w", err)
	}

	for i, row := range rows[startRow:] {
		lineNum := startRow + i + 1

		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		record, parseErr := parseCSVRow(row, categories)
		if parseErr != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: %s", lineNum, parseErr.Error()))
			continue
		}

		if skipExisting {
			exists, _ := checkRecordExists(record)
			if exists {
				result.SkipCount++
				continue
			}
		}

		r := &models.Record{
			Type:     record.Type,
			Amount:   record.Amount,
			Category: record.Category,
			Note:     record.Note,
			Date:     record.Date,
		}
		if _, err := dbdata.ORM.Insert(r); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行插入失败: %v", lineNum, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// ImportFromWeChat 从微信账单导入
func (s *ExportImportService) ImportFromWeChat(data []byte, skipExisting bool) (*models.ImportResult, error) {
	result := &models.ImportResult{
		SuccessCount: 0,
		SkipCount:    0,
		ErrorCount:   0,
		Errors:       make([]string, 0),
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 失败: %w", err)
	}

	categories, err := getAllCategories()
	if err != nil {
		return nil, fmt.Errorf("获取类别失败: %w", err)
	}

	for i, row := range rows {
		lineNum := i + 1

		if len(row) < 6 || strings.Contains(row[0], "交易时间") {
			continue
		}

		record, parseErr := parseWeChatRow(row, categories)
		if parseErr != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: %s", lineNum, parseErr.Error()))
			continue
		}

		if record.Type == "income" {
			continue
		}

		if skipExisting {
			exists, _ := checkRecordExists(record)
			if exists {
				result.SkipCount++
				continue
			}
		}

		r := &models.Record{
			Type:     record.Type,
			Amount:   record.Amount,
			Category: record.Category,
			Note:     record.Note,
			Date:     record.Date,
		}
		if _, err := dbdata.ORM.Insert(r); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行插入失败: %v", lineNum, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// ImportFromAlipay 从支付宝账单导入
func (s *ExportImportService) ImportFromAlipay(data []byte, skipExisting bool) (*models.ImportResult, error) {
	result := &models.ImportResult{
		SuccessCount: 0,
		SkipCount:    0,
		ErrorCount:   0,
		Errors:       make([]string, 0),
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 失败: %w", err)
	}

	categories, err := getAllCategories()
	if err != nil {
		return nil, fmt.Errorf("获取类别失败: %w", err)
	}

	for i, row := range rows {
		lineNum := i + 1

		if len(row) < 11 || strings.Contains(row[0], "交易号") {
			continue
		}

		record, parseErr := parseAlipayRow(row, categories)
		if parseErr != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行: %s", lineNum, parseErr.Error()))
			continue
		}

		if record.Type == "income" {
			continue
		}

		if skipExisting {
			exists, _ := checkRecordExists(record)
			if exists {
				result.SkipCount++
				continue
			}
		}

		r := &models.Record{
			Type:     record.Type,
			Amount:   record.Amount,
			Category: record.Category,
			Note:     record.Note,
			Date:     record.Date,
		}
		if _, err := dbdata.ORM.Insert(r); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("第%d行插入失败: %v", lineNum, err))
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

// ==================== 辅助方法 ====================

func isHeaderRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	header := strings.ToLower(strings.TrimSpace(row[0]))
	return header == "日期" || header == "date" || header == "交易时间"
}

func parseCSVRow(row []string, categories map[string]string) (*models.ExportRecord, error) {
	if len(row) < 4 {
		return nil, fmt.Errorf("字段不足，至少需要4列")
	}

	record := &models.ExportRecord{}

	record.Date = strings.TrimSpace(row[0])
	if _, err := time.Parse("2006-01-02", record.Date); err != nil {
		return nil, fmt.Errorf("日期格式错误: %s", record.Date)
	}

	typeStr := strings.TrimSpace(row[1])
	switch typeStr {
	case "收入", "income", "Income":
		record.Type = "income"
	case "支出", "expense", "Expense", "消费":
		record.Type = "expense"
	default:
		return nil, fmt.Errorf("类型错误: %s", typeStr)
	}

	category := strings.TrimSpace(row[2])
	record.Category = matchCategory(category, categories, record.Type)

	amountStr := strings.TrimSpace(row[3])
	amountStr = strings.ReplaceAll(amountStr, "¥", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("金额错误: %s", amountStr)
	}
	record.Amount = amount

	if len(row) > 4 {
		record.Note = strings.TrimSpace(row[4])
	}

	return record, nil
}

func parseWeChatRow(row []string, categories map[string]string) (*models.ExportRecord, error) {
	if len(row) < 6 {
		return nil, fmt.Errorf("字段不足")
	}

	record := &models.ExportRecord{}

	timeStr := strings.TrimSpace(row[0])
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return nil, fmt.Errorf("时间格式错误: %s", timeStr)
	}
	record.Date = t.Format("2006-01-02")

	transType := strings.TrimSpace(row[1])
	counterparty := strings.TrimSpace(row[2])
	product := strings.TrimSpace(row[3])

	incomeExpense := strings.TrimSpace(row[4])
	switch incomeExpense {
	case "收入":
		record.Type = "income"
	case "支出":
		record.Type = "expense"
	default:
		return nil, fmt.Errorf("跳过非收支记录: %s", incomeExpense)
	}

	amountStr := strings.TrimSpace(row[5])
	amountStr = strings.ReplaceAll(amountStr, "¥", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("金额错误: %s", amountStr)
	}
	record.Amount = amount

	record.Category = inferCategoryFromWeChat(transType, counterparty, product, categories)

	record.Note = product
	if counterparty != "" && counterparty != "/" {
		record.Note = counterparty + " - " + product
	}

	return record, nil
}

func parseAlipayRow(row []string, categories map[string]string) (*models.ExportRecord, error) {
	if len(row) < 11 {
		return nil, fmt.Errorf("字段不足")
	}

	record := &models.ExportRecord{}

	timeStr := strings.TrimSpace(row[2])
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return nil, fmt.Errorf("时间格式错误: %s", timeStr)
	}
	record.Date = t.Format("2006-01-02")

	counterparty := strings.TrimSpace(row[7])
	product := strings.TrimSpace(row[8])

	amountStr := strings.TrimSpace(row[9])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("金额错误: %s", amountStr)
	}
	record.Amount = amount

	incomeExpense := strings.TrimSpace(row[10])
	switch incomeExpense {
	case "收入":
		record.Type = "income"
	case "支出":
		record.Type = "expense"
	default:
		return nil, fmt.Errorf("跳过非收支记录: %s", incomeExpense)
	}

	record.Category = inferCategoryFromAlipay(counterparty, product, categories)

	record.Note = product
	if counterparty != "" {
		record.Note = counterparty + " - " + product
	}

	return record, nil
}

func getAllCategories() (map[string]string, error) {
	categories := make(map[string]string)
	list := make([]models.Category, 0)
	if err := dbdata.ORM.Find(&list); err != nil {
		return nil, err
	}
	for _, c := range list {
		categories[c.Name] = c.Type
	}
	return categories, nil
}

func matchCategory(name string, categories map[string]string, recordType string) string {
	if _, exists := categories[name]; exists {
		return name
	}

	for catName, catType := range categories {
		if catType == recordType && strings.Contains(catName, name) {
			return catName
		}
	}

	if recordType == "income" {
		return "其他收入"
	}
	return "其他支出"
}

func inferCategoryFromWeChat(transType, counterparty, product string, categories map[string]string) string {
	text := transType + counterparty + product
	text = strings.ToLower(text)

	if strings.Contains(text, "餐饮") || strings.Contains(text, "外卖") ||
		strings.Contains(text, "餐厅") || strings.Contains(text, "美团") ||
		strings.Contains(text, "饿了么") || strings.Contains(text, "肯德基") ||
		strings.Contains(text, "麦当劳") || strings.Contains(text, "星巴克") {
		if _, ok := categories["餐饮"]; ok {
			return "餐饮"
		}
	}

	if strings.Contains(text, "滴滴") || strings.Contains(text, "打车") ||
		strings.Contains(text, "地铁") || strings.Contains(text, "公交") ||
		strings.Contains(text, "加油") || strings.Contains(text, "停车") {
		if _, ok := categories["交通"]; ok {
			return "交通"
		}
	}

	if strings.Contains(text, "超市") || strings.Contains(text, "便利店") ||
		strings.Contains(text, "淘宝") || strings.Contains(text, "京东") ||
		strings.Contains(text, "拼多多") || strings.Contains(text, "天猫") {
		if _, ok := categories["购物"]; ok {
			return "购物"
		}
	}

	if strings.Contains(text, "房租") || strings.Contains(text, "物业") ||
		strings.Contains(text, "水电") || strings.Contains(text, "燃气") {
		if _, ok := categories["住房"]; ok {
			return "住房"
		}
	}

	if strings.Contains(text, "电影") || strings.Contains(text, "游戏") ||
		strings.Contains(text, "视频") || strings.Contains(text, "音乐") ||
		strings.Contains(text, "腾讯") || strings.Contains(text, "爱奇艺") ||
		strings.Contains(text, "网易云") {
		if _, ok := categories["娱乐"]; ok {
			return "娱乐"
		}
	}

	return "其他支出"
}

func inferCategoryFromAlipay(counterparty, product string, categories map[string]string) string {
	text := strings.ToLower(counterparty + product)

	if strings.Contains(text, "餐饮") || strings.Contains(text, "外卖") ||
		strings.Contains(text, "餐厅") || strings.Contains(text, "美团") ||
		strings.Contains(text, "饿了么") || strings.Contains(text, "口碑") {
		if _, ok := categories["餐饮"]; ok {
			return "餐饮"
		}
	}

	if strings.Contains(text, "滴滴") || strings.Contains(text, "打车") ||
		strings.Contains(text, "地铁") || strings.Contains(text, "公交") ||
		strings.Contains(text, "加油") || strings.Contains(text, "ETC") {
		if _, ok := categories["交通"]; ok {
			return "交通"
		}
	}

	if strings.Contains(text, "超市") || strings.Contains(text, "便利店") ||
		strings.Contains(text, "淘宝") || strings.Contains(text, "天猫") ||
		strings.Contains(text, "京东") || strings.Contains(text, "盒马") {
		if _, ok := categories["购物"]; ok {
			return "购物"
		}
	}

	if strings.Contains(text, "水电") || strings.Contains(text, "燃气") ||
		strings.Contains(text, "话费") || strings.Contains(text, "宽带") {
		if _, ok := categories["生活缴费"]; ok {
			return "生活缴费"
		}
	}

	return "其他支出"
}

func checkRecordExists(record *models.ExportRecord) (bool, error) {
	count, err := dbdata.ORM.Where("date = ? AND category = ? AND amount = ? AND type = ?",
		record.Date, record.Category, record.Amount, record.Type).
		Count(new(models.Record))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
