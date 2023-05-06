package webserver

type TaskHandler struct {
	taskServer taskServer
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

type taskServer interface {
}
