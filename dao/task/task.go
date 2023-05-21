package task

import (
	"context"
	"gorm.io/gorm"
	"timer/common/model/po"
)

type TaskDao struct {
	db *gorm.DB
}

func NewTaskDao(db *gorm.DB) *TaskDao {
	return &TaskDao{
		db: db,
	}
}

func (dao *TaskDao) TableWithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx).Table(po.TaskTable)
}

func (dao *TaskDao) Table() *gorm.DB {
	return dao.db.Table(po.TaskTable)
}

func (dao *TaskDao) BatchCreateTasks(task []*po.Task) error {
	return dao.Table().Create(&task).Error
}

func (dao *TaskDao) GetTask(ctx context.Context, opts ...Option) (*po.Task, error) {
	db := dao.TableWithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}

	var task po.Task
	return &task, db.First(&task).Error
}

// GetTasks 根据 option 获取 tasks
func (dao *TaskDao) GetTasks(ctx context.Context, opts ...Option) ([]*po.Task, error) {
	db := dao.TableWithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}

	var tasks []*po.Task
	return tasks, db.Scan(&tasks).Error
}

func (dao *TaskDao) UpdateTask(ctx context.Context, task *po.Task) error {
	return dao.TableWithContext(ctx).Updates(task).Error
}
