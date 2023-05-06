package consts

type TimerStatus int

func (t TimerStatus) ToInt() int {
	return int(t)
}

const (
	Unabled TimerStatus = 0
	Enabled TimerStatus = 1
)
