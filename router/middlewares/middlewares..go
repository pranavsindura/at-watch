package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

func DoesFyersAccessTokenExist(ctx *gin.Context) {
	if fyersSDK.GetFyersAccessToken() == "" {
		routerUtils.SendErrorResponse(ctx, http.StatusMethodNotAllowed, fmt.Errorf("fyers access token does not exist"))
		ctx.Abort()
	}
	ctx.Next()
}
