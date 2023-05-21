package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"timer/common/consts"
)

func GetForwardTwoMigrateStepEnd(cur time.Time, diff time.Duration) time.Time {
	end := cur.Add(diff)
	return time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), 0, 0, 0, time.Local)
}

func UnionTimerIDUnix(timerID uint, unix int64) string {
	return fmt.Sprintf("%d_%d", timerID, unix)
}

func GetTimeBucketLockKey(t time.Time, bucketID int) string {
	return fmt.Sprintf("time_bucket_lock_%s_%d", t.Format(consts.MinuteFormat), bucketID)
}

func GetSliceMsgKey(t time.Time, bucketID int) string {
	return fmt.Sprintf("%s_%d", t.Format(consts.MinuteFormat), bucketID)
}

func GetStartMinute(timeStr string) (time.Time, error) {
	return time.ParseInLocation(consts.MinuteFormat, timeStr, time.Local)
}

func SplitTimerIDUnix(str string) (uint, int64, error) {
	timerIDUnix := strings.Split(str, "_")
	if len(timerIDUnix) != 2 {
		return 0, 0, fmt.Errorf("invalid timerID unix str: %s", str)
	}

	timerID, _ := strconv.ParseInt(timerIDUnix[0], 10, 64)
	unix, _ := strconv.ParseInt(timerIDUnix[1], 10, 64)
	return uint(timerID), unix, nil
}

func GetTaskBloomFilterKey(timeStr string) string {
	return "task_bloom_" + timeStr
}

func GetDayStr(t time.Time) string {
	return t.Format(consts.DayFormat)
}

func GetMigratorLockKey(t time.Time) string {
	return fmt.Sprintf("migrator_lock_%s", t.Format(consts.HourFormat))
}

func GetStartHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.Local)
}
