package apiRouter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

type StartMarketWatchRequestBody struct {
	Instruments []string `json:"instruments"`
}

func startMarketWatch(instruments []string) (gin.H, error) {
	fmt.Println(fyersSDK.GetFyersAccessToken(), fyersSDK.IsMarketWatchActive())
	if fyersSDK.GetFyersAccessToken() == "" {
		return gin.H{}, fmt.Errorf("access token not found")
	}
	if fyersSDK.IsMarketWatchActive() {
		return gin.H{}, fmt.Errorf("market watch is already active")
	}

	// fyersSDK.StartMarketWatch(instruments)

	return gin.H{}, nil

}

func StartMarketWatch(ctx *gin.Context) {
	var body StartMarketWatchRequestBody
	err := ctx.BindJSON(&body)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("not able to parse instruments"))
		return
	}

	data, err := startMarketWatch(body.Instruments)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	routerUtils.SendSuccessResponse(ctx, http.StatusOK, data)
}
