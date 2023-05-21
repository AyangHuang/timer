package vo

import (
	"encoding/json"
	"errors"
	"timer/common/consts"
	"timer/common/model/po"
)

type CreateTimerReq struct {
	Timer
}

type CreateTimerRespData struct {
	Id      uint `json:"id"`
	Success bool `json:"success"`
}

type TimerReq struct {
	App string `form:"app" json:"app" binding:"required"`
	ID  uint   `form:"id" json:"id" binding:"required"`
}

type Timer struct {
	ID              uint               `json:"id,omitempty"`
	App             string             `json:"app,omitempty" binding:"required"`             // 所属应用的名称
	Name            string             `json:"name,omitempty" binding:"required"`            // 定时器定义名称
	Status          consts.TimerStatus `json:"status"`                                       // 定时器定义状态，1:未激活, 2:已激活
	Cron            string             `json:"cron,omitempty" binding:"required"`            // 定时器定时配置
	NotifyHTTPParam *NotifyHTTPParam   `json:"notifyHTTPParam,omitempty" binding:"required"` // http 回调参数
}

type NotifyHTTPParam struct {
	Method string            `json:"method,omitempty" binding:"required"` // POST,GET 方法
	URL    string            `json:"url,omitempty" binding:"required"`    // URL 路径
	Header map[string]string `json:"header,omitempty"`                    // header 请求头
	Body   string            `json:"body,omitempty"`                      // 请求参数体
}

func (timer *Timer) Check() error {
	if timer.NotifyHTTPParam == nil {
		return errors.New("empty notify http params")
	}
	return nil
}

func (timer *Timer) ToPo() (*po.Timer, error) {
	var err error

	if err = timer.Check(); err != nil {
		return nil, err
	}

	param, err := json.Marshal(timer.NotifyHTTPParam)
	if err != nil {
		return nil, err
	}

	poTimer := &po.Timer{
		App:             timer.App,
		Name:            timer.Name,
		Status:          timer.Status.ToInt(),
		Cron:            timer.Cron,
		NotifyHTTPParam: string(param),
	}

	return poTimer, nil
}

func NewTimer(timer *po.Timer) (*Timer, error) {
	var param NotifyHTTPParam
	if err := json.Unmarshal([]byte(timer.NotifyHTTPParam), &param); err != nil {
		return nil, err
	}

	return &Timer{
		ID:              timer.ID,
		App:             timer.App,
		Name:            timer.Name,
		Status:          consts.TimerStatus(timer.Status),
		Cron:            timer.Cron,
		NotifyHTTPParam: &param,
	}, nil
}

func NewTimers(timers []*po.Timer) ([]*Timer, error) {
	vTimers := make([]*Timer, 0, len(timers))
	for _, timer := range timers {
		vTimer, err := NewTimer(timer)
		if err != nil {
			return nil, err
		}
		vTimers = append(vTimers, vTimer)
	}
	return vTimers, nil
}
