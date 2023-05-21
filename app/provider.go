package app

import (
	"go.uber.org/dig"
	"timer/app/migrator"
	"timer/app/scheduler"
	"timer/app/webserver"
	"timer/common/conf"
	"timer/dao/task"
	mysqlDao "timer/dao/timer"
	"timer/pkg/bloom"
	"timer/pkg/cron"
	"timer/pkg/hash"
	"timer/pkg/mysql"
	"timer/pkg/redis"
	"timer/pkg/xhttp"
	executorservice "timer/service/executor"
	migratorservice "timer/service/migrator"
	schedulerservice "timer/service/scheduler"
	triggerservice "timer/service/trigger"
	"timer/service/webservice"
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
	contain.Provide(conf.GetDefaultMigratorAppConfig)
	contain.Provide(conf.GetDefaultSchedulerAppConfig)
	contain.Provide(conf.GetDefaultTriggerAppConfig)
	contain.Provide(conf.GetDefaultRedisConfig)
	contain.Provide(conf.GetDefaultMySQLConfig)
	contain.Provide(conf.GetDefaultWebServerAppConfig)
}

func providePKG() {
	contain.Provide(bloom.NewFilter)
	contain.Provide(hash.NewMurmur3Encryptor)
	contain.Provide(hash.NewSHA1Encryptor)
	contain.Provide(mysql.GetClient)
	contain.Provide(redis.GetClient)
	contain.Provide(cron.NewCronParser)
	contain.Provide(xhttp.NewJSONClient)
}

func provideDao() {
	contain.Provide(mysqlDao.NewTimerDao)
	contain.Provide(task.NewTaskDao)
	contain.Provide(task.NewTaskCache)
}

func provideServer() {
	contain.Provide(migratorservice.NewWorker)
	contain.Provide(webservice.NewTimerServer)
	contain.Provide(executorservice.NewTimerService)
	contain.Provide(executorservice.NewWorker)
	contain.Provide(triggerservice.NewWorker)
	contain.Provide(triggerservice.NewTaskService)
	contain.Provide(schedulerservice.NewWorker)
}

func provideHandler() {
	contain.Provide(webserver.NewTimerHandler)
	contain.Provide(webserver.NewTaskHandler)
}

func provideApp() {
	contain.Provide(migrator.NewMigratorApp)
	contain.Provide(webserver.NewServer)
	contain.Provide(scheduler.NewWorkerApp)
	contain.Provide(scheduler.NewWorkerApp)
}

func GetWebApp() *webserver.Server {
	var app *webserver.Server
	if err := contain.Invoke(func(server *webserver.Server) {
		app = server
	}); err != nil {
		panic(err)
	}
	return app
}

func GetSchedulerApp() *scheduler.WorkerApp {
	var app *scheduler.WorkerApp
	if err := contain.Invoke(func(workerApp *scheduler.WorkerApp) {
		app = workerApp
	}); err != nil {
		panic(err)
	}
	return app
}

func GetMigratorApp() *migrator.MigratorApp {
	var migratorApp *migrator.MigratorApp
	if err := contain.Invoke(func(_m *migrator.MigratorApp) {
		migratorApp = _m
	}); err != nil {
		panic(err)
	}
	return migratorApp
}
