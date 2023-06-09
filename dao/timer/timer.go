package timer

import (
	"context"
	"gorm.io/gorm"
	"timer/common/model/po"
	"timer/pkg/logger"
)

type TimerDao struct {
	db *gorm.DB
}

func NewTimerDao(db *gorm.DB) *TimerDao {
	return &TimerDao{
		db: db,
	}
}

func (dao *TimerDao) TableWithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx).Table(po.TimerTable)
}

func (dao *TimerDao) Table() *gorm.DB {
	return dao.db.Table(po.TimerTable)
}

func (dao *TimerDao) CreateTimer(ctx context.Context, timer *po.Timer) (uint, error) {
	return timer.ID, dao.TableWithContext(ctx).Create(timer).Error
}

func (dao *TimerDao) DeleteTimer(ctx context.Context, id uint) error {
	return dao.TableWithContext(ctx).Delete(&po.Timer{}, id).Error
}

func (dao *TimerDao) GetTimerByID(ctx context.Context, timer *po.Timer) error {
	return dao.TableWithContext(ctx).First(timer).Error
}

func (dao *TimerDao) DoWithTransactionAndLock(ctx context.Context, id uint, do func(context.Context, *TimerDao, *po.Timer) error) error {
	// 数据库事务
	return dao.db.Transaction(func(tx *gorm.DB) error {
		defer func() {
			if err := recover(); err != nil {
				// 事务回滚
				tx.Rollback()
				logger.ErrorContextf(ctx, "transaction with lock err: %v, timer id: %d", err, id)
			}
		}()

		var timer po.Timer

		// 设置该事务为锁读
		if err := tx.Set("gorm:query_option", "FOR UPDATE").WithContext(ctx).Table(po.TimerTable).First(&timer, id).Error; err != nil {
			return err
		}

		return do(ctx, NewTimerDao(tx), &timer)
	})
}

func (dao *TimerDao) UpdateTimerStatus(ctx context.Context, id uint, timerStatus int) error {
	return dao.TableWithContext(ctx).Where("id=?", id).Update("status", timerStatus).Error
}

func (dao *TimerDao) GetTimer(ctx context.Context, opts ...Option) (*po.Timer, error) {
	db := dao.TableWithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}
	var timer po.Timer
	return &timer, db.First(&timer).Error
}

func (dao *TimerDao) GetTimers(ctx context.Context, opts ...Option) ([]*po.Timer, error) {
	db := dao.TableWithContext(ctx)
	for _, opt := range opts {
		db = opt(db)
	}
	var timers []*po.Timer
	return timers, db.Scan(&timers).Error
}

func (dao *TimerDao) BatchCreateRecords(ctx context.Context, tasks []*po.Task) error {
	return dao.TableWithContext(ctx).CreateInBatches(tasks, len(tasks)).Error
}
