package generator

import (
	"encoding/json"
)

func GenerateLoginState(telegramUserID int64) string {
	state := map[string]int64{
		"telegramUserID": telegramUserID,
	}
	stateString, _ := json.Marshal(state)
	return string(stateString)
}
