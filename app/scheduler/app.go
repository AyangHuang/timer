package scheduler

import (
	"context"
	"timer/service/scheduler"
)

type WorkerApp struct {
	service workerService
	ctx     context.Context
	stop    func()
}

func NewWorkerApp(service *scheduler.Worker) *WorkerApp {
	return &WorkerApp{
		service: service,
	}
}

func (app *WorkerApp) Start() {
	app.ctx = context.Background()
	go func() {
		if err := app.service.Start(app.ctx); err != nil {
			panic(err)
		}
	}()
	return
}

var _ workerService = &scheduler.Worker{}

type workerService interface {
	Start(ctx context.Context) error
}
