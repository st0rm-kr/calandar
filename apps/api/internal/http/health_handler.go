package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
	respondOK(c, http.StatusOK, gin.H{"status": "ok"})
}
