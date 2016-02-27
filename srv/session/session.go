package session //import "go.iondynamics.net/siteMgr/srv/session"

import (
	"sync"
	"time"

	"go.iondynamics.net/iDhelper/randGen"

	"go.iondynamics.net/siteMgr"
)

var m sync.RWMutex
var t map[string]*siteMgr.User = make(map[string]*siteMgr.User)
var n map[string]*siteMgr.User = make(map[string]*siteMgr.User)

func Start(usr *siteMgr.User) string {
	k := randGen.String(64)
	Set(k, usr)
	Timeout(k, 10*time.Hour)
	return k
}

func Get(key string) *siteMgr.User {
	m.RLock()
	defer m.RUnlock()

	return t[key]
}

func GetByName(name string) *siteMgr.User {
	m.RLock()
	defer m.RUnlock()

	return n[name]
}

func GetKeyByName(name string) string {
	m.RLock()
	defer m.RUnlock()

	for key, usr := range t {
		if usr.Name == name {
			return key
		}
	}
	return ""
}

func Set(key string, usr *siteMgr.User) {
	m.Lock()
	defer m.Unlock()

	t[key] = usr
	n[usr.Name] = usr
}

func Del(key string) {
	m.Lock()
	defer m.Unlock()

	delete(t, key)
}

func Timeout(key string, dur time.Duration) {
	go func(k string, d time.Duration) {
		<-time.After(d)
		Del(k)
	}(key, dur)
}
