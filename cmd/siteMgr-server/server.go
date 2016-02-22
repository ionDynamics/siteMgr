package main

import (
	"encoding/json"
	"flag"
	html "html/template"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/GeorgeMac/idicon/icon"
	"github.com/labstack/echo"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"go.iondynamics.net/iDechoLog"
	idl "go.iondynamics.net/iDlogger"
	kv "go.iondynamics.net/kvStore"
	"go.iondynamics.net/kvStore/backend/bolt"
	"go.iondynamics.net/templice"

	"go.iondynamics.net/siteMgr"
	reg "go.iondynamics.net/siteMgr/srv/registry"
	"go.iondynamics.net/siteMgr/srv/route"
	"go.iondynamics.net/siteMgr/srv/template"
)

var (
	debug          = flag.Bool("debug", false, "Enable debug output")
	httpListener   = flag.String("http", ":9211", "Bind HTTP-Listener to ...")
	clientListener = flag.String("client", ":9210", "Bind Client-Listener to ...")
	autoupdate     = flag.Bool("autoupdate", true, "Enable or disable automatic updates")

	mu  sync.RWMutex
	uim map[string]clientInfo = make(map[string]clientInfo)

	key = []byte{0x42, 0x89, 0x10, 0x01, 0x07, 0xAC, 0xDC, 0x77, 0x70, 0x07, 0x66, 0x6B, 0xCD, 0xFF, 0x13, 0xCC}

	VERSION = "0.2.0"
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	idl.Info("Version: ", VERSION)
	idl.Info("Starting...")

	update()

	setupPersistence()

	setupHttp()

	runClientListener()
}

func handleConnection(c net.Conn) {
	defer c.Close()

	enc := json.NewEncoder(c)
	dec := json.NewDecoder(c)
	msg := &siteMgr.Message{}
	usr := siteMgr.NewUser()

	var hash string

	var authed bool
	for !authed && dec.More() {
		if err := dec.Decode(msg); err != nil {
			idl.Err("decoding message:", err)
		}
		err := json.Unmarshal(msg.Body, usr)
		if err != nil {
			idl.Err("unmarshal body:", err)
			idl.Debug("ouch:", string(msg.Body))
			return
		}

		if siteMgr.AtLeast("0.1.0", msg) {
			hash = usr.Sites["identicon-hash"].Name
			setClientMsgVersion(usr.Name, msg.Version)
		}
		if siteMgr.AtLeast("0.2.0", msg) {
			cl := usr.Sites["client"]

			host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			setClientInfo(usr.Name, cl.Name, cl.Login, cl.Version, host)
		}

		msg.Version = VERSION
		msg.Body = []byte{}
		if usr.Login() == nil {
			idl.Debug("Login success")
			authed = true
			msg.Type = "LOGIN:SUCCESS"
		} else {
			msg.Type = "LOGIN:ERROR"
		}

		idl.Debug("Sending message", msg)
		err = enc.Encode(msg)
		if err != nil {
			idl.Err(err)
		}
	}

	if !authed {
		return
	}

	idl.Debug("User Client logged in: ", usr.Name)

	closed := make(chan bool)
	go func() {
		gomsg := &siteMgr.Message{}
		for dec.More() {
			err := dec.Decode(gomsg)
			if err != nil {
				continue
			}
			if gomsg.Type == "LOGOUT" {
				idl.Debug("client logout: ", usr.Name)
				break
			}
		}
		closed <- true
	}()

	setIdenticon(usr.Name, hash)

	ch := make(chan siteMgr.Site, 1)
	idl.Debug("registering channel ", usr.Name, ch)
	reg.Set(usr.Name, ch)

loop:
	for {
		idl.Debug("ready for sending messages")
		var err error
		select {
		case s := <-ch:
			idl.Debug("channel sent site")

			msg.Type = "siteMgr.Site"
			msg.Body, err = json.Marshal(s)
			if err != nil {
				idl.Err(err)
				continue
			}
		case <-time.After(3 * time.Minute):
			msg.Type = "HEARTBEAT"
			msg.Body = []byte{}
		case <-closed:
			break loop
		}
		msg.Version = VERSION
		idl.Debug("Sending message", msg)
		err = enc.Encode(msg)
		if err != nil {
			idl.Warn(err)
			break
		}
	}

	reg.Del(usr.Name)
	idl.Debug("connection handler shutdown")
}

func update() {
	if *autoupdate {
		var updater = &selfupdate.Updater{
			CurrentVersion: VERSION,
			ApiURL:         "https://update.slpw.de/",
			BinURL:         "https://update.slpw.de/",
			DiffURL:        "https://update.slpw.de/",
			Dir:            "siteMgr-server/",
			CmdName:        "siteMgr-server", // app name
		}

		if updater != nil {
			go updater.BackgroundRun()
		}
	}
}

func setupPersistence() {
	pp, err := bolt.InitBolt("siteMgr.db")
	if err != nil {
		idl.Emerg(err)
	}
	kv.Init(pp)
}

func setupHttp() {
	tpl := template.NewTpl()
	tpl.SetPrep(func(f html.FuncMap) templice.Func {
		return func(templ *html.Template) *html.Template {
			return templ.Funcs(f)
		}
	}(html.FuncMap{
		"serverVersion": func() string {
			return VERSION
		},
		"clientInfo": getClientInfo,
	}))

	if *debug {
		tpl.Dev()
	}
	r, err := template.Renderer(tpl)
	if err != nil {
		idl.Emerg(err)
	}

	e := echo.New()
	e.SetRenderer(r)
	if false {
		e.Use(iDechoLog.New())
	}

	assetHandler := http.FileServer(rice.MustFindBox("assets").HTTPBox())
	e.Get("/assets/*", func(c *echo.Context) error {
		http.StripPrefix("/assets/", assetHandler).ServeHTTP(c.Response().Writer(), c.Request())
		return nil
	})

	route.Init(e, getClientAddress)
	e.HTTP2(true)

	idl.Info("HTTP on", *httpListener)
	go e.Run(*httpListener)
}

func runClientListener() {
	idl.Info("Listener on", *clientListener)
	ln, err := net.Listen("tcp", *clientListener)
	if err != nil {
		idl.Emerg(err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			idl.Err(err)
		}
		go handleConnection(conn)
	}
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
		inf := uim[name]
		inf.Identicon = html.HTML(icn.String())
		uim[name] = inf
		mu.Unlock()
	}
}

func setClientMsgVersion(name, ver string) {
	mu.Lock()
	defer mu.Unlock()

	inf := uim[name]
	inf.MsgVersion = ver
	uim[name] = inf
}

func setClientInfo(name, vendor, client, variant, address string) {
	mu.Lock()
	defer mu.Unlock()

	inf := uim[name]
	inf.Vendor = vendor
	inf.Client = client
	inf.Variant = variant
	inf.Address = address
	uim[name] = inf
}

func getClientInfo(user string) clientInfo {
	mu.RLock()
	defer mu.RUnlock()
	return uim[user]
}

func getClientAddress(user string) string {
	return getClientInfo(user).Address
}

type clientInfo struct {
	Identicon  html.HTML
	MsgVersion string
	Vendor     string
	Client     string
	Variant    string
	Address    string
}
