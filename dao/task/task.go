package task

import (
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

func (dao *TaskDao) BatchCreateTasks(task []*po.Task) error {
	return dao.db.Create(&task).Error
}
