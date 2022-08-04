package router

import (
	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
	"github.com/pranavsindura/at-watch/router/apiRouter"
	"github.com/pranavsindura/at-watch/router/publicRouter"
)

func New() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ginzerolog.Logger("gin"))

	apiRouter.AddRoutesToRouter(router)
	publicRouter.AddRoutesToRouter(router)

	return router
}
