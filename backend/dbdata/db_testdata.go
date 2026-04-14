package dbdata

import (
	"fmt"
	"time"

	"piggy-accounting/backend/models"

	"xorm.io/xorm"
)

// seedTestData 在新账本中写入完整的测试数据，让用户可以体验所有功能
// 数据来源：用户手动录入的真实测试数据，覆盖 1-4 月多月份记录
// 仅在 ensureDefaultLedger 创建新账本时调用，使用传入的 engine 而非全局 ORM
func seedTestData(eng *xorm.Engine) error {
	session := eng.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	now := time.Now()
	year := now.Year()

	// ============ 1. 类别（19 个系统类别） ============
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
	for i := range categories {
		if _, err := session.Insert(&categories[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入类别失败: %w", err)
		}
	}

	// ============ 2. 多账户（4 个） ============
	accounts := []models.Account{
		{Name: "现金", Icon: "💵", Balance: 500, IsDefault: true, SortOrder: 0},
		{Name: "微信", Icon: "💬", Balance: 3250, IsDefault: false, SortOrder: 1},
		{Name: "支付宝", Icon: "🔵", Balance: 15800.50, IsDefault: false, SortOrder: 2},
		{Name: "银行卡", Icon: "🏦", Balance: 28600, IsDefault: false, SortOrder: 3},
	}
	for i := range accounts {
		if _, err := session.Insert(&accounts[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入账户失败: %w", err)
		}
	}

	// ============ 3. 标签（5 个） ============
	tags := []models.Tag{
		{Name: "日常", Color: "#6366f1"},
		{Name: "工作", Color: "#f59e0b"},
		{Name: "社交", Color: "#ec4899"},
		{Name: "健康", Color: "#10b981"},
		{Name: "学习", Color: "#3b82f6"},
	}
	for i := range tags {
		if _, err := session.Insert(&tags[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入标签失败: %w", err)
		}
	}

	// ============ 4. 记账记录（70 条，覆盖 1-4 月） ============
	// accountID: 1=现金 2=微信 3=支付宝 4=银行卡
	records := []models.Record{
		// ===== 1 月 =====
		{Type: "expense", Amount: 2500, Category: "住房", Note: "房租", Date: "2026-01-01", AccountID: 4},
		{Type: "expense", Amount: 28, Category: "餐饮", Note: "午餐", Date: "2026-01-03", AccountID: 2},
		{Type: "expense", Amount: 599, Category: "购物", Note: "机械键盘", Date: "2026-01-05", AccountID: 3},
		{Type: "expense", Amount: 120, Category: "娱乐", Note: "游戏充值", Date: "2026-01-08", AccountID: 3},
		{Type: "expense", Amount: 75, Category: "餐饮", Note: "朋友聚餐", Date: "2026-01-10", AccountID: 2},
		{Type: "expense", Amount: 380, Category: "医疗", Note: "感冒药", Date: "2026-01-12", AccountID: 4},
		{Type: "expense", Amount: 200, Category: "服饰", Note: "围巾", Date: "2026-01-15", AccountID: 3},
		{Type: "expense", Amount: 90, Category: "餐饮", Note: "日料自助", Date: "2026-01-18", AccountID: 2},
		{Type: "expense", Amount: 50, Category: "交通", Note: "加油", Date: "2026-01-20", AccountID: 1},
		{Type: "expense", Amount: 300, Category: "教育", Note: "网课", Date: "2026-01-22", AccountID: 3},
		{Type: "expense", Amount: 55, Category: "宠物", Note: "驱虫药", Date: "2026-01-25", AccountID: 3},
		{Type: "expense", Amount: 260, Category: "购物", Note: "年货", Date: "2026-01-28", AccountID: 3},
		{Type: "income", Amount: 15000, Category: "工资", Note: "1月工资", Date: "2026-01-05", AccountID: 4},
		{Type: "income", Amount: 500, Category: "投资", Note: "基金收益", Date: "2026-01-15", AccountID: 4},

		// ===== 2 月 =====
		{Type: "expense", Amount: 2500, Category: "住房", Note: "房租", Date: "2026-02-01", AccountID: 4},
		{Type: "expense", Amount: 1288, Category: "餐饮", Note: "春节聚餐", Date: "2026-02-01", AccountID: 2},
		{Type: "expense", Amount: 600, Category: "其他支出", Note: "红包", Date: "2026-02-02", AccountID: 1},
		{Type: "expense", Amount: 320, Category: "服饰", Note: "新年衣服", Date: "2026-02-03", AccountID: 3},
		{Type: "expense", Amount: 2800, Category: "购物", Note: "给爸妈的年货", Date: "2026-02-04", AccountID: 3},
		{Type: "expense", Amount: 65, Category: "交通", Note: "打车回老家", Date: "2026-02-05", AccountID: 2},
		{Type: "expense", Amount: 42, Category: "餐饮", Note: "外卖", Date: "2026-02-10", AccountID: 2},
		{Type: "expense", Amount: 88, Category: "娱乐", Note: "密室逃脱", Date: "2026-02-14", AccountID: 2},
		{Type: "expense", Amount: 200, Category: "运动", Note: "健身月卡", Date: "2026-02-15", AccountID: 3},
		{Type: "expense", Amount: 35, Category: "餐饮", Note: "下午茶", Date: "2026-02-18", AccountID: 2},
		{Type: "expense", Amount: 150, Category: "通讯", Note: "宽带年费分摊", Date: "2026-02-20", AccountID: 3},
		{Type: "income", Amount: 15000, Category: "工资", Note: "2月工资", Date: "2026-02-05", AccountID: 4},
		{Type: "income", Amount: 2000, Category: "奖金", Note: "年终奖", Date: "2026-02-10", AccountID: 4},
		{Type: "income", Amount: 800, Category: "红包", Note: "收到的红包", Date: "2026-02-02", AccountID: 2},

		// ===== 3 月 =====
		{Type: "expense", Amount: 2500, Category: "住房", Note: "房租", Date: "2026-03-01", AccountID: 4},
		{Type: "expense", Amount: 30, Category: "餐饮", Note: "食堂", Date: "2026-03-03", AccountID: 1},
		{Type: "expense", Amount: 45, Category: "餐饮", Note: "外卖披萨", Date: "2026-03-05", AccountID: 2},
		{Type: "expense", Amount: 399, Category: "购物", Note: "运动鞋", Date: "2026-03-07", AccountID: 3},
		{Type: "expense", Amount: 22, Category: "娱乐", Note: "电影票", Date: "2026-03-08", AccountID: 2},
		{Type: "expense", Amount: 68, Category: "餐饮", Note: "火锅", Date: "2026-03-10", AccountID: 2},
		{Type: "expense", Amount: 350, Category: "医疗", Note: "体检", Date: "2026-03-12", AccountID: 4},
		{Type: "expense", Amount: 85, Category: "通讯", Note: "手机套餐", Date: "2026-03-15", AccountID: 3},
		{Type: "expense", Amount: 520, Category: "购物", Note: "衣服", Date: "2026-03-18", AccountID: 3},
		{Type: "expense", Amount: 38, Category: "餐饮", Note: "日料", Date: "2026-03-20", AccountID: 2},
		{Type: "expense", Amount: 150, Category: "教育", Note: "书籍", Date: "2026-03-22", AccountID: 3},
		{Type: "expense", Amount: 55, Category: "宠物", Note: "猫罐头", Date: "2026-03-25", AccountID: 3},
		{Type: "income", Amount: 15000, Category: "工资", Note: "3月工资", Date: "2026-03-05", AccountID: 4},
		{Type: "income", Amount: 1200, Category: "奖金", Note: "季度奖金", Date: "2026-03-10", AccountID: 4},
		{Type: "income", Amount: 100, Category: "红包", Note: "微信红包", Date: "2026-03-15", AccountID: 2},

		// ===== 4 月（当月，日期根据今天动态调整，确保不超过今天） =====
		{Type: "expense", Amount: 28, Category: "餐饮", Note: "外卖麻辣烫", Date: "2026-04-01", AccountID: 2},
		{Type: "expense", Amount: 35.5, Category: "餐饮", Note: "午餐烤肉饭", Date: "2026-04-02", AccountID: 1},
		{Type: "expense", Amount: 6, Category: "交通", Note: "地铁", Date: "2026-04-02", AccountID: 2},
		{Type: "expense", Amount: 199, Category: "购物", Note: "护肤品", Date: "2026-04-03", AccountID: 3},
		{Type: "expense", Amount: 42, Category: "餐饮", Note: "和朋友吃火锅", Date: "2026-04-04", AccountID: 2},
		{Type: "expense", Amount: 15, Category: "娱乐", Note: "视频会员续费", Date: "2026-04-05", AccountID: 3},
		{Type: "expense", Amount: 8.5, Category: "交通", Note: "打车", Date: "2026-04-06", AccountID: 2},
		{Type: "expense", Amount: 120, Category: "服饰", Note: "T恤", Date: "2026-04-07", AccountID: 3},
		{Type: "expense", Amount: 56, Category: "餐饮", Note: "超市食材", Date: "2026-04-07", AccountID: 1},
		{Type: "expense", Amount: 2500, Category: "住房", Note: "房租", Date: "2026-04-08", AccountID: 4},
		{Type: "expense", Amount: 180, Category: "通讯", Note: "话费充值", Date: "2026-04-08", AccountID: 3},
		{Type: "expense", Amount: 32, Category: "餐饮", Note: "咖啡", Date: "2026-04-09", AccountID: 2},
		{Type: "expense", Amount: 450, Category: "医疗", Note: "牙科检查", Date: "2026-04-09", AccountID: 4},
		{Type: "expense", Amount: 88, Category: "宠物", Note: "猫粮", Date: "2026-04-10", AccountID: 3},
		{Type: "expense", Amount: 25, Category: "餐饮", Note: "早餐+午餐", Date: "2026-04-10", AccountID: 1},
		{Type: "expense", Amount: 150, Category: "教育", Note: "在线课程", Date: "2026-04-11", AccountID: 3},
		{Type: "expense", Amount: 68, Category: "餐饮", Note: "烧烤", Date: "2026-04-11", AccountID: 2},
		{Type: "expense", Amount: 4, Category: "交通", Note: "公交", Date: "2026-04-11", AccountID: 1},
		{Type: "expense", Amount: 299, Category: "购物", Note: "耳机", Date: "2026-04-12", AccountID: 3},
		{Type: "expense", Amount: 45, Category: "运动", Note: "羽毛球场地", Date: "2026-04-12", AccountID: 2},
		{Type: "expense", Amount: 18, Category: "餐饮", Note: "奶茶", Date: "2026-04-12", AccountID: 2},
		{Type: "income", Amount: 15000, Category: "工资", Note: "4月工资", Date: "2026-04-05", AccountID: 4},
		{Type: "income", Amount: 2000, Category: "兼职", Note: "周末兼职翻译", Date: "2026-04-06", AccountID: 3},
		{Type: "income", Amount: 500, Category: "红包", Note: "生日红包", Date: "2026-04-08", AccountID: 2},
		{Type: "income", Amount: 200, Category: "退款", Note: "退货退款", Date: "2026-04-10", AccountID: 3},
	}

	// 4 月 13 日的记录：只在今天 >= 13 时插入
	if now.Day() >= 13 {
		records = append(records,
			models.Record{Type: "expense", Amount: 7.5, Category: "交通", Note: "地铁", Date: "2026-04-13", AccountID: 2},
			models.Record{Type: "expense", Amount: 36, Category: "餐饮", Note: "外卖寿司", Date: "2026-04-13", AccountID: 2},
		)
	}

	// 插入所有记录
	for i := range records {
		if _, err := session.Insert(&records[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入测试记录失败: %w", err)
		}
	}

	// ============ 5. 记录-标签关联（9 条） ============
	// 标签 IDs: 日常=1, 工作=2, 社交=3, 健康=4, 学习=5
	// 记录 IDs（按插入顺序）：
	//   id=2: 餐饮/午餐/01-03
	//   id=3: 购物/机械键盘/01-05
	//   id=4: 娱乐/游戏充值/01-08
	//   id=5: 餐饮/朋友聚餐/01-10
	//   id=6: 医疗/感冒药/01-12
	//   id=48: 餐饮/和朋友吃火锅/04-04
	//   id=52: 餐饮/超市食材/04-07
	//   id=55: 餐饮/咖啡/04-09
	//   id=59: 教育/在线课程/04-11
	recordTags := []models.RecordTag{
		{RecordID: 2, TagID: 1},  // 午餐 → 日常
		{RecordID: 3, TagID: 1},  // 机械键盘 → 日常
		{RecordID: 4, TagID: 2},  // 游戏充值 → 工作（摸鱼）
		{RecordID: 5, TagID: 3},  // 朋友聚餐 → 社交
		{RecordID: 6, TagID: 1},  // 感冒药 → 日常
		{RecordID: 48, TagID: 3}, // 和朋友吃火锅(4月) → 社交
		{RecordID: 52, TagID: 4}, // 超市食材(4月) → 健康
		{RecordID: 55, TagID: 2}, // 咖啡(4月) → 工作
		{RecordID: 59, TagID: 5}, // 在线课程(4月) → 学习
	}
	for i := range recordTags {
		if _, err := session.Insert(&recordTags[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入记录标签关联失败: %w", err)
		}
	}

	// ============ 6. 预算（4 个月度 + 1 个年度） ============
	budgets := []models.Budget{
		{Type: "monthly", Year: year, Month: 1, Amount: 5000},
		{Type: "monthly", Year: year, Month: 2, Amount: 8000},
		{Type: "monthly", Year: year, Month: 3, Amount: 5000},
		{Type: "monthly", Year: year, Month: int(now.Month()), Amount: 5000},
		{Type: "yearly", Year: year, Month: 0, Amount: 60000},
	}
	for i := range budgets {
		if _, err := session.Insert(&budgets[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入预算失败: %w", err)
		}
	}

	// ============ 7. 待办任务（5 个，含 1 个已完成） ============
	tasks := []models.Task{
		{Title: "还信用卡", DueDate: "2026-04-15", Amount: 3500, Completed: false},
		{Title: "交水电费", DueDate: "2026-04-18", Amount: 280, Completed: false},
		{Title: "续费会员", DueDate: "2026-04-20", Amount: 15, Completed: true},
		{Title: "买猫砂", DueDate: "2026-04-25", Amount: 50, Completed: false},
		{Title: "季度理财", DueDate: "2026-04-30", Amount: 5000, Completed: false},
	}
	for i := range tasks {
		if _, err := session.Insert(&tasks[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入任务失败: %w", err)
		}
	}

	// ============ 8. 转账记录（2 笔） ============
	// 转账1: 银行卡 → 支付宝（生活费）
	transfer1 := &models.Transfer{
		FromAccount: 4, ToAccount: 3, Amount: 2000, Note: "生活费", Date: "2026-04-05",
	}
	if _, err := session.Insert(transfer1); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账记录失败: %w", err)
	}
	outRec1 := &models.Record{Type: "expense", Amount: 2000, Category: "转账", Note: "转账至支付宝 - 生活费", Date: "2026-04-05", AccountID: 4}
	if _, err := session.Insert(outRec1); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账支出记录失败: %w", err)
	}
	inRec1 := &models.Record{Type: "income", Amount: 2000, Category: "转账", Note: "来自银行卡的转账 - 生活费", Date: "2026-04-05", AccountID: 3}
	if _, err := session.Insert(inRec1); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账收入记录失败: %w", err)
	}
	transfer1.RecordOutID = outRec1.ID
	transfer1.RecordInID = inRec1.ID
	if _, err := session.ID(transfer1.ID).Cols("record_out_id", "record_in_id").Update(transfer1); err != nil {
		session.Rollback()
		return fmt.Errorf("更新转账关联ID失败: %w", err)
	}

	// 转账2: 银行卡 → 微信（零花钱）
	transfer2 := &models.Transfer{
		FromAccount: 4, ToAccount: 2, Amount: 500, Note: "零花钱", Date: "2026-04-05",
	}
	if _, err := session.Insert(transfer2); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账记录2失败: %w", err)
	}
	outRec2 := &models.Record{Type: "expense", Amount: 500, Category: "转账", Note: "转账至微信 - 零花钱", Date: "2026-04-05", AccountID: 4}
	if _, err := session.Insert(outRec2); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账2支出记录失败: %w", err)
	}
	inRec2 := &models.Record{Type: "income", Amount: 500, Category: "转账", Note: "来自银行卡的转账 - 零花钱", Date: "2026-04-05", AccountID: 2}
	if _, err := session.Insert(inRec2); err != nil {
		session.Rollback()
		return fmt.Errorf("插入转账2收入记录失败: %w", err)
	}
	transfer2.RecordOutID = outRec2.ID
	transfer2.RecordInID = inRec2.ID
	if _, err := session.ID(transfer2.ID).Cols("record_out_id", "record_in_id").Update(transfer2); err != nil {
		session.Rollback()
		return fmt.Errorf("更新转账2关联ID失败: %w", err)
	}

	// ============ 9. 周期记账规则（3 条） ============
	recurringRecords := []models.RecurringRecord{
		{Type: "expense", Amount: 2500, Category: "住房", Note: "房租", AccountID: 4,
			Frequency: "monthly", MonthDay: 1, NextDate: "2026-05-01", Enabled: true, LastRunDate: "2026-04-01"},
		{Type: "expense", Amount: 15, Category: "娱乐", Note: "视频会员", AccountID: 3,
			Frequency: "monthly", MonthDay: 5, NextDate: "2026-05-05", Enabled: true, LastRunDate: "2026-04-05"},
		{Type: "expense", Amount: 180, Category: "通讯", Note: "话费", AccountID: 3,
			Frequency: "monthly", MonthDay: 8, NextDate: "2026-05-08", Enabled: true, LastRunDate: "2026-04-08"},
	}
	for i := range recurringRecords {
		if _, err := session.Insert(&recurringRecords[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入周期记账规则失败: %w", err)
		}
	}

	// ============ 10. 提醒设置（5 条 + 1 条全局设置） ============
	reminders := []models.Reminder{
		{Type: "budget_warning", BudgetType: "monthly", Threshold: 80, Enabled: true, Message: "本月预算已使用超过80%"},
		{Type: "budget_warning", BudgetType: "yearly", Threshold: 80, Enabled: true, Message: "本年度预算已使用超过80%"},
		{Type: "daily_reminder", Enabled: true, Message: "记得记账哦～", ReminderTime: "20:00"},
		{Type: "weekly_summary", Enabled: true, Message: "本周消费汇总"},
		{Type: "task_reminder", Enabled: true, Message: "您有待办任务即将到期"},
	}
	for i := range reminders {
		if _, err := session.Insert(&reminders[i]); err != nil {
			session.Rollback()
			return fmt.Errorf("插入提醒设置失败: %w", err)
		}
	}

	reminderSettings := &models.ReminderSettings{
		WebhookURL:                "",
		WebhookEnabled:            false,
		PopupEnabled:              true,
		SystemNotificationEnabled: true,
		TaskReminderDays:          3,
	}
	if _, err := session.Insert(reminderSettings); err != nil {
		session.Rollback()
		return fmt.Errorf("插入提醒全局设置失败: %w", err)
	}

	return session.Commit()
}
