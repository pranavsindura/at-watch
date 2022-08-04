package publicRouter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pranavsindura/at-watch/cache"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
)

func redirect(authCode string) (gin.H, error) {
	accessToken, err := fyersSDK.ValidateAuthCode(authCode)
	if err != nil {
		return gin.H{}, err
	}

	fyersSDK.SetFyersAccessToken(accessToken)
	cache.SetFyersAccessToken(accessToken)

	return gin.H{"accessToken": accessToken}, nil
}

func Redirect(ctx *gin.Context) {
	authCode, authCodeExists := ctx.GetQuery("auth_code")

	if !authCodeExists {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("auth_code does not exist"))
		return
	}

	_, err := redirect(authCode)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><head><script>window.close()</script></head></html>"))
}
