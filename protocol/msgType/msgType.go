package msgType //import "go.iondynamics.net/siteMgr/protocol/msgType"

type Code uint8

const (
	HEARTBEAT Code = iota
	SITEMGR_SITE
	SITEMGR_USER
	LOGIN_SUCCESS
	LOGIN_ERROR
	LOGOUT
	INCOMPATIBLE
	NOTICE
	CLIPCONTENT
	DEC_CREDENTIALS
	ENC_CREDENTIALS
	CONNECTION_INFO
	METRICS
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

	case INCOMPATIBLE:
		return "INCOMPATIBLE"

	case NOTICE:
		return "NOTICE"

	case CLIPCONTENT:
		return "CLIPCONTENT"

	case DEC_CREDENTIALS:
		return "DEC_CREDENTIALS"

	case ENC_CREDENTIALS:
		return "ENC_CREDENTIALS"

	case CONNECTION_INFO:
		return "CONNECTION_INFO"

	case METRICS:
		return "METRICS"

	default:
		return "UNKNOWN"
	}
}
