package protocol //import "go.iondynamics.net/siteMgr/protocol"

import (
	"go.iondynamics.net/siteMgr/protocol/msgType"
)

const (
	Version = "0.7.0"
)

type Message struct {
	Type    msgType.Code
	Body    []byte
	Version string
}
