package scheduler

import (
	"context"
	"time"
	"timer/common/conf"
	"timer/common/utils"
	"timer/pkg/logger"
	"timer/pkg/redis"
	"timer/service/trigger"
)

type Worker struct {
	conf          *conf.SchedulerAppConfig
	trigger       *trigger.Worker
	lockService   lockService
	bucketGetter  bucketGetter
	minuteBuckets map[string]int
}

func NewWorker(trigger *trigger.Worker, redisClient *redis.Client, conf *conf.SchedulerAppConfig) *Worker {
	return &Worker{
		trigger:       trigger,
		lockService:   redisClient,
		bucketGetter:  redisClient,
		conf:          conf,
		minuteBuckets: make(map[string]int),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.trigger.Start(ctx)

	// 桶在时间维度是根据分钟来切割的
	// 100 毫秒执行一次
	ticker := time.NewTicker(time.Duration(w.conf.TryLockGapMilliSeconds) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			logger.WarnContext(ctx, "stopped")
			return nil
		default:
		}

		w.handleSlices(ctx)
	}
	return nil
}

func (w *Worker) handleSlices(ctx context.Context) {
	// 获取 1 分钟内再分桶的数量
	for i := 0; i < w.getValidBucket(ctx); i++ {
		// 逐个获取分布式锁
		w.handleSlice(ctx, i)
	}
}

func (w *Worker) getValidBucket(ctx context.Context) int {
	return w.conf.BucketsNum
}

func (w *Worker) handleSlice(ctx context.Context, bucketID int) {
	now := time.Now()
	// logger.InfoContextf(ctx, "scheduler_1 start: %v", time.Now())

	// 处理前一分钟的，因为有可能前一分钟失败了，分布式锁没有续期。也就是说允许拿到锁后崩掉一次
	go w.asyncHandleSlice(ctx, now.Add(-time.Minute), bucketID)

	// 处理当前分钟的
	go w.asyncHandleSlice(ctx, now, bucketID)

	// 提前获取下一个分钟的，这样可以减少分钟开头的延迟，但是会增加无效的空转时间
	// go w.asyncHandleSlice(ctx, now.Add(time.Minute), bucketID)
	// 另外的解决方法：把分钟的前几秒划分到前一个分钟里去，实现：在 tableName format 那里处理一下
	// 这里采用另外的解决方法（还没处理，后续 update）

	// logger.InfoContextf(ctx, "scheduler_1 end: %v", time.Now())
}

func (w *Worker) asyncHandleSlice(ctx context.Context, t time.Time, bucketID int) {
	// logger.InfoContextf(ctx, "scheduler_2 start: %v", time.Now())
	// defer func() {
	// 	logger.InfoContextf(ctx, "scheduler_2 end: %v", time.Now())
	// }()

	locker := w.lockService.GetDistributionLock(utils.GetTimeBucketLockKey(t, bucketID))

	// 获取该分片的分布式锁
	// 获取到分布式锁并设置分布式锁过期时间 70 s。（成功的话会再续期增加 130 s）
	if err := locker.Lock(ctx, int64(w.conf.TryLockSeconds)); err != nil {
		// log.WarnContextf(ctx, "get lock failed, err: %v, key: %s", err, utils.GetTimeBucketLockKey(t, bucketID))
		return
	}

	logger.InfoContextf(ctx, "get scheduler lock success, key: %s", utils.GetTimeBucketLockKey(t, bucketID))

	ack := func() {
		// 时间片执行成功后，更新的分布式锁时间为 130 s
		if err := locker.ExpireLock(ctx, int64(w.conf.SuccessExpireSeconds)); err != nil {
			logger.ErrorContextf(ctx, "expire lock failed, lock key: %s, err: %v", utils.GetTimeBucketLockKey(t, bucketID), err)
		}
	}

	// 处理该分片（也就是该分钟的某一个桶的全部任务）
	if err := w.trigger.Work(ctx, utils.GetSliceMsgKey(t, bucketID), ack); err != nil {
		logger.ErrorContextf(ctx, "trigger work failed, err: %v", err)
	}
}

type lockService interface {
	GetDistributionLock(key string) redis.DistributeLocker
}

type bucketGetter interface {
	Get(ctx context.Context, key string) (string, error)
}
