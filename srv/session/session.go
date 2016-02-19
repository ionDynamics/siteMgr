package session //import "go.iondynamics.net/siteMgr/srv/session"

import (
	"sync"
	"time"

	"go.iondynamics.net/iDhelper/randGen"

	"go.iondynamics.net/siteMgr"
)

var d map[string]*siteMgr.User = make(map[string]*siteMgr.User)
var m sync.RWMutex

func Start(usr *siteMgr.User) string {
	k := randGen.String(64)
	Set(k, usr)
	Timeout(k, 10*time.Hour)
	return k
}

func Get(key string) *siteMgr.User {
	m.RLock()
	defer m.RUnlock()

	return d[key]
}

func Set(key string, usr *siteMgr.User) {
	m.Lock()
	defer m.Unlock()

	d[key] = usr
}

func Del(key string) {
	m.Lock()
	defer m.Unlock()

	delete(d, key)
}

func Timeout(key string, dur time.Duration) {
	go func(k string, d time.Duration) {
		<-time.After(d)
		Del(k)
	}(key, dur)
}
