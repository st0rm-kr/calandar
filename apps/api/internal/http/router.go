package http

import (
	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/bytedance/calandar/apps/api/internal/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config       config.Config
	DB           *gorm.DB
	EventService *events.Service
	UserService  *users.Service
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.GET("/health", healthHandler)

	userService := deps.UserService
	if userService == nil {
		userService = users.NewService(users.NewRepository(deps.DB))
	}
	userHandler := users.NewHandler(userService)
	authenticated := api.Group("", auth.RequireUser(deps.Config.SupabaseJWTSecret))
	authenticated.GET("/auth/me", userHandler.HandleAuthMe)
	authenticated.GET("/users/me", userHandler.HandleGetMe)
	authenticated.PATCH("/users/me", userHandler.HandlePatchMe)
	authenticated.GET("/users/search", userHandler.HandleSearch)

	eventService := deps.EventService
	if eventService == nil {
		eventService = events.NewService(events.NewRepository(deps.DB))
	}
	eventHandler := events.NewHandler(eventService)
	authenticated.POST("/events", eventHandler.HandleCreate)
	authenticated.GET("/events/mine", eventHandler.HandleListMine)
	authenticated.POST("/events/:id/rsvp", eventHandler.HandleRSVP)
	authenticated.DELETE("/events/:id/rsvp", eventHandler.HandleDeleteRSVP)
	authenticated.POST("/events/:id/cancel", eventHandler.HandleCancel)
	api.GET("/events/:slug", eventHandler.HandleGetDetail)

	return router
}
