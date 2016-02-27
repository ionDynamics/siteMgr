package msgType //import "go.iondynamics.net/siteMgr/msgType"

type Code uint8

const (
	HEARTBEAT Code = iota
	SITEMGR_SITE
	SITEMGR_USER
	LOGIN_SUCCESS
	LOGIN_ERROR
	LOGOUT
	UPDATE_AVAIL
	NOTICE
	CLIPCONTENT
	DEC_CREDENTIALS
	ENC_CREDENTIALS
)

func (c Code) String() string {
	switch c {
	case HEARTBEAT:
		return "HEARTBEAT"
	case SITEMGR_SITE:
		return "SITEMGR_SITE"

	case SITEMGR_USER:
		return "SITEMGR_USER"

	case LOGIN_SUCCESS:
		return "LOGIN_SUCCESS"

	case LOGIN_ERROR:
		return "LOGIN_ERROR"

	case LOGOUT:
		return "LOGOUT"

	case NOTICE:
		return "NOTICE"

	case CLIPCONTENT:
		return "CLIPCONTENT"

	case DEC_CREDENTIALS:
		return "DEC_CREDENTIALS"

	case ENC_CREDENTIALS:
		return "ENC_CREDENTIALS"

	default:
		return "UNKNOWN"
	}
}
