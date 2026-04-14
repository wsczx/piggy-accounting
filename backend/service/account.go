package service

import (
	"fmt"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// AccountService 账户业务逻辑层
type AccountService struct{}

// NewAccountService 创建账户服务实例
func NewAccountService() *AccountService {
	return &AccountService{}
}

// GetAll 获取所有账户
func (s *AccountService) GetAll() ([]models.Account, error) {
	accounts := make([]models.Account, 0)
	if err := dbdata.ORM.OrderBy("sort_order ASC, id ASC").Find(&accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAllWithBalance 获取所有账户及实时余额
func (s *AccountService) GetAllWithBalance() ([]models.AccountInfo, error) {
	accounts, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]models.AccountInfo, 0, len(accounts))
	for _, acc := range accounts {
		realBalance, err := s.CalculateBalance(acc.ID)
		if err != nil {
			realBalance = acc.Balance // 出错时回退到初始余额
		}
		result = append(result, models.AccountInfo{
			ID:          acc.ID,
			Name:        acc.Name,
			Icon:        acc.Icon,
			Balance:     acc.Balance,
			RealBalance: acc.Balance + realBalance,
			IsDefault:   acc.IsDefault,
		})
	}
	return result, nil
}

// CalculateBalance 计算账户的流水余额（收入 - 支出）
// 注意：转账已在 records 表中以 expense/income 记录体现，无需额外计算 transfers 表
func (s *AccountService) CalculateBalance(accountID int64) (float64, error) {
	var incomeSum, expenseSum float64

	if _, err := dbdata.ORM.Table("records").
		Where("account_id = ? AND type = ?", accountID, "income").
		Sums(&incomeSum); err != nil {
		return 0, fmt.Errorf("计算收入失败: %w", err)
	}

	if _, err := dbdata.ORM.Table("records").
		Where("account_id = ? AND type = ?", accountID, "expense").
		Sums(&expenseSum); err != nil {
		return 0, fmt.Errorf("计算支出失败: %w", err)
	}

	return incomeSum - expenseSum, nil
}

// Create 创建账户
func (s *AccountService) Create(name, icon string, balance float64, isDefault bool) (*models.Account, error) {
	if name == "" {
		return nil, fmt.Errorf("账户名称不能为空")
	}

	account := &models.Account{
		Name:      name,
		Icon:      icon,
		Balance:   balance,
		IsDefault: isDefault,
		SortOrder: 0,
	}

	session := dbdata.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	if isDefault {
		// 取消其他默认账户
		if _, err := session.Where("is_default = ?", true).Cols("is_default").Update(&models.Account{IsDefault: false}); err != nil {
			session.Rollback()
			return nil, fmt.Errorf("设置默认账户失败: %w", err)
		}
	}

	if _, err := session.Insert(account); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("创建账户失败: %w", err)
	}

	// 如果没有其他账户，自动设为默认
	if c, _ := session.Where("id != ?", account.ID).Count(new(models.Account)); c == 0 {
		account.IsDefault = true
		session.ID(account.ID).Cols("is_default").Update(account)
	}

	return account, session.Commit()
}

// Update 更新账户
func (s *AccountService) Update(id int64, name, icon string, balance float64, isDefault bool) error {
	session := dbdata.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	if isDefault {
		if _, err := session.Where("is_default = ? AND id != ?", true, id).Cols("is_default").Update(&models.Account{IsDefault: false}); err != nil {
			session.Rollback()
			return fmt.Errorf("设置默认账户失败: %w", err)
		}
	}

	if _, err := session.ID(id).Update(&models.Account{
		Name:      name,
		Icon:      icon,
		Balance:   balance,
		IsDefault: isDefault,
	}); err != nil {
		session.Rollback()
		return fmt.Errorf("更新账户失败: %w", err)
	}

	return session.Commit()
}

// Delete 删除账户
func (s *AccountService) Delete(id int64) error {
	// 检查是否有记录关联此账户
	if c, err := dbdata.ORM.Where("account_id = ?", id).Count(new(models.Record)); err == nil && c > 0 {
		return fmt.Errorf("该账户下有 %d 条记录，无法删除", c)
	}

	// 检查是否有转账记录关联
	if c, err := dbdata.ORM.Where("from_account = ? OR to_account = ?", id, id).Count(new(models.Transfer)); err == nil && c > 0 {
		return fmt.Errorf("该账户下有转账记录，无法删除")
	}

	_, err := dbdata.ORM.ID(id).Delete(new(models.Account))
	return err
}

// GetTotalAssets 获取总资产
func (s *AccountService) GetTotalAssets() (float64, error) {
	infos, err := s.GetAllWithBalance()
	if err != nil {
		return 0, err
	}
	var total float64
	for _, info := range infos {
		total += info.RealBalance
	}
	return total, nil
}
