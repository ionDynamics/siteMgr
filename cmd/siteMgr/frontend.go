package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh/terminal"
)

func getFullname(s *bufio.Scanner) string {
	var fullname string
	if !*noEnv {
		fullname = os.Getenv(nameEnv)
	}

	if fullname == "" {
		say("Enter Full Name:")

		var err error
		fullname, err = getString(s)
		if err != nil {
			say(err)
			os.Exit(1)
		}

		if !*noEnv {
			err = os.Setenv(nameEnv, fullname)
			if err != nil {
				say(err)
				os.Exit(1)
			}
		}
	} else {
		say("Generating passwords for", fullname)
	}

	return strings.TrimSpace(fullname)
}

func getPassword(s *bufio.Scanner) []byte {
	var err error
	var mpw []byte

	if *noTerminal {
		var str string
		str, err = getString(s)
		mpw = []byte(str)
	} else {
		mpw, err = terminal.ReadPassword(int(os.Stdin.Fd()))
	}

	if err != nil {
		say(err)
		os.Exit(1)
	}
	return mpw
}

func getMasterpassword(s *bufio.Scanner) []byte {
	say("Enter Master Password:")
	return getPassword(s)
}

func getLoginName(s *bufio.Scanner) string {
	var login string
	var err error
	if !*noEnv {
		login = os.Getenv(loginEnv)
	}

	if login == "" {
		say("Enter Login Name:")
		login, err = getString(s)
		if err != nil {
			say(err)
			os.Exit(1)
		}
	}

	if !*noEnv {
		err := os.Setenv(loginEnv, login)
		if err != nil {
			say(err)
			os.Exit(1)
		}
	}
	return login
}

func getLoginPassword(s *bufio.Scanner) string {
	var pass string
	if !*noEnv {
		pass = os.Getenv(passEnv)
	}

	if pass == "" {
		say("Enter Login Password:")
		pass = string(getPassword(s))
	}

	if !*noEnv {
		err := os.Setenv(passEnv, pass)
		if err != nil {
			say(err)
			os.Exit(1)
		}
	}

	return pass
}

func returnPw(pwch chan string) {
	var pwd string
	select {
	case pw := <-pwch:
		pwd = pw
	case <-time.After(750 * time.Millisecond):
		say("Generating...")
		pwd = <-pwch
	}

	if *noTerminal {
		fmt.Print(pwd)
		return
	}

	clip(pwd)
}

func clip(str string) {
	before, err := clipboard.ReadAll()
	clipboard.WriteAll(str)
	say("\aCopied to clipboard! ")
	time.Sleep(5 * time.Second)
	say("Restoring clipboard in 5 seconds...")
	time.Sleep(5 * time.Second)
	if err != nil {
		clipboard.WriteAll("")
	} else {
		clipboard.WriteAll(before)
	}
	say("\aClipboard restored")
}

func getString(s *bufio.Scanner) (string, error) {
	for s.Scan() {
		return s.Text(), nil
		break
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func say(s ...interface{}) {
	if *noTerminal {
		return
	}

	fmt.Println(s...)
}
