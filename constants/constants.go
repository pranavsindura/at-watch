package constants

const (
	ModeDevelopment string = "DEVELOPMENT"
	ModeProduction  string = "PRODUCTION"
)

const (
	Yes string = "YES"
	No  string = "NO"
)

const (
	AccessLevelCustom = iota
	AccessLevelNewUser
	AccessLevelUser
	AccessLevelAdmin
	AccessLevelCreator
)

const (
	AccessLevelCustomText  string = "CUSTOM"
	AccessLevelNewUserText string = "NEW_USER"
	AccessLevelUserText    string = "USER"
	AccessLevelAdminText   string = "ADMIN"
	AccessLevelCreatorText string = "CREATOR"
)

var AccessLevelToTextMap = map[int]string{
	AccessLevelCustom:  AccessLevelCustomText,
	AccessLevelNewUser: AccessLevelNewUserText,
	AccessLevelUser:    AccessLevelUserText,
	AccessLevelAdmin:   AccessLevelAdminText,
	AccessLevelCreator: AccessLevelCreatorText,
}
