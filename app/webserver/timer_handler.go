package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"timer/common/model/vo"
	"timer/pkg/logger"
	"timer/service/webservice"
)

type TimerHandler struct {
	timerServer timerServer
}

func NewTimerHandler(server *webservice.TimerServer) *TimerHandler {
	return &TimerHandler{
		timerServer: server,
	}
}

// CreateTimer 创建计时器
// @Summary      创建计时器
// @Description  创建计时器
// @Tags         创建计时器
// @Accept       json
// @Produce      json
// @Param        timer body vo.CreateTimerReq  true  "请求参数"
// @Success      200  {object}  vo.ResponseData{data=vo.CreateTimerRespData}
// @Router       /timer/create [post]
func (handler *TimerHandler) CreateTimer(ctx *gin.Context) {
	var err error

	var req vo.CreateTimerReq
	if err = ctx.ShouldBindJSON(&req); err != nil {
		vo.ResponseError(ctx, vo.CodeInvalidParam)
		return
	}

	// 业务处理：
	// 生成 timer 存入数据库中
	id, err := handler.timerServer.CreateTimer(ctx.Request.Context(), &req.Timer)
	if err != nil {
		logger.Errorf("%s", err)
		vo.ResponseError(ctx, vo.CodeServerBusy)
		return
	}

	vo.ResponseSuccess(ctx, vo.CreateTimerRespData{
		Id:      id,
		Success: true})
}

// DeleteTimer 删除计时器
// @Summary      删除计时器
// @Description  删除计时器
// @Tags         删除计时器
// @Accept       json
// @Produce      json
// @Param        timer body vo.TimerReq  true  "请求参数"
// @Success      200  {object}  vo.ResponseData{data=boolean}
// @Router       /timer/delete [delete]
func (handler *TimerHandler) DeleteTimer(ctx *gin.Context) {
	var err error

	var req vo.TimerReq
	if err = ctx.ShouldBindJSON(&req); err != nil {
		vo.ResponseError(ctx, vo.CodeInvalidParam)
		return
	}

	// 业务处理：直接删除数据库中的 timer 定义
	if err = handler.timerServer.DeleteTimer(ctx, req.ID); err != nil {
		vo.ResponseError(ctx, vo.CodeServerBusy)
		return
	}

	vo.ResponseSuccess(ctx, true)
}

// EnableTimer 激活计时器
// @Summary      激活计时器
// @Description  激活计时器
// @Tags         激活计时器
// @Accept       json
// @Produce      json
// @Param        timer body vo.TimerReq  true  "请求参数"
// @Success      200  {object}  vo.ResponseData{data=boolean}
// @Router       /timer/enable [post]
func (handler *TimerHandler) EnableTimer(ctx *gin.Context) {
	var err error

	var req vo.TimerReq
	if err = ctx.ShouldBindJSON(&req); err != nil {
		vo.ResponseError(ctx, vo.CodeInvalidParam)
		return
	}

	// 业务处理：
	// 创建两个一级迁移时间的 task，加入 MySQL 中，再加入 redis zset 中。
	if err = handler.timerServer.EnableTimer(ctx, req.ID); err != nil {
		logger.Errorf("%s", err)
		vo.ResponseError(ctx, vo.CodeServerBusy)
		return
	}

	vo.ResponseSuccess(ctx, true)
}

// UnableTimer 去激活计时器
// @Summary      去激活计时器
// @Description  去激活计时器
// @Tags         去激活计时器
// @Accept       json
// @Produce      json
// @Param        timer body vo.TimerReq  true  "请求参数"
// @Success      200  {object}  vo.ResponseData{data=boolean}
// @Router      /timer/unable [post]
func (handler *TimerHandler) UnableTimer(ctx *gin.Context) {
	var err error

	var req vo.TimerReq
	if err = ctx.ShouldBindJSON(&req); err != nil {
		vo.ResponseError(ctx, vo.CodeInvalidParam)
		return
	}

	// 业务处理：直接 update 数据库中 timer 定义的状态
	if err = handler.timerServer.UnableTimer(ctx, req.ID); err != nil {
		vo.ResponseError(ctx, vo.CodeServerBusy)
		return
	}

	vo.ResponseSuccess(ctx, true)
}

// 编译时检查
var _ timerServer = &webservice.TimerServer{}

type timerServer interface {
	CreateTimer(context.Context, *vo.Timer) (uint, error)
	DeleteTimer(ctx context.Context, id uint) error
	EnableTimer(ctx context.Context, id uint) error
	UnableTimer(ctx *gin.Context, id uint) error
}
