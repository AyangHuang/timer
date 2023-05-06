package webserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"timer/common/conf"
	_ "timer/docs"
)

type Server struct {
	engine *gin.Engine

	timerHandler *TimerHandler
	taskHandler  *TaskHandler

	timerRouter *gin.RouterGroup
	taskRouter  *gin.RouterGroup

	conf *conf.WebServerAppConfig
}

func NewServer(timerHandler *TimerHandler, taskHandler *TaskHandler, conf *conf.WebServerAppConfig) *Server {
	server := &Server{
		engine:       gin.Default(),
		timerHandler: timerHandler,
		taskHandler:  taskHandler,
		conf:         conf,
	}

	// 跨域和 设置 http header 头选项
	server.engine.Use(CrosHandler())

	// 设置路由组
	server.timerRouter = server.engine.Group("api/timer/v1/timer")
	server.taskRouter = server.engine.Group("api/task/v1/task")

	// 注册路由
	// swagger
	server.registerBaseRouter()

	server.registerTimerRouter()
	server.registerTaskRouter()

	return server
}

func (s *Server) Start() {
	go func() {
		// 开启 web 服务端
		if err := s.engine.Run(fmt.Sprintf(":%d", s.conf.Port)); err != nil {
			panic(err)
		}
	}()
}

func (s *Server) registerBaseRouter() {
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (s *Server) registerTimerRouter() {
	s.timerRouter.POST("/create", s.timerHandler.CreateTimer)
}

func (s *Server) registerTaskRouter() {

}
