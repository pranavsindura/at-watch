package publicRouter

import "github.com/gin-gonic/gin"

func AddRoutesToRouter(router *gin.Engine) {
	publicRouterGroup := router.Group("/public")
	publicRouterGroup.GET("/ping", Ping)
	// publicRouterGroup.GET("/login", Login)
	publicRouterGroup.GET("/redirect", Redirect)
	publicRouterGroup.GET("/test", Test)
}
