package main

import (
	html "html/template"
	"net"
	"sync"

	"github.com/GeorgeMac/idicon/icon"
	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/protocol"
	"go.iondynamics.net/siteMgr/srv"
	reg "go.iondynamics.net/siteMgr/srv/registry"
)

var (
	mu  sync.RWMutex
	cim map[string]*siteMgr.ConnectionInfo = make(map[string]*siteMgr.ConnectionInfo)
)

func identicon(ci *siteMgr.ConnectionInfo) html.HTML {
	if len(ci.IdenticonHash) > 0 {
		props := icon.DefaultProps()
		props.Size = 21
		props.Padding = 0
		props.BorderRadius = 7
		generator, err := icon.NewGenerator(5, 5, icon.With(props))
		if err != nil {
			idl.Err(err)
			return html.HTML("")
		}
		icn := generator.Generate(ci.IdenticonHash)
		return html.HTML(icn.String())
	}
	return html.HTML("")
}

func connectionInfo(user string) *siteMgr.ConnectionInfo {
	mu.RLock()
	defer mu.RUnlock()
	return cim[user]
}

func delConnectionInfo(user string) {
	mu.Lock()
	defer mu.Unlock()

	delete(cim, user)
}

func setConnectionInfo(usr *srv.User, info *siteMgr.ConnectionInfo, c net.Conn, send chan protocol.Message) {
	mu.Lock()
	defer mu.Unlock()

	host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
	info.RemoteAddress = host
	cim[usr.Name] = info

	reg.Set(usr.Name, send)
}

func cleanupConnectionInfo(usr *srv.User) {
	if usr != nil && usr.Name != "" {
		delConnectionInfo(usr.Name)
		reg.Del(usr.Name)
	}
}
