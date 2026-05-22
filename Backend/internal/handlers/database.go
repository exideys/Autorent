package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DatabaseHealthHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	}
}
