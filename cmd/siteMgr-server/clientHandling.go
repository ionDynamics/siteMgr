package main

import (
	"encoding/json"
	"net"
	"time"

	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/protocol"
	"go.iondynamics.net/siteMgr/protocol/msgType"
	"go.iondynamics.net/siteMgr/srv"
	reg "go.iondynamics.net/siteMgr/srv/registry"
	"go.iondynamics.net/siteMgr/srv/session"
)

func handleConnection(c net.Conn) {
	defer c.Close()
	abort := false

	send := make(chan protocol.Message, 1)

	go sendLoop(send, c, &abort)
	usr := recvLoop(send, c, &abort)

	abort = true
	close(send)
	cleanupConnectionInfo(usr)
	idl.Debug("connection handler shutdown")
}

func recvLoop(send chan protocol.Message, c net.Conn, abort *bool) *srv.User {
	idl.Debug("Starting recvLoop")
	msg := &protocol.Message{}
	usr := srv.NewUser()
	dec := json.NewDecoder(c)
	authFails := 0
	errCount := 0
	authed := false

recvLoop:
	for dec.More() && !*abort {
		if errCount > 10 {
			fail, err := encoder.Do("too many faulty messages")
			if err == nil {
				send <- fail
				<-time.After(3 * time.Second)
			}
			break recvLoop
		}

		err := dec.Decode(msg)
		if err != nil {
			idl.Warn(err)
			errCount++
			continue recvLoop
		}

		if !siteMgr.Constraint(protocolConstraint, msg.Version) {
			incomp, err := encoder.Do(msgType.INCOMPATIBLE)
			if err == nil {
				send <- incomp
				<-time.After(3 * time.Second)
			}
			break recvLoop
		}

		switch msg.Type {
		case msgType.LOGOUT:
			idl.Debug("client logout: ", usr.Name)
			break recvLoop

		case msgType.SITEMGR_USER:
			cleanupConnectionInfo(usr)
			usr = srv.NewUser()
			if !auth(send, msg, usr) {
				authFails++
				authed = false
			} else {
				session.Sync(usr)
				authed = true
			}

		case msgType.ENC_CREDENTIALS:
			cred := &siteMgr.Credentials{}
			err := json.Unmarshal(msg.Body, cred)
			if err != nil {
				errCount++
				idl.Warn(err)
				continue recvLoop
			}
			err = usr.SetCredentials(*cred)
			if err != nil {
				idl.Err(err)
				continue recvLoop
			}

		case msgType.CONNECTION_INFO:
			if !authed {
				errCount++
				continue recvLoop
			}

			conInfo := &siteMgr.ConnectionInfo{}
			err := json.Unmarshal(msg.Body, conInfo)
			if err != nil {
				errCount++
				idl.Warn(err)
				continue recvLoop
			}
			if conInfo.ProtocolVersion != msg.Version {
				errCount++
				idl.Debug("conInfo.ProtocolVersion != msg.Version")
				continue recvLoop
			}
			setConnectionInfo(usr, conInfo, c, send)

		default:
			errCount++
			idl.Warn("msg not handled", msg)
			say, err := encoder.Do("message not handled")
			if err == nil {
				send <- say
			}
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

func sendLoop(ch chan protocol.Message, c net.Conn, abort *bool) {
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

func auth(send chan protocol.Message, msg *protocol.Message, usr *srv.User) bool {
	var authed bool

	err := json.Unmarshal(msg.Body, usr)
	if err != nil {
		idl.Err("unmarshal body:", err)
		idl.Debug("ouch:", string(msg.Body))
		return false
	}

	var loginresult protocol.Message
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

func broadcast(msg protocol.Message) {
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
