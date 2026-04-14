package service

import (
	"errors"
	"time"

	"piggy-accounting/backend/models"
	"piggy-accounting/backend/dbdata"
)

// TagService 标签业务逻辑层
type TagService struct{}

// NewTagService 创建标签服务实例
func NewTagService() *TagService {
	return &TagService{}
}

// GetAllTags 获取所有标签
func (s *TagService) GetAllTags() ([]models.Tag, error) {
	tags := make([]models.Tag, 0)
	if err := dbdata.ORM.OrderBy("id ASC").Find(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// CreateTag 创建标签
func (s *TagService) CreateTag(name, color string) (*models.Tag, error) {
	if name == "" {
		return nil, errors.New("标签名称不能为空")
	}

	// 检查是否已存在
	exists, err := dbdata.ORM.Where("name = ?", name).Exist(new(models.Tag))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("标签已存在")
	}

	// 默认颜色
	if color == "" {
		color = "#6366f1"
	}

	tag := &models.Tag{
		Name:  name,
		Color: color,
	}

	_, err = dbdata.ORM.Insert(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(id int64, name, color string) error {
	tag := new(models.Tag)
	found, err := dbdata.ORM.ID(id).Get(tag)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("标签不存在")
	}

	if name != "" {
		// 检查新名称是否与其他标签冲突
		exists, err := dbdata.ORM.Where("name = ? AND id != ?", name, id).Exist(new(models.Tag))
		if err != nil {
			return err
		}
		if exists {
			return errors.New("标签名称已存在")
		}
		tag.Name = name
	}

	if color != "" {
		tag.Color = color
	}

	tag.UpdatedAt = time.Now()
	_, err = dbdata.ORM.ID(id).Update(tag)
	return err
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(id int64) error {
	// 先删除关联
	if _, err := dbdata.ORM.Where("tag_id = ?", id).Delete(new(models.RecordTag)); err != nil {
		return err
	}
	_, err := dbdata.ORM.ID(id).Delete(new(models.Tag))
	return err
}

// GetRecordTags 获取记录的标签
func (s *TagService) GetRecordTags(recordID int64) ([]models.Tag, error) {
	tags := make([]models.Tag, 0)
	err := dbdata.ORM.Table("tag").
		Join("INNER", "record_tag", "tag.id = record_tag.tag_id").
		Where("record_tag.record_id = ?", recordID).
		Find(&tags)
	return tags, err
}

// SetRecordTags 设置记录的标签（替换原有标签）
func (s *TagService) SetRecordTags(recordID int64, tagIDs []int64) error {
	session := dbdata.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 删除原有标签关联
	if _, err := session.Where("record_id = ?", recordID).Delete(&models.RecordTag{}); err != nil {
		session.Rollback()
		return err
	}

	// 添加新标签关联
	for _, tagID := range tagIDs {
		recordTag := &models.RecordTag{
			RecordID: recordID,
			TagID:    tagID,
		}
		if _, err := session.Insert(recordTag); err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}
