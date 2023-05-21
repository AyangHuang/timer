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

// NewServer title 和 version，不然不符合 swagger 标准，不能导入 postman
// swag init -g ./app/webservice/app.go 指定搜索该文件
// @title           timer API
// @version         0.0.0
// @host 127.0.0.1:8080
// @BasePath /api/dev
func NewServer(timerHandler *TimerHandler, taskHandler *TaskHandler, conf *conf.WebServerAppConfig) *Server {
	server := &Server{
		engine:       gin.Default(),
		timerHandler: timerHandler,
		taskHandler:  taskHandler,
		conf:         conf,
	}

	// 跨域和 设置 http header 头选项
	server.engine.Use(CrosHandler())

	baseGroup := server.engine.Group("api/dev")

	// 设置路由组
	server.timerRouter = baseGroup.Group("/timer")
	server.taskRouter = baseGroup.Group("/task")

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
	s.timerRouter.DELETE("/delete", s.timerHandler.DeleteTimer)
	s.timerRouter.POST("/enable", s.timerHandler.EnableTimer)
	s.timerRouter.POST("/unable", s.timerHandler.UnableTimer)
}

func (s *Server) registerTaskRouter() {

}
