package http

import (
	"github.com/bytedance/calandar/apps/api/internal/auth"
	"github.com/bytedance/calandar/apps/api/internal/calendar"
	"github.com/bytedance/calandar/apps/api/internal/config"
	"github.com/bytedance/calandar/apps/api/internal/events"
	"github.com/bytedance/calandar/apps/api/internal/friends"
	"github.com/bytedance/calandar/apps/api/internal/groups"
	"github.com/bytedance/calandar/apps/api/internal/inbox"
	"github.com/bytedance/calandar/apps/api/internal/notifications"
	"github.com/bytedance/calandar/apps/api/internal/schedules"
	"github.com/bytedance/calandar/apps/api/internal/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config              config.Config
	DB                  *gorm.DB
	EventService        *events.Service
	UserService         *users.Service
	ScheduleService     *schedules.Service
	CalendarService     *calendar.Service
	FriendService       *friends.Service
	GroupService        *groups.Service
	NotificationService *notifications.Service
	InboxService        *inbox.Service
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

	friendService := deps.FriendService
	if friendService == nil {
		friendService = friends.NewService(friends.NewRepository(deps.DB))
	}
	groupService := deps.GroupService
	if groupService == nil {
		groupService = groups.NewService(groups.NewRepository(deps.DB))
	}

	eventService := deps.EventService
	if eventService == nil {
		eventService = events.NewService(
			events.NewRepository(deps.DB),
			events.WithFriendAuthorizer(friendService),
			events.WithGroupAuthorizer(groupService),
		)
	}
	eventHandler := events.NewHandler(eventService)
	authenticated.POST("/events", eventHandler.HandleCreate)
	authenticated.GET("/events/mine", eventHandler.HandleListMine)
	authenticated.POST("/events/:id/rsvp", eventHandler.HandleRSVP)
	authenticated.DELETE("/events/:id/rsvp", eventHandler.HandleDeleteRSVP)
	authenticated.POST("/events/:id/cancel", eventHandler.HandleCancel)
	authenticated.POST("/events/:id/invite", eventHandler.HandleInvite)
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

	friendHandler := friends.NewHandler(friendService)
	authenticated.GET("/friends", friendHandler.HandleList)
	authenticated.GET("/friends/requests", friendHandler.HandleListRequests)
	authenticated.POST("/friends/requests", friendHandler.HandleSendRequest)
	authenticated.POST("/friends/requests/:id/accept", friendHandler.HandleAccept)
	authenticated.POST("/friends/requests/:id/reject", friendHandler.HandleReject)
	authenticated.DELETE("/friends/:userId", friendHandler.HandleDelete)

	groupHandler := groups.NewHandler(groupService)
	authenticated.GET("/groups", groupHandler.HandleList)
	authenticated.POST("/groups", groupHandler.HandleCreate)
	authenticated.POST("/groups/join", groupHandler.HandleJoin)
	authenticated.GET("/groups/:id", groupHandler.HandleGetDetail)
	authenticated.PATCH("/groups/:id", groupHandler.HandleUpdate)
	authenticated.DELETE("/groups/:id", groupHandler.HandleDissolve)
	authenticated.DELETE("/groups/:id/members/:userId", groupHandler.HandleLeave)
	authenticated.POST("/group-invites", groupHandler.HandleInvite)
	authenticated.POST("/group-invites/:id/accept", groupHandler.HandleAcceptInvite)
	authenticated.POST("/group-invites/:id/reject", groupHandler.HandleRejectInvite)

	notificationService := deps.NotificationService
	if notificationService == nil {
		notificationService = notifications.NewService(notifications.NewRepository(deps.DB))
	}
	notificationHandler := notifications.NewHandler(notificationService)
	authenticated.GET("/notifications", notificationHandler.HandleList)
	authenticated.POST("/notifications/read-all", notificationHandler.HandleMarkAllRead)
	authenticated.POST("/notifications/:id/read", notificationHandler.HandleMarkRead)

	inboxService := deps.InboxService
	if inboxService == nil {
		inboxService = inbox.NewService(inbox.NewRepository(deps.DB))
	}
	inboxHandler := inbox.NewHandler(inboxService)
	authenticated.GET("/inbox", inboxHandler.HandleList)

	return router
}
