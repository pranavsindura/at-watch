package routerClient

import (
	"os"

	"github.com/gin-gonic/gin"
	envConstants "github.com/pranavsindura/at-watch/constants/env"
	"github.com/pranavsindura/at-watch/router"
	"github.com/rs/zerolog/log"
)

var routerClient *gin.Engine

func Init() {
	log.Info().Msg("init router")
	router := router.New()
	Port := os.Getenv(envConstants.Port)
	err := router.Run(":" + Port)

	if err != nil {
		log.Fatal().Err(err)
	}

	routerClient = router
}

func Client() *gin.Engine {
	return routerClient
}
