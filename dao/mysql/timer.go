package mysql

import (
	"context"
	"gorm.io/gorm"
	"timer/common/model/po"
)

type TimerDao struct {
	db *gorm.DB
}

func NewTimerDao(db *gorm.DB) *TimerDao {
	return &TimerDao{
		db: db,
	}
}

func (dao *TimerDao) CreateTimer(ctx context.Context, timer *po.Timer) (uint, error) {
	return timer.ID, dao.db.WithContext(ctx).Table(po.TimerTable).Create(timer).Error
}
