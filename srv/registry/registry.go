package registry //import "go.iondynamics.net/siteMgr/srv/registry"

import (
	"sync"

	"go.iondynamics.net/siteMgr/protocol"
)

var d map[string]chan protocol.Message = make(map[string]chan protocol.Message)
var m sync.RWMutex

func Get(key string) chan protocol.Message {
	m.RLock()
	defer m.RUnlock()

	return d[key]
}

func Set(key string, ch chan protocol.Message) {
	m.Lock()
	defer m.Unlock()

	d[key] = ch
}

func Del(key string) {
	m.Lock()
	defer m.Unlock()

	delete(d, key)
}
