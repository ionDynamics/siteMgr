package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	html "html/template"
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
	debug        = flag.Bool("debug", false, "Enable debug output")
	httpsListen  = flag.String("httpsBind", "localhost:9211", "Bind HTTP-Listener to ...")
	clientListen = flag.String("clientBind", ":9210", "Bind Client-Listener to ...")
	autoupdate   = flag.Bool("autoupdate", true, "Enable or disable automatic updates")
	pemCertPath  = flag.String("pemCertPath", "", "Path to pem-encoded certificate")
	pemKeyPath   = flag.String("pemKeyPath", "", "Path to pem-encoded key")

	mu  sync.RWMutex
	cim map[string]clientInfo = make(map[string]clientInfo)

	VERSION = "0.7.1"

	updater = &selfupdate.Updater{
		CurrentVersion: VERSION,
		ApiURL:         "https://update.slpw.de/",
		BinURL:         "https://update.slpw.de/",
		DiffURL:        "https://update.slpw.de/",
		Dir:            "siteMgr-server/",
		CmdName:        "siteMgr-server", // app name
	}

	cert tls.Certificate
	lcs  string
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	VERSION = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(VERSION), "'"), "'")
	idl.Info("Version: ", VERSION)

	if *pemCertPath == "" || *pemKeyPath == "" {
		idl.Emerg("Specify --pemCertPath and --pemKeyPath")
	}

	idl.Info("Starting...")

	var err error
	cert, err = tls.LoadX509KeyPair(*pemCertPath, *pemKeyPath)
	if err != nil {
		idl.Emerg(err)
	}

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
	err = dec.Decode(&m)
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

	idl.Info("HTTP on", *httpsListen)
	go e.RunTLS(*httpsListen, *pemCertPath, *pemKeyPath)
}

func runClientListener() {
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	idl.Info("Listener on", *clientListen)
	ln, err := tls.Listen("tcp", *clientListen, cfg)
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
