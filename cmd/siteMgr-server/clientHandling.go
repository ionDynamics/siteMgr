package main

import (
	"encoding/json"
	"net"
	"time"

	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/msgType"
	reg "go.iondynamics.net/siteMgr/srv/registry"
)

func handleConnection(c net.Conn) {
	defer c.Close()
	abort := false

	send := make(chan siteMgr.Message, 1)

	go sendLoop(send, c, &abort)
	usr := recvLoop(send, c, &abort)

	abort = true
	close(send)
	cleanupClientInfo(usr)
	idl.Debug("connection handler shutdown")
}

func recvLoop(send chan siteMgr.Message, c net.Conn, abort *bool) *siteMgr.User {
	idl.Debug("Starting recvLoop")
	msg := &siteMgr.Message{}
	usr := siteMgr.NewUser()
	dec := json.NewDecoder(c)
	authFails := 0

recvLoop:
	for dec.More() && !*abort {
		err := dec.Decode(msg)
		if err != nil {
			idl.Warn(err)
			continue recvLoop
		}

		switch msg.Type {
		case msgType.LOGOUT:
			idl.Debug("client logout: ", usr.Name)
			break recvLoop

		case msgType.SITEMGR_USER:
			cleanupClientInfo(usr)
			usr = siteMgr.NewUser()
			if !auth(send, msg, usr) {
				authFails++
			} else {
				saveClientInfo(msg, c, send)
			}

		default:
			idl.Warn("msg not handled", msg)
		}

		if authFails > 2 {
			fail, err := encoder.Do("too many failed login attempts")
			if err == nil {
				send <- fail
				<-time.After(3 * time.Second)
			}
			break recvLoop
		}

	}
	*abort = true
	idl.Debug("recvLoop exit")
	return usr
}

func sendLoop(ch chan siteMgr.Message, c net.Conn, abort *bool) {
	idl.Debug("Starting sendLoop")
	enc := json.NewEncoder(c)
	for msg := range ch {
		if !*abort {
			idl.Debug("Sending message", msg)
			err := enc.Encode(msg)
			if err != nil {
				idl.Err(err)
				break
			}
		}
	}
	*abort = true
}

func auth(send chan siteMgr.Message, msg *siteMgr.Message, usr *siteMgr.User) bool {
	var authed bool

	err := json.Unmarshal(msg.Body, usr)
	if err != nil {
		idl.Err("unmarshal body:", err)
		idl.Debug("ouch:", string(msg.Body))
		return false
	}

	var loginresult siteMgr.Message
	if usr.Login() == nil {
		idl.Debug("Login success")
		loginresult, err = encoder.Do(msgType.LOGIN_SUCCESS)
		authed = true
	} else {
		idl.Debug("Login error")
		loginresult, err = encoder.Do(msgType.LOGIN_ERROR)
	}

	if err != nil {
		idl.Err(err)
		return false
	}

	send <- loginresult

	if authed {
		idl.Debug("User Client logged in: ", usr.Name)
	}

	return authed
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
