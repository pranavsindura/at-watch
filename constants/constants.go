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
	AccessLevelNone = iota
	AccessLevelNewUser
	AccessLevelUser
	AccessLevelAdmin
	AccessLevelCreator
)

const (
	AccessLevelNoneText    string = "NONE"
	AccessLevelNewUserText string = "NEW_USER"
	AccessLevelUserText    string = "USER"
	AccessLevelAdminText   string = "ADMIN"
	AccessLevelCreatorText string = "CREATOR"
)

var AccessLevelToTextMap = map[int]string{
	AccessLevelNone:    AccessLevelNoneText,
	AccessLevelNewUser: AccessLevelNewUserText,
	AccessLevelUser:    AccessLevelUserText,
	AccessLevelAdmin:   AccessLevelAdminText,
	AccessLevelCreator: AccessLevelCreatorText,
}
