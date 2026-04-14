package service

import (
	"fmt"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// BudgetService 预算业务逻辑层
type BudgetService struct{}

// NewBudgetService 创建预算服务实例
func NewBudgetService() *BudgetService {
	return &BudgetService{}
}

// SetBudget 设置预算
func (s *BudgetService) SetBudget(budgetType string, year, month int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("预算金额必须大于0")
	}
	if budgetType != "monthly" && budgetType != "yearly" {
		return fmt.Errorf("类型必须是 monthly 或 yearly")
	}
	if budgetType == "monthly" && (month < 1 || month > 12) {
		return fmt.Errorf("月度预算月份必须在1-12之间")
	}

	budget := &models.Budget{
		Type:   budgetType,
		Year:   year,
		Month:  month,
		Amount: amount,
	}

	// 查找已有记录
	existing := &models.Budget{}
	found, err := dbdata.ORM.Where("type = ? AND year = ? AND month = ?", budgetType, year, month).Get(existing)
	if err != nil {
		return fmt.Errorf("查询预算失败: %w", err)
	}

	if found {
		_, err = dbdata.ORM.ID(existing.ID).Cols("amount").Update(budget)
		if err != nil {
			return fmt.Errorf("更新预算失败: %w", err)
		}
	} else {
		_, err = dbdata.ORM.Insert(budget)
		if err != nil {
			return fmt.Errorf("添加预算失败: %w", err)
		}
	}

	return nil
}

// GetBudgetInfo 获取预算信息（含支出统计和进度）
func (s *BudgetService) GetBudgetInfo(budgetType string, year, month int) (*models.BudgetInfo, error) {
	// 1. 查询预算
	budget := &models.Budget{}
	found, err := dbdata.ORM.Where("type = ? AND year = ? AND month = ?", budgetType, year, month).Get(budget)
	if err != nil {
		return nil, fmt.Errorf("查询预算失败: %w", err)
	}
	if !found {
		return nil, nil // 没设预算
	}

	// 2. 计算日期范围
	var startDate, endDate string
	if budgetType == "monthly" {
		startDate = fmt.Sprintf("%d-%02d-01", year, month)
		switch month {
		case 1, 3, 5, 7, 8, 10, 12:
			endDate = fmt.Sprintf("%d-%02d-31", year, month)
		case 4, 6, 9, 11:
			endDate = fmt.Sprintf("%d-%02d-30", year, month)
		case 2:
			if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
				endDate = fmt.Sprintf("%d-%02d-29", year, month)
			} else {
				endDate = fmt.Sprintf("%d-%02d-28", year, month)
			}
		}
	} else {
		startDate = fmt.Sprintf("%d-01-01", year)
		endDate = fmt.Sprintf("%d-12-31", year)
	}

	// 3. 查询该范围内的总支出
	spent, err := dbdata.ORM.Where("type = ? AND date >= ? AND date <= ?", "expense", startDate, endDate).
		Sum(new(models.Record), "amount")
	if err != nil {
		return nil, fmt.Errorf("查询支出失败: %w", err)
	}

	// 4. 计算百分比
	percentage := 0.0
	if budget.Amount > 0 {
		percentage = spent / budget.Amount * 100
		if percentage > 100 {
			percentage = 100
		}
	}

	return &models.BudgetInfo{
		BudgetType:   budgetType,
		Year:         year,
		Month:        month,
		BudgetAmount: budget.Amount,
		Spent:        spent,
		Remaining:    budget.Amount - spent,
		Percentage:   percentage,
	}, nil
}

// DeleteBudget 删除预算
func (s *BudgetService) DeleteBudget(budgetType string, year, month int) error {
	affected, err := dbdata.ORM.Where("type = ? AND year = ? AND month = ?", budgetType, year, month).
		Delete(new(models.Budget))
	if err != nil {
		return fmt.Errorf("删除预算失败: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("预算不存在")
	}
	return nil
}
