package service

import (
	"fmt"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// TransferAdder 转账服务需要注入记录添加能力（避免循环引用）
type TransferAdder interface {
	Add(recordType, category, note, date string, amount float64) (int64, error)
}

// TransferService 转账业务逻辑层
type TransferService struct {
	recordService TransferAdder
}

// NewTransferService 创建转账服务实例
func NewTransferService() *TransferService {
	return &TransferService{}
}

// SetRecordService 注入 RecordService
func (s *TransferService) SetRecordService(svc TransferAdder) {
	s.recordService = svc
}

// Create 创建转账
func (s *TransferService) Create(fromAccount, toAccount int64, amount float64, note, date string) (*models.Transfer, error) {
	if fromAccount == toAccount {
		return nil, fmt.Errorf("源账户和目标账户不能相同")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("转账金额必须大于0")
	}

	// 查询账户名称
	fromAcc := new(models.Account)
	if found, err := dbdata.ORM.ID(fromAccount).Get(fromAcc); err != nil || !found {
		return nil, fmt.Errorf("源账户不存在")
	}
	toAcc := new(models.Account)
	if found, err := dbdata.ORM.ID(toAccount).Get(toAcc); err != nil || !found {
		return nil, fmt.Errorf("目标账户不存在")
	}

	session := dbdata.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	// 创建转账记录
	transfer := &models.Transfer{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Amount:      amount,
		Note:        note,
		Date:        date,
	}

	if _, err := session.Insert(transfer); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("创建转账记录失败: %w", err)
	}

	// 生成转出记录（支出类）
	noteOut := "转账至" + toAcc.Name
	if note != "" {
		noteOut += " - " + note
	}
	outRecord := &models.Record{
		Type:      "expense",
		Amount:    amount,
		Category:  "转账",
		Note:      noteOut,
		Date:      date,
		AccountID: fromAccount,
	}
	if _, err := session.Insert(outRecord); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("创建转出记录失败: %w", err)
	}

	// 生成转入记录（收入类）
	noteIn := "来自" + fromAcc.Name + "的转账"
	if note != "" {
		noteIn += " - " + note
	}
	inRecord := &models.Record{
		Type:      "income",
		Amount:    amount,
		Category:  "转账",
		Note:      noteIn,
		Date:      date,
		AccountID: toAccount,
	}
	if _, err := session.Insert(inRecord); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("创建转入记录失败: %w", err)
	}

	// 回写关联的 Record ID
	transfer.RecordOutID = outRecord.ID
	transfer.RecordInID = inRecord.ID
	if _, err := session.ID(transfer.ID).Cols("record_out_id", "record_in_id").Update(transfer); err != nil {
		session.Rollback()
		return nil, fmt.Errorf("更新转账关联ID失败: %w", err)
	}

	return transfer, session.Commit()
}

// GetAll 获取所有转账记录
func (s *TransferService) GetAll() ([]models.Transfer, error) {
	transfers := make([]models.Transfer, 0)
	if err := dbdata.ORM.OrderBy("date DESC, created_at DESC").Find(&transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

// Delete 删除转账记录
func (s *TransferService) Delete(id int64) error {
	transfer := new(models.Transfer)
	found, err := dbdata.ORM.ID(id).Get(transfer)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("转账记录不存在")
	}

	session := dbdata.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 删除转账记录
	if _, err := session.ID(id).Delete(new(models.Transfer)); err != nil {
		session.Rollback()
		return fmt.Errorf("删除转账记录失败: %w", err)
	}

	// 删除关联的记账记录
	if transfer.RecordOutID > 0 {
		if _, err := session.ID(transfer.RecordOutID).Delete(new(models.Record)); err != nil {
			session.Rollback()
			return fmt.Errorf("删除转出记录失败: %w", err)
		}
	} else {
		if _, err := session.Where("type = ? AND category = ? AND amount = ? AND date = ? AND account_id = ?",
			"expense", "转账", transfer.Amount, transfer.Date, transfer.FromAccount).
			Delete(new(models.Record)); err != nil {
			session.Rollback()
			return fmt.Errorf("删除转出记录失败: %w", err)
		}
	}

	if transfer.RecordInID > 0 {
		if _, err := session.ID(transfer.RecordInID).Delete(new(models.Record)); err != nil {
			session.Rollback()
			return fmt.Errorf("删除转入记录失败: %w", err)
		}
	} else {
		if _, err := session.Where("type = ? AND category = ? AND amount = ? AND date = ? AND account_id = ?",
			"income", "转账", transfer.Amount, transfer.Date, transfer.ToAccount).
			Delete(new(models.Record)); err != nil {
			session.Rollback()
			return fmt.Errorf("删除转入记录失败: %w", err)
		}
	}

	return session.Commit()
}

// GetByDateRange 按日期范围获取转账记录
func (s *TransferService) GetByDateRange(startDate, endDate string) ([]models.Transfer, error) {
	transfers := make([]models.Transfer, 0)
	session := dbdata.NewSession()
	defer session.Close()

	if startDate != "" && endDate != "" {
		session = session.Where("date >= ? AND date <= ?", startDate, endDate)
	}
	session = session.OrderBy("date DESC, created_at DESC")

	if err := session.Find(&transfers); err != nil {
		return nil, fmt.Errorf("查询转账记录失败: %w", err)
	}
	return transfers, nil
}
