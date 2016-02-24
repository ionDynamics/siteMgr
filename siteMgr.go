package siteMgr //import "go.iondynamics.net/siteMgr"

import (
	"fmt"
	"sync"

	semver "github.com/hashicorp/go-version"
	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr/msgType"
)

type User struct {
	Name     string
	Password string
	m        sync.RWMutex
	Sites    map[string]Site
}

type Site struct {
	Name     string
	Login    string
	Version  string
	Template string
	Email    string
}

type Message struct {
	Type    msgType.Code
	Body    []byte
	Version string
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

func (u *User) GetSites() []Site {
	u.m.RLock()
	defer u.m.RUnlock()

	sites := []Site{}
	for _, site := range u.Sites {
		sites = append(sites, site)
	}

	return sites
}

func (u *User) DelSite(s string) error {
	u.m.Lock()
	defer u.m.Unlock()

	delete(u.Sites, s)
	return u.Upsert()
}

func AtLeast(str string, msg *Message) bool {
	constraint, err := semver.NewConstraint(">= " + str)
	if err != nil {
		idl.Err(err)
		return false
	}

	v, err := semver.NewVersion(msg.Version)
	if err != nil {
		idl.Debug(err)
		return false
	}

	return constraint.Check(v)
}
