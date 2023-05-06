package app

import (
	"go.uber.org/dig"
	webApp "timer/app/webserver"
	"timer/common/conf"
	mysqlDao "timer/dao/mysql"
	"timer/pkg/cron"
	"timer/pkg/mysql"
	"timer/pkg/pool"
	"timer/pkg/redis"
	"timer/service/webserver"
)

var contain *dig.Container

func init() {
	contain = dig.New()
	provideConf()
	providePKG()
	provideDao()
	provideServer()
	provideHandler()
	provideApp()
}

func provideConf() {
	contain.Provide(conf.GetDefaultRedisConfig)
	contain.Provide(conf.GetDefaultMySQLConfig)
	contain.Provide(conf.GetDefaultPoolConfig)
	contain.Provide(conf.GetDefaultPoolConfig)
	contain.Provide(conf.GetDefaultWebServerAppConfig)
}

func providePKG() {
	contain.Provide(mysql.GetClient)
	contain.Provide(redis.GetClient)
	contain.Provide(pool.NewGoWorkerPool)
	contain.Provide(cron.NewCronParser)
}

func provideDao() {
	contain.Provide(mysqlDao.NewTimerDao)
}

func provideServer() {
	contain.Provide(webserver.NewTimerServer)
}

func provideHandler() {
	contain.Provide(webApp.NewTimerHandler)
	contain.Provide(webApp.NewTaskHandler)
}

func provideApp() {
	contain.Provide(webApp.NewServer)
}

func GerWebApp() *webApp.Server {
	var app *webApp.Server
	if err := contain.Invoke(func(server *webApp.Server) {
		app = server
	}); err != nil {
		panic(err)
	}
	return app
}
