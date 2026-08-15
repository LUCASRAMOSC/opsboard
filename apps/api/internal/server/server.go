package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/handler"
)

func New(
	db *pgxpool.Pool,
	workspaceHandler *handler.WorkspaceHandler,
) (*gin.Engine, error) {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	router.POST("/workspaces", workspaceHandler.Create)
	router.GET("/workspaces", workspaceHandler.List)
	router.GET("/workspaces/:workspaceID", workspaceHandler.Get)

	return router, nil
}
