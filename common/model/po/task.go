package po

import (
	"gorm.io/gorm"
	"time"
)

// Task 运行流水记录
type Task struct {
	gorm.Model
	// 整个 task 的 id = timerID_runTimer
	App      string    `gorm:"column:app;NOT NULL"`           // 定义ID
	TimerID  uint      `gorm:"column:timer_id;NOT NULL"`      // 所属 timer 的 ID
	Output   string    `gorm:"column:output;default:null"`    // 执行结果
	RunTimer time.Time `gorm:"column:run_timer;default:null"` // 执行时间
	CostTime int       `gorm:"column:cost_time"`              // 执行耗时
	Status   int       `gorm:"column:status;NOT NULL"`        // 当前状态
}

func (t *Task) TableName() string {
	return "task"
}

type MinuteTaskCnt struct {
	Minute string `gorm:"column:minute"`
	Cnt    int64  `gorm:"column:cnt"`
}
