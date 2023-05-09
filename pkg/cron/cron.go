package cron

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type Parser struct {
}

func NewCronParser() *Parser {
	return &Parser{}
}

// IsValidCronExpr 判断是否符合 cron 表达式
func (c *Parser) IsValidCronExpr(cron string) bool {
	_, err := cronexpr.Parse(cron)
	return err == nil
}

func (c *Parser) NextsBefore(cron string, end time.Time) ([]time.Time, error) {
	return c.NextsBetween(cron, time.Now(), end)
}

func (c *Parser) NextsBetween(cron string, start, end time.Time) ([]time.Time, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("end can not earlier than start, start: %v, end: %v", start, end)
	}

	expr, err := cronexpr.Parse(cron)
	if err != nil {
		return nil, err
	}

	var nexts []time.Time
	for start.Before(end) {
		next := expr.Next(start)
		if next.UnixNano() < 0 {
			return nil, fmt.Errorf("fail to parse time from cron: %s", cron)
		}
		nexts = append(nexts, next)
		start = next
	}

	return nexts, nil
}
