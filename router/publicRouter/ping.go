package publicRouter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

func ping() (gin.H, error) {
	data := gin.H{
		"pong": true,
	}
	return data, nil
}

func Ping(ctx *gin.Context) {
	data, err := ping()

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	routerUtils.SendSuccessResponse(ctx, http.StatusOK, data)
}
