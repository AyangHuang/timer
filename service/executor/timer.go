package executor

import (
	"context"
	"sync"
	"time"
	"timer/common/conf"
	"timer/common/consts"
	"timer/common/model/po"
	"timer/common/model/vo"
	"timer/dao/task"
	"timer/dao/timer"
	"timer/pkg/logger"
)

type TimerService struct {
	sync.Once
	config   *conf.MigratorAppConfig
	ctx      context.Context
	stop     func()
	timers   map[uint]*vo.Timer
	timerDAO timerDAO
	taskDAO  *task.TaskDao
}

func NewTimerService(timerDAO *timer.TimerDao, taskDAO *task.TaskDao, conf *conf.MigratorAppConfig) *TimerService {
	return &TimerService{
		config:   conf,
		timers:   make(map[uint]*vo.Timer),
		timerDAO: timerDAO,
		taskDAO:  taskDAO,
	}
}

// Start 二级迁移模块，2 分钟就把下一个 2 分钟需要用到全部的 tiimer 定义从数据库迁移到进程缓存中
// 注意：每个节点都需要
func (t *TimerService) Start(ctx context.Context) {
	t.Do(func() {
		go func() {
			t.ctx, t.stop = context.WithCancel(ctx)

			// 2 级迁移时间 2 分钟，即 2 分钟执行一次
			stepMinutes := t.config.TimerDetailCacheMinutes
			ticker := time.NewTicker(time.Duration(stepMinutes) * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				select {
				case <-t.ctx.Done():
					return
				default:
				}

				// 开启一个协程处理二级迁移
				go func() {
					start := time.Now()
					// 注意是重新覆盖，即每 2 分钟就会覆盖掉缓存的 timer。
					// 也就是 2 分钟后保证最终一致性
					t.timers, _ = t.getTimersByTime(ctx, start, start.Add(time.Duration(stepMinutes)*time.Minute))
				}()
			}
		}()
	})
}

func (t *TimerService) getTimersByTime(ctx context.Context, start, end time.Time) (map[uint]*vo.Timer, error) {
	// 从 mysql 获取 2 分钟后需要执行的 task 完整定义
	tasks, err := t.taskDAO.GetTasks(ctx, task.WithStartTime(start), task.WithEndTime(end))
	if err != nil {
		return nil, err
	}

	// 获取 task 中定时器 id
	timerIDs := getTimerIDs(tasks)
	if len(timerIDs) == 0 {
		return nil, nil
	}
	// 获取 2分钟后要执行的 timer
	pTimers, err := t.timerDAO.GetTimers(ctx, timer.WithIDs(timerIDs), timer.WithStatus(int32(consts.Enabled)))
	if err != nil {
		return nil, err
	}

	// 存入本地缓存map中，直接更新替换即可
	return getTimersMap(pTimers)
}

func getTimerIDs(tasks []*po.Task) []uint {
	timerIDSet := make(map[uint]struct{})
	// 使用 map 去重
	for _, task := range tasks {
		if _, ok := timerIDSet[task.TimerID]; ok {
			continue
		}
		timerIDSet[task.TimerID] = struct{}{}
	}
	timerIDs := make([]uint, 0, len(timerIDSet))
	// 再把 map 传入 slice 中
	for id := range timerIDSet {
		timerIDs = append(timerIDs, id)
	}
	return timerIDs
}

func getTimersMap(pTimers []*po.Timer) (map[uint]*vo.Timer, error) {
	vTimers, err := vo.NewTimers(pTimers)
	if err != nil {
		return nil, err
	}

	timers := make(map[uint]*vo.Timer, len(vTimers))
	for _, vTimer := range vTimers {
		timers[vTimer.ID] = vTimer
	}
	return timers, nil
}

func (t *TimerService) GetTimer(ctx context.Context, id uint) (*vo.Timer, error) {
	// 先查本地缓存（二级迁移模块干的事情）
	if vTimer, ok := t.timers[id]; ok {
		// log.InfoContextf(ctx, "get timer from local cache success, timer: %+v", vTimer)
		return vTimer, nil
	}

	logger.WarnContextf(ctx, "get timer from local cache failed, timerID: %d", id)

	// 再查 mysql
	timer, err := t.timerDAO.GetTimer(ctx, timer.WithID(id))
	if err != nil {
		return nil, err
	}

	return vo.NewTimer(timer)
}

func (t *TimerService) Stop() {
	t.stop()
}

var _ timerDAO = &timer.TimerDao{}

type timerDAO interface {
	GetTimer(context.Context, ...timer.Option) (*po.Timer, error)
	GetTimers(ctx context.Context, opts ...timer.Option) ([]*po.Timer, error)
}
