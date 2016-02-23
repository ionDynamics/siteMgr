package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"go.iondynamics.net/go-selfupdate"
	idl "go.iondynamics.net/iDlogger"
	"go.iondynamics.net/reProc"
	"go.iondynamics.net/statelessPassword"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/msgType"
)

var (
	nameEnv  = "ID_SLPW_FULLNAME"
	loginEnv = "ID_SITEMGR_USER"
	passEnv  = "ID_SITEMGR_PASS"

	noEnv      = flag.Bool("no-env", false, "Set this to true if no environment variables should be used")
	noTerminal = flag.Bool("no-terminal", false, "Set this to true if your standard input isn't a terminal")
	debug      = flag.Bool("debug", false, "Enable debug logging")
	server     = flag.String("server", "sitemgr.ion.onl:9210", "siteMgr-server host:port")
	autoupdate = flag.Bool("autoupdate", true, "Enable or disable automatic updates")

	VERSION = "0.4.0"

	updater = &selfupdate.Updater{
		CurrentVersion: VERSION,
		ApiURL:         "https://update.slpw.de/",
		BinURL:         "https://update.slpw.de/",
		DiffURL:        "https://update.slpw.de/",
		Dir:            "siteMgr/",
		CmdName:        "siteMgr", // app name
	}
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	VERSION = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(VERSION), "'"), "'")

	idl.Debug(VERSION)
	encoder.Init(VERSION)

	if *autoupdate == false {
		updater = nil
	}

	if updater != nil {
		say("Checking for updates...")
		reload, _ := updater.Update()

		idl.Debug("reload ", reload)
		if reload {
			say("Your Client has been updated")
			err := reProc.Reload(VERSION)
			if err != nil {
				idl.Debug(err)
			}
		} else {
			say("Your Client is up-to-date")
		}
	}

	s := bufio.NewScanner(os.Stdin)
	fullname := getFullname(s)
	bytMasterPw := getMasterpassword(s)

	sites := make(chan *siteMgr.Site, 1)
	pwd := make(chan string, 1)

	go func() {
		algo, err := statelessPassword.New([]byte(fullname), bytMasterPw, 5)
		if err != nil {
			idl.Err(err)
			os.Exit(1)
		}
		for {
			s := <-sites
			pw, err := algo.Password(s.Name, s.Version, []string{s.Template})
			if err != nil {
				idl.Err(err)
				os.Exit(1)
			}
			pwd <- pw
		}
	}()

	for {
		var conn net.Conn
		var err error
		for i := 1; i <= 5; i++ {
			conn, err = net.Dial("tcp", *server)
			if err != nil {
				dur := time.Duration(i) * 5 * time.Second
				say("Connection could not be established!")
				say("Retrying in", dur)
				time.Sleep(dur)
				continue
			}
			break
		}
		if err != nil {
			idl.Err(err)
			time.Sleep(time.Minute)
			return
		}
		defer conn.Close()

		usr := siteMgr.NewUser()
		enc := json.NewEncoder(conn)
		dec := json.NewDecoder(conn)
		msg := &siteMgr.Message{}

		for msg.Type != msgType.LOGIN_SUCCESS {
			usr.Name = getLoginName(s)
			usr.Password = getLoginPassword(s)

			hash := sha512.New()
			for i := 0; i < 10000; i++ {
				hash.Write([]byte(fullname))
				hash.Write(bytMasterPw)
				hash.Write([]byte(usr.Name))
				hash.Write([]byte(usr.Password))
				hash.Write(hash.Sum(nil))
			}
			usr.Sites["identicon-hash"] = siteMgr.Site{Name: string(hash.Sum(nil))}
			usr.Sites["client"] = siteMgr.Site{Name: "ionDynamics", Login: "siteMgr", Version: "CLI"}

			usrmsg, err := encoder.Do(usr)
			if err != nil {
				idl.Err(err)
				return
			}
			idl.Debug(string(usrmsg.Body))

			idl.Debug("Sending credentials")
			err = enc.Encode(usrmsg)
			if err != nil {
				idl.Err("encoder error:", err)
				return
			}

			idl.Debug("decoding msg ", dec.More())
			err = dec.Decode(msg)
			if err != nil {
				idl.Err("decoder error:", err)
				return
			}
		}

		say("Successfully logged in")

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for {
				<-c
				gomsg := &siteMgr.Message{}
				gomsg.Version = VERSION
				gomsg.Type = msgType.LOGOUT
				idl.Debug(enc.Encode(gomsg))
				conn.Close()
				os.Exit(0)
			}
		}()

		for dec.More() {
			err = dec.Decode(msg)
			if err != nil {
				idl.Err(err)
				continue
			}
			idl.Debug("Received message: ", msg)

			switch msg.Type {
			case msgType.SITEMGR_SITE:
				site := &siteMgr.Site{}
				err = json.Unmarshal(msg.Body, site)
				if err != nil {
					idl.Err(err)
					continue
				}

				sites <- site
				returnPw(pwd)

			case msgType.UPDATE_AVAIL:
				if updater != nil {
					updater.Update()
				}

			case msgType.NOTICE:
				say("Notice from server:", string(msg.Body))
			}
		}
		say("Connection lost")
		say("Retrying...")
	}
}
