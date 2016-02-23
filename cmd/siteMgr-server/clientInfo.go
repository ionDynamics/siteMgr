package main

import (
	"encoding/json"
	html "html/template"
	"net"

	"github.com/GeorgeMac/idicon/icon"
	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr"
	reg "go.iondynamics.net/siteMgr/srv/registry"
)

type clientInfo struct {
	Identicon  html.HTML
	MsgVersion string
	Vendor     string
	Client     string
	Variant    string
	Address    string
}

func setIdenticon(name, hash string) {
	if hash != "" {
		props := icon.DefaultProps()
		props.Size = 21
		props.Padding = 0
		props.BorderRadius = 7
		generator, err := icon.NewGenerator(5, 5, icon.With(props))
		if err != nil {
			idl.Err(err)
			return
		}
		icn := generator.Generate([]byte(hash))

		mu.Lock()
		defer mu.Unlock()

		inf := cim[name]
		inf.Identicon = html.HTML(icn.String())
		cim[name] = inf
	}
}

func setClientMsgVersion(name, ver string) {
	mu.Lock()
	defer mu.Unlock()

	inf := cim[name]
	inf.MsgVersion = ver
	cim[name] = inf
}

func setClientInfo(name, vendor, client, variant, address string) {
	mu.Lock()
	defer mu.Unlock()

	inf := cim[name]
	inf.Vendor = vendor
	inf.Client = client
	inf.Variant = variant
	inf.Address = address
	cim[name] = inf
}

func getClientInfo(user string) clientInfo {
	mu.RLock()
	defer mu.RUnlock()
	return cim[user]
}

func delClientInfo(user string) {
	mu.Lock()
	defer mu.Unlock()

	delete(cim, user)
}

func saveClientInfo(msg *siteMgr.Message, c net.Conn, send chan siteMgr.Message) {
	var hash string
	usr := siteMgr.NewUser()
	json.Unmarshal(msg.Body, usr)

	if siteMgr.AtLeast("0.1.0", msg) {
		hash = usr.Sites["identicon-hash"].Name
		setClientMsgVersion(usr.Name, msg.Version)
	}
	if siteMgr.AtLeast("0.2.0", msg) {
		cl := usr.Sites["client"]

		host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
		setClientInfo(usr.Name, cl.Name, cl.Login, cl.Version, host)

	}

	setIdenticon(usr.Name, hash)
	reg.Set(usr.Name, send)
}

func cleanupClientInfo(usr *siteMgr.User) {
	if usr != nil && usr.Name != "" {
		delClientInfo(usr.Name)
		reg.Del(usr.Name)
	}
}

func getClientAddress(user string) string {
	return getClientInfo(user).Address
}
