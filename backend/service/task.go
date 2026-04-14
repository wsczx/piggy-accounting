package service

import (
	"fmt"

	"piggy-accounting/backend/dbdata"
	"piggy-accounting/backend/models"
)

// TaskService 待办任务业务逻辑层
type TaskService struct{}

// NewTaskService 创建任务服务实例
func NewTaskService() *TaskService {
	return &TaskService{}
}

// GetAll 获取所有任务
func (s *TaskService) GetAll() ([]models.Task, error) {
	tasks := make([]models.Task, 0)
	if err := dbdata.ORM.OrderBy("completed ASC, due_date ASC").Find(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Create 创建任务
func (s *TaskService) Create(title, dueDate string, amount float64) (*models.Task, error) {
	if title == "" {
		return nil, fmt.Errorf("任务名称不能为空")
	}
	if dueDate == "" {
		return nil, fmt.Errorf("到期日期不能为空")
	}

	task := &models.Task{
		Title:     title,
		DueDate:   dueDate,
		Amount:    amount,
		Completed: false,
	}
	if _, err := dbdata.ORM.Insert(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}
	return task, nil
}

// Update 更新任务
func (s *TaskService) Update(id int64, title, dueDate string, amount float64) error {
	task := new(models.Task)
	found, err := dbdata.ORM.ID(id).Get(task)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("任务不存在")
	}
	if title != "" {
		task.Title = title
	}
	if dueDate != "" {
		task.DueDate = dueDate
	}
	task.Amount = amount
	_, err = dbdata.ORM.ID(id).Update(task)
	return err
}

// ToggleComplete 切换完成状态
func (s *TaskService) ToggleComplete(id int64) error {
	task := new(models.Task)
	found, err := dbdata.ORM.ID(id).Get(task)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("任务不存在")
	}
	task.Completed = !task.Completed
	_, err = dbdata.ORM.ID(id).Cols("completed", "updated_at").Update(task)
	return err
}

// Delete 删除任务
func (s *TaskService) Delete(id int64) error {
	_, err := dbdata.ORM.ID(id).Delete(new(models.Task))
	return err
}
