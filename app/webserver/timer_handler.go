package webserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"timer/common/model/vo"
	"timer/pkg/logger"
	"timer/service/webserver"
)

type TimerHandler struct {
	timerServer timerServer
}

func NewTimerHandler(server *webserver.TimerServer) *TimerHandler {
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
// @Param        timer body vo.CreateTimerReq  true  "创建 timer 的完整定义"
// @Success      200  {object}  vo.ResponseData{Data=vo.CreateTimerRespData}
// @Router       /api/timer/v1/timer/create [post]
func (handler *TimerHandler) CreateTimer(ctx *gin.Context) {
	var err error

	var req vo.CreateTimerReq
	if err = ctx.ShouldBindJSON(&req); err != nil {
		vo.ResponseError(ctx, vo.CodeInvalidParam)
	}

	id, err := handler.timerServer.CreateTimer(ctx.Request.Context(), &req.Timer)
	if err != nil {
		logger.Errorf("mysql err:%s", err)
		vo.ResponseError(ctx, vo.CodeServerBusy)
		return
	}

	vo.ResponseSuccess(ctx, vo.CreateTimerRespData{
		Id:      id,
		Success: true})
}

// 编译时检查
var _ timerServer = &webserver.TimerServer{}

type timerServer interface {
	CreateTimer(context.Context, *vo.Timer) (uint, error)
}
