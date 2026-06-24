package http

import (
	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/calendar"
	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/bytedance/calandar/apps/api/internal/friends"
	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/bytedance/calandar/apps/api/internal/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config          config.Config
	DB              *gorm.DB
	EventService    *events.Service
	UserService     *users.Service
	ScheduleService *schedules.Service
	CalendarService *calendar.Service
	FriendService   *friends.Service
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
	router.GET("/e/:slug", eventHandler.HandleOGPage)

	scheduleService := deps.ScheduleService
	if scheduleService == nil {
		scheduleService = schedules.NewService(schedules.NewRepository(deps.DB))
	}
	scheduleHandler := schedules.NewHandler(scheduleService)
	authenticated.GET("/schedules", scheduleHandler.HandleList)
	authenticated.POST("/schedules", scheduleHandler.HandleCreate)
	authenticated.GET("/schedules/conflicts", scheduleHandler.HandleConflicts)
	authenticated.PATCH("/schedules/:id", scheduleHandler.HandleUpdate)
	authenticated.DELETE("/schedules/:id", scheduleHandler.HandleDelete)

	calendarService := deps.CalendarService
	if calendarService == nil {
		calendarService = calendar.NewService(schedules.NewRepository(deps.DB))
	}
	calendarHandler := calendar.NewHandler(calendarService)
	authenticated.GET("/calendar", calendarHandler.HandleMonth)

	friendService := deps.FriendService
	if friendService == nil {
		friendService = friends.NewService(friends.NewRepository(deps.DB))
	}
	friendHandler := friends.NewHandler(friendService)
	authenticated.GET("/friends", friendHandler.HandleList)
	authenticated.GET("/friends/requests", friendHandler.HandleListRequests)
	authenticated.POST("/friends/requests", friendHandler.HandleSendRequest)
	authenticated.POST("/friends/requests/:id/accept", friendHandler.HandleAccept)
	authenticated.POST("/friends/requests/:id/reject", friendHandler.HandleReject)
	authenticated.DELETE("/friends/:userId", friendHandler.HandleDelete)

	return router
}
