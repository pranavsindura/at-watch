package publicRouter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
)

func Login(ctx *gin.Context) {
	loginUrl := fyersSDK.GenerateAuthCodeURL()
	ctx.Redirect(http.StatusPermanentRedirect, loginUrl)
}
