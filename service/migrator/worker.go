package service

import (
	"context"
	"time"
	"timer/common/conf"
	"timer/common/consts"
	"timer/common/utils"
	"timer/dao/task"
	"timer/dao/timer"
	"timer/pkg/cron"
	"timer/pkg/logger"
	"timer/pkg/pool"
	"timer/pkg/redis"
)

type Worker struct {
	timerDAO    *timer.TimerDao
	taskDAO     *task.TaskDao
	taskCache   *task.TaskCache
	cronParser  *cron.Parser
	lockService *redis.Client
	appConfig   *conf.MigratorAppConfig
	pool        pool.WorkerPool
}

func NewWorker(timerDAO *timer.TimerDao, taskDAO *task.TaskDao, taskCache *task.TaskCache, lockService *redis.Client,
	cronParser *cron.Parser, appConfig *conf.MigratorAppConfig) *Worker {
	return &Worker{
		pool:        pool.NewGoWorkerPool(appConfig.WorkersNum),
		timerDAO:    timerDAO,
		taskDAO:     taskDAO,
		taskCache:   taskCache,
		lockService: lockService,
		cronParser:  cronParser,
		appConfig:   appConfig,
	}
}

// Start 一级迁移模块
// 负责扫描全部的 timer 打点生成 task，存入数据库，存入 redis.zset
func (w *Worker) Start(ctx context.Context) error {
	// 一级迁移时间 60 分钟，
	// 但我觉得一级迁移时间 60 分钟，获取分布式锁的轮询时间应该是 1 分钟
	// 这样有节点工作一半宕机了，分布式锁没延期，其他节点才能迅速获取到
	ticker := time.NewTicker(time.Duration(w.appConfig.MigrateStepMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		logger.InfoContext(ctx, "migrator ticking...")
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// 获取 migrator 的每一个小时的分布式锁，获取得到则表示由该节点来处理改小时的模块。
		locker := w.lockService.GetDistributionLock(utils.GetMigratorLockKey(utils.GetStartHour(time.Now())))
		// 初设过期时间 20 min
		if err := locker.Lock(ctx, int64(w.appConfig.MigrateTryLockMinutes)*int64(time.Minute/time.Second)); err != nil {
			logger.ErrorContext(ctx, "migrator get lock failed, key: %s, err: %v", utils.GetMigratorLockKey(utils.GetStartHour(time.Now())), err)
			continue
		}

		// 获取分布式锁，代表获取该时间段的处理权
		if err := w.migrate(ctx); err != nil {
			logger.ErrorContext(ctx, "migrate failed, err: %v", err)
			continue
		}

		// 迁移成功过期时间设置为 120 分钟
		_ = locker.ExpireLock(ctx, int64(w.appConfig.MigrateSuccessExpireMinutes)*int64(time.Minute/time.Second))
	}
	return nil
}

func (w *Worker) migrate(ctx context.Context) error {
	// 返回所有的 定时器，牛逼，这么多啊
	// 其实内存还算很小，就算 1 w 个定时器，每个定时器 0.1KB（实际小很多），也就 1 mb
	timers, err := w.timerDAO.GetTimers(ctx, timer.WithStatus(int32(consts.Enabled.ToInt())))
	if err != nil {
		return err
	}

	now := time.Now()
	// 步长 60，即下一个 60 的时间
	start, end := utils.GetStartHour(now.Add(time.Duration(w.appConfig.MigrateStepMinutes)*time.Minute)), utils.GetStartHour(now.Add(2*time.Duration(w.appConfig.MigrateStepMinutes)*time.Minute))
	for _, timer := range timers {
		// 获取下一个时间段执行的 task
		nexts, _ := w.cronParser.NextsBetween(timer.Cron, start, end)
		// 根据时间创建 task，加入 mysql 中
		if err := w.timerDAO.BatchCreateRecords(ctx, timer.BatchTasksFromTimer(nexts)); err != nil {
			logger.ErrorContextf(ctx, "migrator batch create records for timer: %d failed, err: %v", timer.ID, err)
		}
	}

	// log.InfoContext(ctx, "migrator batch create db tasks susccess")
	// 迁移完成后，将所有添加的 task 取出，添加到 redis 当中
	return w.migrateToCache(ctx, start, end)
}

func (w *Worker) migrateToCache(ctx context.Context, start, end time.Time) error {
	// 迁移完成后，将所有添加的 task 取出，添加到 redis 当中
	tasks, err := w.taskDAO.GetTasks(ctx, task.WithStartTime(start), task.WithEndTime(end))
	if err != nil {
		logger.ErrorContextf(ctx, "migrator batch get tasks failed, err: %v", err)
		return err
	}
	// log.InfoContext(ctx, "migrator batch get tasks susccess")
	return w.taskCache.BatchCreateTasks(ctx, tasks)
}
