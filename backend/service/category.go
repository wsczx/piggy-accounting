package service

import (
	"fmt"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// CategoryService 类别业务逻辑层
type CategoryService struct{}

// NewCategoryService 创建类别服务实例
func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

// GetAll 获取所有类别
func (s *CategoryService) GetAll() ([]models.Category, error) {
	categories := make([]models.Category, 0)
	session := dbdata.ORM.OrderBy("is_system DESC, id ASC")
	if err := session.Find(&categories); err != nil {
		return nil, fmt.Errorf("查询类别失败: %w", err)
	}
	return categories, nil
}

// GetByType 按类型获取类别
func (s *CategoryService) GetByType(categoryType string) ([]models.Category, error) {
	categories := make([]models.Category, 0)
	session := dbdata.ORM.Where("type = ?", categoryType).OrderBy("is_system DESC, id ASC")
	if err := session.Find(&categories); err != nil {
		return nil, fmt.Errorf("查询类别失败: %w", err)
	}
	return categories, nil
}

// Add 添加自定义类别
func (s *CategoryService) Add(name, icon, categoryType string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("类别名称不能为空")
	}
	if categoryType != "income" && categoryType != "expense" {
		return 0, fmt.Errorf("无效的类别类型")
	}

	// 检查重名
	count, err := dbdata.ORM.Where("name = ? AND type = ?", name, categoryType).Count(new(models.Category))
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, fmt.Errorf("类别 %s 已存在", name)
	}

	c := &models.Category{
		Name:     name,
		Icon:     icon,
		Type:     categoryType,
		IsSystem: false,
	}
	_, err = dbdata.ORM.Insert(c)
	if err != nil {
		return 0, fmt.Errorf("添加类别失败: %w", err)
	}
	return c.ID, nil
}

// Delete 删除自定义类别
func (s *CategoryService) Delete(id int64) error {
	if id <= 0 {
		return fmt.Errorf("无效的类别ID")
	}

	c := new(models.Category)
	found, err := dbdata.ORM.ID(id).Get(c)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("类别不存在")
	}
	if c.IsSystem {
		return fmt.Errorf("系统类别不可删除")
	}

	// 检查是否有记录使用此类别
	recordCount, err := dbdata.ORM.Where("category = ?", c.Name).Count(new(models.Record))
	if err != nil {
		return err
	}
	if recordCount > 0 {
		return fmt.Errorf("该类别下有 %d 条记录，无法删除", recordCount)
	}

	_, err = dbdata.ORM.ID(id).Delete(new(models.Category))
	return err
}

// GetCategoryIcon 获取类别图标
func (s *CategoryService) GetCategoryIcon(categoryName, categoryType string) string {
	c := new(models.Category)
	found, err := dbdata.ORM.Where("name = ? AND type = ?", categoryName, categoryType).Get(c)
	if err != nil || !found {
		if categoryType == "income" {
			return "💰"
		}
		return "📦"
	}
	return c.Icon
}
