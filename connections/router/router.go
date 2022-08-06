package routerClient

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pranavsindura/at-watch/constants"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	"github.com/pranavsindura/at-watch/router"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	"github.com/rs/zerolog/log"
)

var routerClient *gin.Engine

func Init() {
	log.Info().Msg("init router")
	router := router.New()
	Port := os.Getenv(envConstants.Port)
	notifications.Broadcast(constants.AccessLevelAdmin, "Server is Listening on PORT: "+Port)
	routerClient = router
	err := router.Run(":" + Port)

	if err != nil {
		log.Fatal().Err(err)
	}
}

func Client() *gin.Engine {
	return routerClient
}
