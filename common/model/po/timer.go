package po

import (
	"gorm.io/gorm"
)

const TimerTable = "timer"

// Timer 定时器定义
type Timer struct {
	gorm.Model
	App             string `gorm:"column:app;NOT NULL" json:"app,omitempty"`                             // 定时器定义名称
	Name            string `gorm:"column:name;NOT NULL" json:"name,omitempty"`                           // 定时器定义名称
	Status          int    `gorm:"column:status;NOT NULL" json:"status,omitempty"`                       // 定时器定义状态，1:未激活, 2:已激活
	Cron            string `gorm:"column:cron;NOT NULL" json:"cron,omitempty"`                           // 定时器定时配置
	NotifyHTTPParam string `gorm:"column:notify_http_param;NOT NULL" json:"notify_http_param,omitempty"` // Http 回调参数
}
