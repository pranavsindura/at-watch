package apiRouter

import (
	"github.com/gin-gonic/gin"
	"github.com/pranavsindura/at-watch/router/middlewares"
)

func AddRoutesToRouter(router *gin.Engine) {
	apiRouterGroup := router.Group("/api")
	apiRouterGroup.Use(middlewares.DoesFyersAccessTokenExist)
	// apiRouterGroup.POST("/startMarketWatch", StartMarketWatch)
	// apiRouterGroup.POST("/stopMarketWatch", StopMarketWatch)
	// apiRouterGroup.POST("/backtest", Backtest)
}
