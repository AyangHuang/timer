package task

import (
	"context"
	"fmt"
	"time"
	"timer/common/conf"
	"timer/common/consts"
	"timer/common/model/po"
	"timer/common/utils"
	"timer/pkg/redis"
)

type TaskCache struct {
	rdb  cacheClient
	conf *conf.SchedulerAppConfig
}

func NewTaskCache(rdb *redis.Client, conf *conf.SchedulerAppConfig) *TaskCache {
	return &TaskCache{
		rdb:  rdb,
		conf: conf,
	}
}

func (tc *TaskCache) BatchCreateTasks(ctx context.Context, tasks []*po.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	commands := make([]*redis.Command, 0, 2*len(tasks))
	for _, task := range tasks {
		unix := task.RunTimer.UnixMilli()
		// 获取桶的 value，根据两个维度，时间，同一个时间再分桶
		tableName := tc.GetTableName(task)

		// zadd key score member
		// zadd key(minute_bucket)  score(runTime) member(timerID_runTime)
		commands = append(commands, redis.NewZAddCommand(tableName, unix, utils.UnionTimerIDUnix(task.TimerID, unix)))
		// zset 一天后过期（其实可以再缩短很多，例如：4 小时过期）
		aliveSeconds := int64(time.Until(task.RunTimer.Add(24*time.Hour)) / time.Second)
		commands = append(commands, redis.NewExpireCommand(tableName, aliveSeconds))
	}

	_, err := tc.rdb.Transaction(ctx, commands...)
	return err
}

func (t *TaskCache) GetTableName(task *po.Task) string {
	maxBucket := t.conf.BucketsNum

	// 二位分片，根据一分钟，一分钟里再分桶
	// %s:%d 某一个分钟_哪个桶
	return fmt.Sprintf("%s:%d", task.RunTimer.Format(consts.MinuteFormat), int64(task.TimerID)%int64(maxBucket))
}

func (t *TaskCache) GetTasksByTime(ctx context.Context, table string, start, end int64) ([]*po.Task, error) {
	// zrangebyscore 获取 score 指定范围内内 value
	timerIDUnixs, err := t.rdb.ZrangeByScore(ctx, table, start, end-1)
	if err != nil {
		return nil, err
	}

	tasks := make([]*po.Task, 0, len(timerIDUnixs))
	for _, timerIDUnix := range timerIDUnixs {
		timerID, unix, _ := utils.SplitTimerIDUnix(timerIDUnix)
		tasks = append(tasks, &po.Task{
			TimerID:  timerID,
			RunTimer: time.UnixMilli(unix),
		})
	}

	return tasks, nil
}

var _ cacheClient = &redis.Client{}

type cacheClient interface {
	Transaction(ctx context.Context, commands ...*redis.Command) ([]interface{}, error)
	ZrangeByScore(ctx context.Context, table string, score1, score2 int64) ([]string, error)
	Expire(ctx context.Context, key string, expireSeconds int64) error
}
