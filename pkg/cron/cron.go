package cron

import (
	"github.com/gorhill/cronexpr"
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
