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
	"github.com/labstack/echo"
	"go.iondynamics.net/go-selfupdate"
	"go.iondynamics.net/iDechoLog"
	idl "go.iondynamics.net/iDlogger"
	kv "go.iondynamics.net/kvStore"
	"go.iondynamics.net/kvStore/backend/bolt"
	"go.iondynamics.net/templice"

	"go.iondynamics.net/siteMgr/encoder"
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

	VERSION = "0.5.0"

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
