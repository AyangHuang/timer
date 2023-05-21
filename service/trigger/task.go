package trigger

import (
	"context"
	"time"
	"timer/common/conf"
	"timer/common/consts"
	"timer/common/model/po"
	"timer/common/model/vo"
	"timer/dao/task"
)

type TaskService struct {
	config *conf.SchedulerAppConfig
	cache  *task.TaskCache
	dao    taskDAO
}

func NewTaskService(dao *task.TaskDao, cache *task.TaskCache, confPrivder *conf.SchedulerAppConfig) *TaskService {
	return &TaskService{
		config: confPrivder,
		dao:    dao,
		cache:  cache,
	}
}

func (t *TaskService) GetTasksByTime(ctx context.Context, key string, bucket int, start, end time.Time) ([]*vo.Task, error) {
	// 先走 redis zset， timerID_runTime
	if tasks, err := t.cache.GetTasksByTime(ctx, key, start.UnixMilli(), end.UnixMilli()); err == nil && len(tasks) > 0 {
		return vo.NewTasks(tasks), nil
	}

	// 倘若缓存 miss 再走 db
	// 注意：db 是查 task 表；其实是 migrator 把 task 提取到 redis 中的
	tasks, err := t.dao.GetTasks(ctx, task.WithStartTime(start), task.WithEndTime(end), task.WithStatus(int32(consts.NotRunned.ToInt())))
	if err != nil {
		return nil, err
	}

	maxBucket := t.config.BucketsNum
	var validTask []*po.Task
	for _, task := range tasks {
		// 提取属于该桶的，因为 mysql 存的 task 并没有分桶，所有这里需要分桶
		if task.TimerID%uint(maxBucket) != uint(bucket) {
			continue
		}
		validTask = append(validTask, task)
	}

	return vo.NewTasks(validTask), nil
}

type taskDAO interface {
	GetTasks(ctx context.Context, opts ...task.Option) ([]*po.Task, error)
}
