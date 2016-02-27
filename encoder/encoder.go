package encoder //import "go.iondynamics.net/siteMgr/encoder"

import (
	"encoding/json"
	"errors"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/msgType"
)

var vers string

func Init(version string) {
	vers = version
}

func Do(v interface{}) (msg siteMgr.Message, err error) {
	if vers == "" {
		err = errors.New("not initialized")
		return
	}
	msg.Version = vers
	switch t := v.(type) {
	case siteMgr.Site, *siteMgr.Site:
		msg.Type = msgType.SITEMGR_SITE
		msg.Body, err = json.Marshal(t)

	case siteMgr.User, *siteMgr.User:
		msg.Type = msgType.SITEMGR_USER
		msg.Body, err = json.Marshal(t)

	case msgType.Code:
		msg.Type = t

	case string:
		msg.Type = msgType.NOTICE
		msg.Body = []byte(t)

	case siteMgr.Credentials, *siteMgr.Credentials:
		msg.Type = msgType.ENC_CREDENTIALS
		msg.Body, err = json.Marshal(t)

	default:
		err = errors.New("can't encode interface to message")
	}
	return
}
