package timer

import (
	"fmt"
	"time"
)

func GetForwardTwoMigrateStepEnd(cur time.Time, diff time.Duration) time.Time {
	end := cur.Add(diff)
	return time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), 0, 0, 0, time.Local)
}

func UnionTimerIDUnix(timerID uint, unix int64) string {
	return fmt.Sprintf("%d_%d", timerID, unix)
}
