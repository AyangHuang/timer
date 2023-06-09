package consts

const (
	MinuteFormat = "2006-01-02 15:04"
	HourFormat   = "2006-01-02 15"
	DayFormat    = "2006-01-02"
	// BloomFilterKeyExpireSeconds 一天过期
	BloomFilterKeyExpireSeconds = 24 * 60 * 60
)

type TimerStatus int
type TaskStatus int

func (t TimerStatus) ToInt() int {
	return int(t)
}

func (t TaskStatus) ToInt() int {
	return int(t)
}

const (
	Unabled TimerStatus = 0
	Enabled TimerStatus = 1

	NotRunned TaskStatus = 0
	Running   TaskStatus = 1
	Successed TaskStatus = 2
	Failed    TaskStatus = 3
)
