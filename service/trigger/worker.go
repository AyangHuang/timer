package trigger

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"timer/common/conf"
	"timer/common/model/vo"
	"timer/common/utils"
	"timer/pkg/concurrency"
	"timer/pkg/logger"
	"timer/pkg/pool"
	"timer/pkg/redis"
	"timer/service/executor"
)

type Worker struct {
	task        taskService
	config      *conf.TriggerAppConfig
	pool        pool.WorkerPool
	executor    *executor.Worker
	lockService *redis.Client
}

func NewWorker(executor *executor.Worker, task *TaskService, lockService *redis.Client, conf *conf.TriggerAppConfig) *Worker {
	return &Worker{
		executor:    executor,
		task:        task,
		lockService: lockService,
		pool:        pool.NewGoWorkerPool(conf.WorkersNum),
		config:      conf,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.executor.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, minuteBucketKey string, ack func()) error {
	// log.InfoContextf(ctx, "trigger_1 start: %v", time.Now())
	// defer func() {
	// 	log.InfoContextf(ctx, "trigger_1 end: %v", time.Now())
	// }()

	// 进行为时一分钟的 zrange 处理
	// startTime 是该时间片的开始时间
	startTime, err := getStartMinute(minuteBucketKey)
	if err != nil {
		return err
	}

	// 每一秒轮询一次（轮询该分片的任务）
	ticker := time.NewTicker(time.Duration(w.config.ZRangeGapSeconds) * time.Second)
	defer ticker.Stop()

	endTime := startTime.Add(time.Minute)

	// chan size = 1 分钟除 1 秒 +1 = 61
	// 用来保存 err，有任一一个 1 秒内的任务执行出错，结束该分钟的处理
	notifier := concurrency.NewSafeChan(int(time.Minute/(time.Duration(w.config.ZRangeGapSeconds)*time.Second)) + 1)
	defer notifier.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// 异步协程轮询该时间片的第一秒，因为 Ticker 是一秒后才会有，所有也就是说丧失了第一秒的时间
	go func() {
		// log.InfoContextf(ctx, "trigger_2 start: %v", time.Now())
		// defer func() {
		// 	log.InfoContextf(ctx, "trigger_2 end: %v", time.Now())
		// }()
		defer wg.Done()
		// 该一分钟时间片的第一秒
		if err := w.handleBatch(ctx, minuteBucketKey, startTime, startTime.Add(time.Duration(w.config.ZRangeGapSeconds)*time.Second)); err != nil {
			notifier.Put(err)
		}
	}()

	for range ticker.C {
		select {
		case e := <-notifier.GetChan():
			err, _ = e.(error)
			return err
		default:
		}

		// startTime 增加 1 秒，超过 end time 就停止
		if startTime = startTime.Add(time.Duration(w.config.ZRangeGapSeconds) * time.Second); startTime.Equal(endTime) || startTime.After(endTime) {
			break
		}

		// log.InfoContextf(ctx, "start time: %v", startTime)

		wg.Add(1)
		go func(startTime time.Time) {
			// log.InfoContextf(ctx, "trigger_2 start: %v", time.Now())
			// defer func() {
			// 	log.InfoContextf(ctx, "trigger_2 end: %v", time.Now())
			// }()
			defer wg.Done()
			if err := w.handleBatch(ctx, minuteBucketKey, startTime, startTime.Add(time.Duration(w.config.ZRangeGapSeconds)*time.Second)); err != nil {
				notifier.Put(err)
			}
		}(startTime)
	}

	// 等待全部时间片处理完
	wg.Wait()
	select {
	case e := <-notifier.GetChan():
		err, _ = e.(error)
		return err
	default:
	}

	// 增加分布式锁的过期时间，那么就代表不会有人能抢到这个锁，证明完成成功
	ack()
	logger.InfoContextf(ctx, "ack success, key: %s", minuteBucketKey)
	return nil
}

// handleBatch 处理时间片内（start，end）秒的 task
func (w *Worker) handleBatch(ctx context.Context, key string, start, end time.Time) error {
	// start，end 1 秒中内
	bucket, err := getBucket(key)
	if err != nil {
		return err
	}

	// key 是时间片分桶的整个 key，也就是该分片的 key
	// 利用 zrange 获取该时间片里（1分钟）为该 秒 执行的 task 的 timerID，taskRunTime
	tasks, err := w.task.GetTasksByTime(ctx, key, bucket, start, end)
	if err != nil {
		return err
	}

	timerIDs := make([]uint, 0, len(tasks))
	for _, task := range tasks {
		timerIDs = append(timerIDs, task.TimerID)
	}
	// log.InfoContextf(ctx, "key: %s, get tasks: %+v, start: %v, end: %v", key, timerIDs, start, end)
	for _, task := range tasks {
		task := task
		// 对于该秒内的每一个任务对应一个 G 执行
		if err := w.pool.Submit(func() {
			// log.InfoContextf(ctx, "trigger_3 start: %v", time.Now())
			// defer func() {
			// 	log.InfoContextf(ctx, "trigger_3 end: %v", time.Now())
			// }()

			if err := w.executor.Work(ctx, utils.UnionTimerIDUnix(task.TimerID, task.RunTimer.UnixMilli())); err != nil {
				logger.ErrorContextf(ctx, "executor work failed, err: %v", err)
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func getStartMinute(slice string) (time.Time, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return time.Time{}, fmt.Errorf("invalid format of msg key: %s", slice)
	}

	return utils.GetStartMinute(timeBucket[0])
}

func getBucket(slice string) (int, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return -1, fmt.Errorf("invalid format of msg key: %s", slice)
	}
	return strconv.Atoi(timeBucket[1])
}

type taskService interface {
	GetTasksByTime(ctx context.Context, key string, bucket int, start, end time.Time) ([]*vo.Task, error)
}
