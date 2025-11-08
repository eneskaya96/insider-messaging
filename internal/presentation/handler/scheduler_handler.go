package handler

import (
	"net/http"

	"github.com/eneskaya/insider-messaging/internal/application/dto"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/scheduler"
	"github.com/gin-gonic/gin"
)

type SchedulerHandler struct {
	scheduler *scheduler.Scheduler
}

func NewSchedulerHandler(scheduler *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler: scheduler,
	}
}

// StartScheduler godoc
// @Summary Start the message scheduler
// @Description Start automatic message sending process
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/scheduler/start [post]
func (h *SchedulerHandler) StartScheduler(c *gin.Context) {
	if h.scheduler.IsRunning() {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "scheduler is already running",
		})
		return
	}

	if err := h.scheduler.Start(c.Request.Context()); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "scheduler started successfully",
	})
}

// StopScheduler godoc
// @Summary Stop the message scheduler
// @Description Stop automatic message sending process
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/scheduler/stop [post]
func (h *SchedulerHandler) StopScheduler(c *gin.Context) {
	if !h.scheduler.IsRunning() {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "scheduler is not running",
		})
		return
	}

	if err := h.scheduler.Stop(); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "scheduler stopped successfully",
	})
}

// GetSchedulerStatus godoc
// @Summary Get scheduler status
// @Description Get current status and statistics of the message scheduler
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} dto.SchedulerStatusResponse
// @Router /api/v1/scheduler/status [get]
func (h *SchedulerHandler) GetSchedulerStatus(c *gin.Context) {
	lastRunAt, processed, successful, failed := h.scheduler.GetStats()

	c.JSON(http.StatusOK, dto.SchedulerStatusResponse{
		IsRunning:       h.scheduler.IsRunning(),
		LastRunAt:       lastRunAt,
		TotalProcessed:  processed,
		TotalSuccessful: successful,
		TotalFailed:     failed,
	})
}
