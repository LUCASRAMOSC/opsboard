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
	serviceHandler *handler.ServiceHandler,
	healthEventHandler *handler.HealthEventHandler,
	businessJourneyHandler *handler.BusinessJourneyHandler,
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

	router.POST(
		"/workspaces",
		workspaceHandler.Create,
	)
	router.GET(
		"/workspaces",
		workspaceHandler.List,
	)
	router.GET(
		"/workspaces/:workspaceID",
		workspaceHandler.Get,
	)
	router.POST(
		"/workspaces/:workspaceID/services",
		serviceHandler.Create,
	)

	router.GET(
		"/workspaces/:workspaceID/services",
		serviceHandler.ListByWorkspace,
	)

	router.GET(
		"/services/:serviceID",
		serviceHandler.Get,
	)

	router.POST(
		"/services/:serviceID/health-events",
		healthEventHandler.Create,
	)

	router.GET(
		"/services/:serviceID/health-events",
		healthEventHandler.ListByService,
	)

	router.GET(
		"/services/:serviceID/health-events/latest",
		healthEventHandler.Latest,
	)

	router.POST(
		"/workspaces/:workspaceID/journeys",
		businessJourneyHandler.Create,
	)

	router.GET(
		"/workspaces/:workspaceID/journeys",
		businessJourneyHandler.ListByWorkspace,
	)

	router.GET(
		"/journeys/:journeyID",
		businessJourneyHandler.Get,
	)

	router.POST(
		"/journeys/:journeyID/services/:serviceID",
		businessJourneyHandler.AddService,
	)

	router.GET(
		"/journeys/:journeyID/services",
		businessJourneyHandler.ListServices,
	)

	router.DELETE(
		"/journeys/:journeyID/services/:serviceID",
		businessJourneyHandler.RemoveService,
	)

	return router, nil
}
