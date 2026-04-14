package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"
)

// RecordCreator 记录创建函数类型（供智能识别使用）
type RecordCreator func(record *models.Record) (int64, error)

// SmartRecognizeService 智能识别业务逻辑层
type SmartRecognizeService struct {
	recordService RecordCreator
}

// NewSmartRecognizeService 创建智能识别服务实例
func NewSmartRecognizeService() *SmartRecognizeService {
	return &SmartRecognizeService{}
}

// SetRecordServiceForRecognize 注入 RecordService（供识别后创建记录使用）
func (s *SmartRecognizeService) SetRecordServiceForRecognize(create RecordCreator) {
	s.recordService = create
}

// RecognizeText 智能识别文本
func (s *SmartRecognizeService) RecognizeText(text string) (*models.ParsedRecord, error) {
	return recognizeText(text)
}

// RecognizeAndCreate 识别并创建记录
func (s *SmartRecognizeService) RecognizeAndCreate(text string) (*models.Record, error) {
	parsed, err := recognizeText(text)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, nil
	}

	record := &models.Record{
		Type:     parsed.Type,
		Amount:   parsed.Amount,
		Category: parsed.Category,
		Note:     parsed.Note,
		Date:     parsed.Date,
	}

	if s.recordService != nil {
		_, err = s.recordService(record)
	} else {
		_, err = dbdata.ORM.Insert(record)
	}
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ==================== 内部函数 ====================

func recognizeText(text string) (*models.ParsedRecord, error) {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return nil, nil
	}

	record := &models.ParsedRecord{
		Date: time.Now().Format("2006-01-02"),
		Tags: []string{},
	}

	if amount := extractAmount(text); amount > 0 {
		record.Amount = amount
	}

	record.Type = recognizeType(text)
	record.Category = recognizeCategory(text, record.Type)
	record.Tags = recognizeTags(text)

	if date := extractDate(text); date != "" {
		record.Date = date
	}

	record.Note = generateNote(text)

	return record, nil
}

var categoryKeywords = map[string]map[string][]string{
	"expense": {
		"餐饮":     {"外卖", "美团", "饿了么", "餐厅", "饭店", "火锅", "烧烤", "奶茶", "咖啡", "早餐", "午餐", "晚餐", "吃饭", "food", "meal"},
		"交通":     {"地铁", "公交", "打车", "滴滴", "出租车", "加油", "停车", "高铁", "火车", "飞机", "ticket", "transport"},
		"购物":     {"超市", "便利店", "淘宝", "京东", "拼多多", "商场", "买衣服", "买鞋", "化妆品", "shopping"},
		"娱乐":     {"电影", "游戏", "KTV", "唱歌", "酒吧", "旅游", "旅行", "门票", "entertainment"},
		"居住":     {"房租", "水电", "物业", "宽带", "话费", "燃气", "rent", "utility"},
		"医疗":     {"医院", "药店", "看病", "买药", "体检", "medical", "hospital"},
		"教育":     {"学费", "培训", "课程", "书", "考试", "education", "book"},
		"人情":     {"红包", "礼物", "请客", "聚餐", "份子钱", "gift", "party"},
	},
	"income": {
		"工资":     {"工资", "薪水", "salary", "payroll"},
		"奖金":     {"奖金", "年终奖", "bonus", "award"},
		"投资":     {"理财", "股票", "基金", "利息", "分红", "investment", "stock"},
		"兼职":     {"兼职", "副业", "freelance", "part-time"},
		"红包":     {"红包", "转账", "red packet", "transfer"},
		"退款":     {"退款", "退货", "refund", "return"},
	},
}

var tagKeywords = map[string][]string{
	"出差":     {"出差", "差旅", "business trip"},
	"报销":     {"报销", "发票", "reimburse"},
	"聚餐":     {"聚餐", "请客", "AA", "dinner party"},
	"网购":     {"淘宝", "京东", "拼多多", "网购", "online"},
	"日常":     {"日常", "daily"},
	"紧急":     {"紧急", "urgent"},
}

func extractAmount(text string) float64 {
	patterns := []string{
		`(\d+\.?\d*)\s*[元块￥$]`,
		`[￥$]\s*(\d+\.?\d*)`,
		`(\d+\.?\d*)\s*yuan`,
		`花费?\s*(\d+\.?\d*)`,
		`支出?\s*(\d+\.?\d*)`,
		`收入?\s*(\d+\.?\d*)`,
		`(\d+\.?\d*)\s*元整`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			amount, _ := strconv.ParseFloat(matches[1], 64)
			if amount > 0 {
				return amount
			}
		}
	}

	return 0
}

func recognizeType(text string) string {
	incomeKeywords := []string{"收入", "收到", "到账", "工资", "奖金", "红包", "转账", "退款", "返现", "理财收益", "income", "received", "refund"}
	for _, kw := range incomeKeywords {
		if strings.Contains(text, kw) {
			return "income"
		}
	}

	expenseKeywords := []string{"支出", "花费", "消费", "支付", "付款", "买", "花", "expense", "spent", "pay"}
	for _, kw := range expenseKeywords {
		if strings.Contains(text, kw) {
			return "expense"
		}
	}

	return "expense"
}

func recognizeCategory(text, recordType string) string {
	categories, ok := categoryKeywords[recordType]
	if !ok {
		return "其他"
	}

	for category, keywords := range categories {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				return category
			}
		}
	}

	if recordType == "income" {
		return "其他收入"
	}
	return "其他"
}

func recognizeTags(text string) []string {
	tags := []string{}
	for tag, keywords := range tagKeywords {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				tags = append(tags, tag)
				break
			}
		}
	}
	return tags
}

func extractDate(text string) string {
	if strings.Contains(text, "今天") {
		return time.Now().Format("2006-01-02")
	}
	if strings.Contains(text, "昨天") {
		return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}
	if strings.Contains(text, "明天") {
		return time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	patterns := []string{
		`(\d{4})[-/](\d{1,2})[-/](\d{1,2})`,
		`(\d{1,2})月(\d{1,2})[日号]?`,
		`(\d{1,2})\.(\d{1,2})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) >= 3 {
			now := time.Now()
			year, month, day := now.Year(), now.Month(), now.Day()

			if len(matches) == 4 {
				y, _ := strconv.Atoi(matches[1])
				m, _ := strconv.Atoi(matches[2])
				d, _ := strconv.Atoi(matches[3])
				year, month, day = y, time.Month(m), d
			} else {
				m, _ := strconv.Atoi(matches[1])
				d, _ := strconv.Atoi(matches[2])
				month, day = time.Month(m), d
			}

			t := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
			return t.Format("2006-01-02")
		}
	}

	return ""
}

func generateNote(text string) string {
	note := text
	amountPatterns := []string{
		`\d+\.?\d*\s*[元块￥$]`,
		`[￥$]\s*\d+\.?\d*`,
		`\d+\.?\d*\s*yuan`,
		`花费?\s*\d+\.?\d*`,
		`支出?\s*\d+\.?\d*`,
		`收入?\s*\d+\.?\d*`,
	}
	for _, pattern := range amountPatterns {
		re := regexp.MustCompile(pattern)
		note = re.ReplaceAllString(note, "")
	}

	datePatterns := []string{
		`\d{4}[-/]\d{1,2}[-/]\d{1,2}`,
		`\d{1,2}月\d{1,2}[日号]?`,
		`今天|昨天|明天`,
	}
	for _, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		note = re.ReplaceAllString(note, "")
	}

	note = strings.TrimSpace(note)
	note = regexp.MustCompile(`\s+`).ReplaceAllString(note, " ")

	return note
}
