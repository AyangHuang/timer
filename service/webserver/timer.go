package webserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"timer/common/conf"
	"timer/common/consts"
	"timer/common/model/po"
	"timer/common/model/vo"
	timerUtil "timer/common/utils/timer"
	"timer/dao/task"
	timerD "timer/dao/timer"
	"timer/pkg/cron"
)

type TimerServer struct {
	timerDao      timerDao
	taskDao       taskDao
	taskCache     taskCache
	cronParser    cronParser
	migrateConfig *conf.MigratorAppConfig
}

func NewTimerServer(timer *timerD.TimerDao, task *task.TaskDao, taskCache *task.TaskCache, parser *cron.Parser, config *conf.MigratorAppConfig) *TimerServer {
	return &TimerServer{
		timerDao:      timer,
		taskDao:       task,
		cronParser:    parser,
		migrateConfig: config,
		taskCache:     taskCache,
	}
}

func (server *TimerServer) CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error) {
	// 判断 cron 表达式是否有效
	if !server.cronParser.IsValidCronExpr(timer.Cron) {
		return 0, vo.ErrCronExprUnValid
	}

	// 转换成数据库映射 struct
	poTimer, err := timer.ToPo()
	if err != nil {
		return 0, err
	}

	return server.timerDao.CreateTimer(ctx, poTimer)
}

func (server *TimerServer) DeleteTimer(ctx context.Context, id uint) error {
	return server.timerDao.DeleteTimer(ctx, id)
}

func (server *TimerServer) EnableTimer(ctx context.Context, id uint) error {
	timer := &po.Timer{}
	timer.ID = id

	// 整个 MySQL 操作是事务+独占锁
	do := func(ctx context.Context, dao *timerD.TimerDao, timer *po.Timer) error {
		var err error

		// 获取 timer 完整定义
		err = dao.GetTimerByID(ctx, timer)
		if err != nil {
			fmt.Print(err)
			return err
		}
		// 校验是否处于非激活状态
		if timer.Status != consts.Unabled.ToInt() {
			return fmt.Errorf("not unabled status, enable failed, timer id: %d", timer.ID)
		}

		start := time.Now()
		// 获取两倍一级迁移的时间
		end := timerUtil.GetForwardTwoMigrateStepEnd(start, time.Duration(server.migrateConfig.MigrateStepMinutes)*time.Minute)

		// 获取定时器在两倍一级迁移时间的定时任务时间点
		executeTimers, err := server.cronParser.NextsBefore(timer.Cron, end)
		if err != nil {
			return err
		}

		// 根据执行时间批量生成定时任务
		tasks := timer.BatchTasksFromTimer(executeTimers)

		// task 插入 mysql 中
		err = server.taskDao.BatchCreateTasks(tasks)
		if err != nil {
			return err
		}

		// 加入 redis zset 中
		// score(runtime) member(timerID_runtime)
		err = server.taskCache.BatchCreateTasks(ctx, tasks)
		if err != nil {
			return err
		}

		// 修改数据库中 timer 状态为激活态
		return server.timerDao.UpdateTimerStatus(ctx, timer.ID, int(consts.Enabled))
	}

	return server.timerDao.DoWithTransactionAndLock(ctx, id, do)
}

func (server *TimerServer) UnableTimer(ctx *gin.Context, id uint) error {
	// 其实应该先检验是否处于非激活状态，这里简单就不检验了
	return server.timerDao.UpdateTimerStatus(ctx, id, consts.Unabled.ToInt())
}

var _ timerDao = &timerD.TimerDao{}
var _ cronParser = &cron.Parser{}

type timerDao interface {
	CreateTimer(context.Context, *po.Timer) (uint, error)
	DeleteTimer(ctx context.Context, in uint) error
	GetTimerByID(context.Context, *po.Timer) error
	DoWithTransactionAndLock(ctx context.Context, uid uint, do func(context.Context, *timerD.TimerDao, *po.Timer) error) error
	UpdateTimerStatus(ctx context.Context, id uint, timerStatus int) error
}

type taskDao interface {
	BatchCreateTasks(task []*po.Task) error
}

type taskCache interface {
	BatchCreateTasks(ctx context.Context, tasks []*po.Task) error
}

type cronParser interface {
	IsValidCronExpr(string) bool
	NextsBefore(cron string, end time.Time) ([]time.Time, error)
}
