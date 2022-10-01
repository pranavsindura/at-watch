package publicRouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pranavsindura/at-watch/cache"
	"github.com/pranavsindura/at-watch/constants"
	fyersConstants "github.com/pranavsindura/at-watch/constants/fyers"
	telegramUserModel "github.com/pranavsindura/at-watch/models/telegramUser"
	fyersSDK "github.com/pranavsindura/at-watch/sdk/fyers"
	"github.com/pranavsindura/at-watch/sdk/notifications"
	routerUtils "github.com/pranavsindura/at-watch/utils/router"
	"go.mongodb.org/mongo-driver/bson"
)

func redirect(authCode string, telegramUserID int64) (gin.H, error) {
	accessToken, err := fyersSDK.ValidateAuthCode(authCode)
	if err != nil {
		return gin.H{}, err
	}

	if telegramUserID == fyersConstants.AdminTelegramUserID {
		fyersSDK.SetFyersAccessToken(accessToken)
		cache.SetFyersAccessToken(fyersConstants.AdminTelegramUserID, accessToken)
		notifications.Broadcast(constants.AccessLevelAdmin, "Successfully set Admin Fyers Access Token")
	} else {
		// set it for one particular user
		getResult := telegramUserModel.GetTelegramUserCollection().FindOne(context.Background(), bson.M{"telegramUserID": telegramUserID})

		err := getResult.Err()
		if err != nil {
			return gin.H{}, fmt.Errorf("invalid user")
		}

		user := &telegramUserModel.TelegramUserModel{}
		getResult.Decode(&user)

		cache.SetFyersAccessToken(telegramUserID, accessToken)
		notifications.Notify(user.TelegramChatID, "Login Successful")
	}

	return gin.H{}, nil
}

func Redirect(ctx *gin.Context) {
	authCode, authCodeExists := ctx.GetQuery("auth_code")
	stateString, stateStringExists := ctx.GetQuery("state")

	if !authCodeExists {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("auth_code does not exist"))
		return
	}
	if !stateStringExists {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("state does not exist"))
		return
	}

	state, err := url.QueryUnescape(stateString)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("unable to parse state: "+stateString))
		return
	}

	stateMap := make(map[string]int64)

	err = json.Unmarshal([]byte(state), &stateMap)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("unable to parse state after unmarshal: "+state))
		return
	}

	telegramUserID, telegramUserIDExists := stateMap["telegramUserID"]

	if !telegramUserIDExists {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid state"))
		return
	}

	_, err = redirect(authCode, telegramUserID)

	if err != nil {
		routerUtils.SendErrorResponse(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<html><head><script>window.close()</script></head><body><div id="done">OK</div></body></html>`))
}
