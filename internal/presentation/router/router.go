package router

import (
	"github.com/eneskaya/insider-messaging/internal/presentation/handler"
	"github.com/eneskaya/insider-messaging/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	engine            *gin.Engine
	messageHandler    *handler.MessageHandler
	schedulerHandler  *handler.SchedulerHandler
	healthHandler     *handler.HealthHandler
}

func NewRouter(
	messageHandler *handler.MessageHandler,
	schedulerHandler *handler.SchedulerHandler,
	healthHandler *handler.HealthHandler,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	return &Router{
		engine:            engine,
		messageHandler:    messageHandler,
		schedulerHandler:  schedulerHandler,
		healthHandler:     healthHandler,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.GET("/health", r.healthHandler.HealthCheck)
	r.engine.GET("/ready", r.healthHandler.ReadinessCheck)
	r.engine.GET("/live", r.healthHandler.LivenessCheck)

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.engine.Group("/api/v1")
	{
		scheduler := v1.Group("/scheduler")
		{
			scheduler.POST("/start", r.schedulerHandler.StartScheduler)
			scheduler.POST("/stop", r.schedulerHandler.StopScheduler)
			scheduler.GET("/status", r.schedulerHandler.GetSchedulerStatus)
		}

		messages := v1.Group("/messages")
		{
			messages.GET("/sent", r.messageHandler.GetSentMessages)
			messages.GET("/stats", r.messageHandler.GetStats)
			messages.GET("/:id", r.messageHandler.GetMessage)
			messages.POST("", r.messageHandler.CreateMessage)
		}
	}

	return r.engine
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
