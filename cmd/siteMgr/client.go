package main

import (
	"bufio"
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"flag"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"go.iondynamics.net/go-selfupdate"
	"go.iondynamics.net/iDhelper/crypto"
	idl "go.iondynamics.net/iDlogger"
	"go.iondynamics.net/reProc"
	"go.iondynamics.net/statelessPassword"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/protocol"
	"go.iondynamics.net/siteMgr/protocol/msgType"
)

var (
	nameEnv  = "ID_SLPW_FULLNAME"
	loginEnv = "ID_SITEMGR_USER"
	passEnv  = "ID_SITEMGR_PASS"

	noEnv       = flag.Bool("no-env", false, "Set this to true if no environment variables should be used")
	noTerminal  = flag.Bool("no-terminal", false, "Set this to true if your standard input isn't a terminal")
	debug       = flag.Bool("debug", false, "Enable debug logging")
	server      = flag.String("server", "mgr.slpw.de:9210", "siteMgr-server host:port")
	autoupdate  = flag.Bool("autoupdate", true, "Enable or disable automatic updates")
	insecure    = flag.Bool("insecure", false, "Allow insecure connections")
	noTelemetry = flag.Bool("noTelemetry", false, "Disable sending metrics to server")

	VERSION            = "0.7.0"
	protocolConstraint = ">= 0.7.0"

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
	if !*debug {
		defer func() {
			if r := recover(); r != nil {
				idl.Emerg(r)
			}
		}()
	}

	VERSION = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(VERSION), "'"), "'")

	idl.Debug(VERSION)

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
			conn, err = tls.Dial("tcp", *server, &tls.Config{InsecureSkipVerify: *insecure})
			if err != nil {
				idl.Debug(err)
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
		msg := &protocol.Message{}

		for msg.Type != msgType.LOGIN_SUCCESS {
			usr.Name = getLoginName(s)
			usr.Password = getLoginPassword(s)

			//usr.SetSite("identicon-hash") = siteMgr.Site{Name: string(hash.Sum(nil))}
			//usr.Sites["client"] = siteMgr.Site{Name: "ionDynamics", Login: "siteMgr", Version: "CLI"}

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

		sites <- &siteMgr.Site{
			Name:     usr.Name,
			Template: statelessPassword.Printable16Templates[0],
		}
		pw := <-pwd

		hash := sha512.New()
		for i := 0; i < 10000; i++ {
			hash.Write([]byte(pw))
			hash.Write(hash.Sum(nil))
		}

		ci := &siteMgr.ConnectionInfo{
			ProtocolVersion:    protocol.Version,
			ProtocolConstraint: protocolConstraint,
			ClientVendor:       "ionDynamics",
			ClientName:         "siteMgr",
			ClientVariant:      "CLI",
			ClientVersion:      VERSION,
			IdenticonHash:      hash.Sum(nil),
		}

		cimsg, err := encoder.Do(ci)
		if err == nil {
			err = enc.Encode(cimsg)
			if err != nil {
				idl.Err(err)
			}
		} else {
			idl.Err(err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for {
				<-c
				gomsg, err := encoder.Do(msgType.LOGOUT)
				if err == nil {
					idl.Debug(enc.Encode(gomsg))
				} else {
					idl.Debug(err)
				}
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

			case msgType.INCOMPATIBLE:
				say("This client is incompatible with the conntected server")
				if updater != nil {
					updater.Update()
					say("Restart required")
				}

			case msgType.NOTICE:
				say("Notice from server:", string(msg.Body))

			case msgType.CLIPCONTENT:
				clip(string(msg.Body))

			case msgType.DEC_CREDENTIALS:
				sites <- &siteMgr.Site{
					Name:     "Internal Crypto Key",
					Version:  "version1",
					Template: statelessPassword.Printable32Templates[0] + statelessPassword.Printable32Templates[0],
				}

				cred := &siteMgr.Credentials{}
				err := json.Unmarshal(msg.Body, cred)
				pw := <-pwd
				if err != nil {
					idl.Err(err)
					continue
				}

				cred.Password = crypto.Encrypt(pw, cred.Password)
				cred.Version = "version1"

				ret, err := encoder.Do(cred)
				if err != nil {
					idl.Err(err)
					continue
				}

				err = enc.Encode(ret)
				if err != nil {
					idl.Err(err)
				}

			case msgType.ENC_CREDENTIALS:
				sites <- &siteMgr.Site{
					Name:     "Internal Crypto Key",
					Version:  "version1",
					Template: statelessPassword.Printable32Templates[0] + statelessPassword.Printable32Templates[0],
				}

				cred := &siteMgr.Credentials{}
				err := json.Unmarshal(msg.Body, cred)
				pw := <-pwd
				if err != nil {
					idl.Err(err)
					continue
				}

				clip(crypto.Decrypt(pw, cred.Password))
			}

		}
		say("Connection lost")
		say("Retrying...")
	}
}
