package registry //import "go.iondynamics.net/siteMgr/srv/registry"

import (
	"sync"

	"go.iondynamics.net/siteMgr"
)

var d map[string]chan siteMgr.Site = make(map[string]chan siteMgr.Site)
var m sync.RWMutex

func Get(key string) chan siteMgr.Site {
	m.RLock()
	defer m.RUnlock()

	return d[key]
}

func Set(key string, ch chan siteMgr.Site) {
	m.Lock()
	defer m.Unlock()

	d[key] = ch
}

func Del(key string) {
	m.Lock()
	defer m.Unlock()

	delete(d, key)
}
