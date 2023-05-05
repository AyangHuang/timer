package app

import (
	"go.uber.org/dig"
	"timer/common/conf"
	"timer/pkg/mysql"
	"timer/pkg/pool"
	"timer/pkg/redis"
)

var contain *dig.Container

func init() {
	contain = dig.New()
	provideConf()
	providerPKG()
}

func provideConf() {
	contain.Provide(conf.GetDefaultRedisConfig)
	contain.Provide(conf.GetDefaultMySQLConfig)
	contain.Provide(conf.GetDefaultPoolConfig)
}

func providerPKG() {
	contain.Provide(mysql.GetClient)
	contain.Provide(redis.GetClient)
	contain.Provide(pool.NewGoWorkerPool)
}
