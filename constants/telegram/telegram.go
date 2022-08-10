package telegramConstants

import "github.com/pranavsindura/at-watch/constants"

// var CreatorUserNames []string = []string{"pranavsindura"}
// var CreatorTelegramUserIDs []int32 = []int32{716208519}

const (
	CommandPing              string = "ping"
	CommandStart             string = "start"
	CommandStop              string = "stop"
	CommandLogin             string = "login"
	CommandMe                string = "me"
	CommandUpdateAccessLevel string = "updateaccesslevel"
	CommandBacktest          string = "backtest"
	CommandMaintenance       string = "maintenance"
	CommandAddStrategy       string = "addstrategy"
	CommandRemoveStrategy    string = "removestrategy"
	CommandGetStrategy       string = "getstrategy"
	CommandPauseStrategy     string = "pausestrategy"
	CommandResumeStrategy    string = "resumestrategy"
	CommandAddInstrument     string = "addinstrument"
	CommandRemoveInstrument  string = "removeinstrument"
	CommandGetInstrument     string = "getinstrument"
	CommandRenameInstrument  string = "renameinstrument"
	CommandStartMarket       string = "startmarket"
	CommandStopMarket        string = "stopmarket"
	CommandEnterStrategy     string = "enterstrategy"
	CommandExitStrategy      string = "exitstrategy"
	CommandGetOpenTrade      string = "getopentrade"
	CommandGetClosedTrade    string = "getclosedtrade"
)

var MinimumAccessLevel = map[string]int{
	CommandPing:              constants.AccessLevelNewUser,
	CommandStart:             constants.AccessLevelNone,
	CommandStop:              constants.AccessLevelNewUser,
	CommandLogin:             constants.AccessLevelAdmin,
	CommandMe:                constants.AccessLevelNone,
	CommandUpdateAccessLevel: constants.AccessLevelNone,
	CommandBacktest:          constants.AccessLevelUser,
	CommandMaintenance:       constants.AccessLevelAdmin,
	CommandAddStrategy:       constants.AccessLevelUser,
	CommandRemoveStrategy:    constants.AccessLevelUser,
	CommandGetStrategy:       constants.AccessLevelUser,
	CommandPauseStrategy:     constants.AccessLevelUser,
	CommandResumeStrategy:    constants.AccessLevelUser,
	CommandAddInstrument:     constants.AccessLevelAdmin,
	CommandRemoveInstrument:  constants.AccessLevelAdmin,
	CommandGetInstrument:     constants.AccessLevelUser,
	CommandRenameInstrument:  constants.AccessLevelAdmin,
	CommandStartMarket:       constants.AccessLevelAdmin,
	CommandStopMarket:        constants.AccessLevelAdmin,
	CommandEnterStrategy:     constants.AccessLevelUser,
	CommandExitStrategy:      constants.AccessLevelUser,
	CommandGetOpenTrade:      constants.AccessLevelUser,
	CommandGetClosedTrade:    constants.AccessLevelUser,
}
