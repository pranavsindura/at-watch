package apiRouter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

func stopMarketWatch() (gin.H, error) {
	if fyersSDK.IsMarketWatchActive() {
		fyersSDK.StopMarketWatch()
		return gin.H{}, nil
	}

	return gin.H{}, fmt.Errorf("market watch is not active")
}

func StopMarketWatch(ctx *gin.Context) {
	data, err := stopMarketWatch()

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	routerUtils.SendSuccessResponse(ctx, http.StatusOK, data)
}
