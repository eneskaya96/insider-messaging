package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/eneskaya/insider-messaging/internal/infrastructure/cache"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/persistence"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db    *persistence.PostgresGormDB
	redis *cache.RedisCache
}

func NewHealthHandler(db *persistence.PostgresGormDB, redis *cache.RedisCache) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check the health status of the application and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	if err := h.db.HealthCheck(ctx); err != nil {
		services["database"] = "unhealthy"
		allHealthy = false
	} else {
		services["database"] = "healthy"
	}

	if err := h.redis.HealthCheck(ctx); err != nil {
		services["redis"] = "unhealthy"
		allHealthy = false
	} else {
		services["redis"] = "healthy"
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:   status,
		Services: services,
	})
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Check if the application is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "ready",
	})
}

// LivenessCheck godoc
// @Summary Liveness check endpoint
// @Description Check if the application is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "alive",
	})
}
