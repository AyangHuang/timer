package webserver

import (
	"context"
	"timer/common/model/po"
	"timer/common/model/vo"
	"timer/dao/mysql"
	"timer/pkg/cron"
)

type TimerServer struct {
	dao        timerDao
	cronParser cronParser
}

func NewTimerServer(dao *mysql.TimerDao, parser *cron.Parser) *TimerServer {
	return &TimerServer{
		dao:        dao,
		cronParser: parser,
	}
}

func (t *TimerServer) CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error) {
	// 判断 cron 表达式是否有效
	if !t.cronParser.IsValidCronExpr(timer.Cron) {
		return 0, vo.ErrCronExprUnValid
	}

	// 转换成数据库映射 struct
	poTimer, err := timer.ToPo()
	if err != nil {
		return 0, err
	}

	return t.dao.CreateTimer(ctx, poTimer)
}

var _ timerDao = &mysql.TimerDao{}
var _ cronParser = &cron.Parser{}

type timerDao interface {
	CreateTimer(context.Context, *po.Timer) (uint, error)
}

type cronParser interface {
	IsValidCronExpr(string) bool
}
