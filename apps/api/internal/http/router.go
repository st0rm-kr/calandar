package http

import "github.com/gin-gonic/gin"

type Dependencies struct{}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	api.GET("/health", healthHandler)
	return router
}
