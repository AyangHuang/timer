package executor

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"
	"time"
	"timer/common/consts"
	"timer/common/model/vo"
	"timer/common/utils"
	"timer/dao/task"
	"timer/pkg/bloom"
	"timer/pkg/logger"
	"timer/pkg/xhttp"
)

type Worker struct {
	timerService *TimerService
	taskDAO      *task.TaskDao
	httpClient   *xhttp.JSONClient
	bloomFilter  *bloom.Filter
}

func NewWorker(timerService *TimerService, taskDAO *task.TaskDao, httpClient *xhttp.JSONClient, bloomFilter *bloom.Filter) *Worker {
	return &Worker{
		timerService: timerService,
		taskDAO:      taskDAO,
		httpClient:   httpClient,
		bloomFilter:  bloomFilter,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.timerService.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, timerIDUnixKey string) error {
	// log.InfoContextf(ctx, "executor_1 start: %v", time.Now())
	// defer func() {
	// 	log.InfoContextf(ctx, "executor_1 end: %v", time.Now())
	// }()
	// 拿到消息，查询一次完整的 timer 定义
	timerID, unix, err := utils.SplitTimerIDUnix(timerIDUnixKey)
	if err != nil {
		return err
	}

	// bloomFilter 布隆过滤器查看是否存在，已经存在则无法判断，没有存在则一定没有执行过
	// bloomFilter key 为某天，即每一天就会重新创建一个 bloomFilter，防止数据太多，判断率下降
	if exist, err := w.bloomFilter.Exist(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timerIDUnixKey); err != nil || exist {
		logger.WarnContextf(ctx, "bloom filter check failed, start to check db, bloom key: %s, timerIDUnixKey: %s, err: %v, exist: %t", utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timerIDUnixKey, err, exist)
		// 查 mysql 判断 task 被执行过没有
		// 这里不需要事务+独占锁，因为不会产生并发的情况，一个 task 只能在该 1 分钟后分布式锁过期，才可能被执行
		task, err := w.taskDAO.GetTask(ctx, task.WithTimerID(timerID), task.WithRunTimer(time.UnixMilli(unix)))
		if err == nil && task.Status != consts.NotRunned.ToInt() {
			// 重复执行的任务
			logger.WarnContextf(ctx, "task is already executed, timerID: %d, exec_time: %v", timerID, task.RunTimer)
			return nil
		}
	}

	// 执行定时任务
	return w.executeAndPostProcess(ctx, timerID, unix)
}

func (w *Worker) executeAndPostProcess(ctx context.Context, timerID uint, unix int64) error {
	// 查询 timer 完整的定义，执行回调
	// 1. 先查本地进程的缓存，看有没有
	// 2. 再查 mysql 的 timer
	timer, err := w.timerService.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("get timer failed, id: %d, err: %w", timerID, err)
	}

	// 定时器已经处于去激活态，则无需处理任务
	if timer.Status != consts.Enabled {
		logger.WarnContextf(ctx, "timer has alread been unabled, timerID: %d", timerID)
		return nil
	}

	execTime := time.Now()
	// 执行 task 的 http 回调请求
	resp, err := w.execute(ctx, timer)
	// log.InfoContextf(ctx, "execute timer: %d, resp: %v, err: %v", timerID, resp, err)
	// 加入布隆过滤器和更新 task 状态
	return w.postProcess(ctx, resp, err, timer.App, timerID, unix, execTime)
}

func (w *Worker) execute(ctx context.Context, timer *vo.Timer) (map[string]interface{}, error) {
	var (
		resp map[string]interface{}
		err  error
	)
	switch strings.ToUpper(timer.NotifyHTTPParam.Method) {
	case nethttp.MethodGet:
		err = w.httpClient.Get(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, nil, &resp)
	case nethttp.MethodPatch:
		err = w.httpClient.Patch(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodDelete:
		err = w.httpClient.Delete(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodPost:
		err = w.httpClient.Post(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	default:
		err = fmt.Errorf("invalid http method: %s, timer: %s", timer.NotifyHTTPParam.Method, timer.Name)
	}

	return resp, err
}

func (w *Worker) postProcess(ctx context.Context, resp map[string]interface{}, execErr error, app string, timerID uint, unix int64, execTime time.Time) error {
	// 布隆过滤器设置已经执行
	if err := w.bloomFilter.Set(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), utils.UnionTimerIDUnix(timerID, unix), consts.BloomFilterKeyExpireSeconds); err != nil {
		logger.ErrorContextf(ctx, "set bloom filter failed, key: %s, err: %v", utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), err)
	}

	// 查询 mysql 的整个 task
	task, err := w.taskDAO.GetTask(ctx, task.WithTimerID(timerID), task.WithRunTimer(time.UnixMilli(unix)))
	if err != nil {
		return fmt.Errorf("get task failed, timerID: %d, runTimer: %d, err: %w", timerID, time.UnixMilli(unix), err)
	}

	respBody, _ := json.Marshal(resp)
	task.Output = string(respBody)

	if execErr != nil {
		task.Status = consts.Failed.ToInt()
	} else {
		task.Status = consts.Successed.ToInt()
	}

	// update task 数据库的状态
	return w.taskDAO.UpdateTask(ctx, task)
}
