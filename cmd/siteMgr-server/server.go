package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"go.iondynamics.net/iDechoLog"
	idl "go.iondynamics.net/iDlogger"
	kv "go.iondynamics.net/kvStore"
	"go.iondynamics.net/kvStore/backend/bolt"

	"go.iondynamics.net/siteMgr"
	reg "go.iondynamics.net/siteMgr/srv/registry"
	"go.iondynamics.net/siteMgr/srv/route"
	"go.iondynamics.net/siteMgr/srv/template"
)

var (
	debug          = flag.Bool("debug", false, "Enable debug output")
	httpListener   = flag.String("http", ":9211", "Bind HTTP-Listener to ...")
	clientListener = flag.String("client", ":9210", "Bind Client-Listener to ...")
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	idl.Info("Starting...")

	pp, err := bolt.InitBolt("siteMgr.db")
	if err != nil {
		idl.Emerg(err)
	}
	kv.Init(pp)

	tpl, err := template.New()
	if *debug {
		tpl, err = template.Dev()
	}
	if err != nil {
		idl.Emerg(err)
	}

	e := echo.New()
	e.SetRenderer(tpl)
	if false {
		e.Use(iDechoLog.New())
	}

	assetHandler := http.FileServer(rice.MustFindBox("assets").HTTPBox())
	e.Get("/assets/*", func(c *echo.Context) error {
		http.StripPrefix("/assets/", assetHandler).ServeHTTP(c.Response().Writer(), c.Request())
		return nil
	})

	route.Init(e)

	idl.Info("HTTP on", *httpListener)
	go e.Run(*httpListener)

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

func handleConnection(c net.Conn) {
	defer c.Close()

	enc := json.NewEncoder(c)
	dec := json.NewDecoder(c)
	msg := &siteMgr.Message{}
	usr := siteMgr.NewUser()

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

	ch := make(chan siteMgr.Site, 1)
	idl.Debug("registering channel ", usr.Name, ch)
	reg.Set(usr.Name, ch)

loop:
	for {
		idl.Debug("ready for sending messages")
		var err error
		select {
		case s := <-ch:
			idl.Debug("channel sent side")

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
