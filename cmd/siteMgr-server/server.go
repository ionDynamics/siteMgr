package main

import (
	"encoding/json"
	"flag"
	html "html/template"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/GeorgeMac/idicon/icon"
	"github.com/labstack/echo"
	"go.iondynamics.net/go-selfupdate"
	"go.iondynamics.net/iDechoLog"
	idl "go.iondynamics.net/iDlogger"
	kv "go.iondynamics.net/kvStore"
	"go.iondynamics.net/kvStore/backend/bolt"
	"go.iondynamics.net/templice"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/msgType"
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
	cim map[string]clientInfo = make(map[string]clientInfo)

	key = []byte{0x42, 0x89, 0x10, 0x01, 0x07, 0xAC, 0xDC, 0x77, 0x70, 0x07, 0x66, 0x6B, 0xCD, 0xFF, 0x13, 0xCC}

	VERSION = "0.4.0"

	updater = &selfupdate.Updater{
		CurrentVersion: VERSION,
		ApiURL:         "https://update.slpw.de/",
		BinURL:         "https://update.slpw.de/",
		DiffURL:        "https://update.slpw.de/",
		Dir:            "siteMgr-server/",
		CmdName:        "siteMgr-server", // app name
	}

	lcs string
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	VERSION = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(VERSION), "'"), "'")
	idl.Info("Version: ", VERSION)
	idl.Info("Starting...")

	encoder.Init(VERSION)
	update()
	setupPersistence()
	setupHttp()
	runClientListener()
}

func update() {
	if *autoupdate {
		go func() {
			for {
				lcs = latestClientVersion()

				reload, _ := updater.Update()
				idl.Debug("reload ", reload)
				if reload {
					msg, err := encoder.Do("Server is restarting")
					if err == nil {
						broadcast(msg)
					} else {
						idl.Err(err)
					}
					<-time.After(10 * time.Second)
					os.Exit(1)
				}
				<-time.After(12 * time.Hour)
			}
		}()
	}
}

func latestClientVersion() string {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://update.slpw.de/siteMgr-server/windows-amd64.json")
	if err != nil {
		idl.Err("latestClientVersion", err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	m := make(map[string]string)
	err = dec.Decode(m)
	if err != nil {
		idl.Err("latestClientVersion", err)
	}
	vers, ok := m["Version"]
	if !ok {
		idl.Err("latestClientVersion", "no version")
	}
	return vers
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
		inf := cim[name]
		inf.Identicon = html.HTML(icn.String())
		cim[name] = inf
		mu.Unlock()
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
			msg.Type = msgType.LOGIN_SUCCESS
		} else {
			msg.Type = msgType.LOGIN_ERROR
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
			if gomsg.Type == msgType.LOGOUT {
				idl.Debug("client logout: ", usr.Name)
				break
			}
		}
		closed <- true
	}()

	setIdenticon(usr.Name, hash)

	ch := make(chan siteMgr.Message, 1)
	idl.Debug("registering channel ", usr.Name, ch)
	reg.Set(usr.Name, ch)

loop:
	for {
		idl.Debug("ready for sending messages")
		var err error
		var send siteMgr.Message
		select {
		case send = <-ch:
			idl.Debug("channel sent msg")
		case <-time.After(3 * time.Minute):
			send, err = encoder.Do(msgType.HEARTBEAT)
			if err != nil {
				idl.Err(err)
				continue loop
			}
		case <-closed:
			break loop
		}
		idl.Debug("Sending message", send)
		err = enc.Encode(send)
		if err != nil {
			idl.Warn(err)
			break
		}
	}

	delClientInfo(usr.Name)
	reg.Del(usr.Name)
	idl.Debug("connection handler shutdown")
}

func broadcast(msg siteMgr.Message) {
	mu.RLock()
	defer mu.RUnlock()

	for name, _ := range cim {
		ch := reg.Get(name)
		if ch == nil {
			continue
		}
		ch <- msg
	}
}
