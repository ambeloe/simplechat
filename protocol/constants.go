package protocol

const (
	TokenLength = 32
	KeyLength   = 32
	PWHashLen   = 32
)

const (
	_ = iota //0 is the default value in the protobuf
	StatusOK
	StatusDM
	StatusUsernameOK
	StatusUsernameTaken
	StatusGoodbye

	ErrInvalidCommand
	ErrIncorrectCredentials
	ErrInvalidKeyLength
	ErrUnauthorized
	ErrNotFound
)

const (
	_ = iota
	CmdGetUid
	CmdGetUsername
	CmdGetMsgs

	CmdRegCheckUsernameAvailability
	CmdRegRegister
	CmdRegLogin
	CmdLogout
)

func StatToString(stat uint32) string {
	switch stat {
	case StatusOK:
		return "OK"
	case StatusDM:
		return "Contains DMs"
	case StatusUsernameOK:
		return "Username is not taken"
	case StatusUsernameTaken:
		return "Username is taken"
	case StatusGoodbye:
		return "Goodbye"
	default:
		return "Unknown status"
	}
}

func ErrToString(err uint32) string {
	switch err {
	case ErrInvalidCommand:
		return "Invalid Command"
	case ErrIncorrectCredentials:
		return "Incorrect Credentials"
	case ErrInvalidKeyLength:
		return "Invalid Key Length"
	case ErrUnauthorized:
		return "Unauthorized"
	case ErrNotFound:
		return "Not Found"
	default:
		return "Unknown Error"
	}
}
