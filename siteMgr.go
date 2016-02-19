package siteMgr //import "go.iondynamics.net/siteMgr"

import (
	"fmt"
	"sync"
)

type User struct {
	Name     string
	Password string
	//Fullname string
	Sites map[string]Site
	m     sync.RWMutex
}

type Site struct {
	Name     string
	Login    string
	Version  string
	Template string
	Email    string
	//Password string
}

type Message struct {
	Type string
	Body []byte
}

func NewUser() *User {
	return &User{
		Sites: make(map[string]Site),
	}

}

func (u *User) Update() error {
	u2, err := ReadUser(u.Name, u.Password)
	if err != nil {
		return err
	}
	*u = *u2
	return nil
}

func (u *User) Register() error {
	exists, err := UserExists(u.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s", "already exists")
	}
	return u.Upsert()
}

func (u *User) Login() error {
	u2, err := ReadUser(u.Name, u.Password)
	if err != nil {
		return err
	}
	if u2.Password == u.Password && u2.Name == u.Name {
		*u = *u2
		return nil
	}
	return fmt.Errorf("%s", "invalid login")

}

func (u *User) Upsert() error {
	return UpsertUser(u)
}

func (u *User) SetSite(s Site) error {
	u.m.Lock()
	defer u.m.Unlock()

	u.Sites[s.Name] = s
	return u.Upsert()
}

func (u *User) GetSite(s string) Site {
	u.m.RLock()
	defer u.m.RUnlock()

	return u.Sites[s]
}

func (u *User) DelSite(s string) error {
	u.m.Lock()
	defer u.m.Unlock()

	delete(u.Sites, s)
	return u.Upsert()
}
