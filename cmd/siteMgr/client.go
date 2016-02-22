package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/sanbornm/go-selfupdate/selfupdate"
	idl "go.iondynamics.net/iDlogger"
	"go.iondynamics.net/statelessPassword"

	"go.iondynamics.net/siteMgr"
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

	VERSION = "0.2.0"
)

func main() {
	flag.Parse()
	*idl.StandardLogger() = *idl.WithDebug(*debug)

	idl.Debug(VERSION)

	if *autoupdate {
		var updater = &selfupdate.Updater{
			CurrentVersion: VERSION,
			ApiURL:         "https://update.slpw.de/",
			BinURL:         "https://update.slpw.de/",
			DiffURL:        "https://update.slpw.de/",
			Dir:            "siteMgr/",
			CmdName:        "siteMgr", // app name
		}

		if updater != nil {
			go updater.BackgroundRun()
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

	conn, err := net.Dial("tcp", *server)
	if err != nil {
		idl.Err(err)
		return
	}
	defer conn.Close()

	usr := siteMgr.NewUser()
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	msg := &siteMgr.Message{}

	for msg.Type != "LOGIN:SUCCESS" {
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

		msg.Version = VERSION
		msg.Type = "siteMgr.User"
		msg.Body, err = json.Marshal(usr)
		if err != nil {
			idl.Err(err)
			return
		}
		idl.Debug(string(msg.Body))

		idl.Debug("Sending credentials")
		err = enc.Encode(msg)
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
			gomsg.Type = "LOGOUT"
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
		if msg.Type == "siteMgr.Site" {
			site := &siteMgr.Site{}
			err = json.Unmarshal(msg.Body, site)
			if err != nil {
				idl.Err(err)
				continue
			}

			sites <- site
			returnPw(pwd)
		}
	}
}
