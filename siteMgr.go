package siteMgr //import "go.iondynamics.net/siteMgr"

import (
	"fmt"
	"sync"

	semver "github.com/hashicorp/go-version"
	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr/msgType"
)

type User struct {
	Name        string
	Password    string
	m           sync.RWMutex
	Sites       map[string]Site
	Credentials map[string]Credentials
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

type Credentials struct {
	Name     string
	Login    string
	Email    string
	Password string
	Version  string
}

func NewUser() *User {
	return &User{
		Sites:       make(map[string]Site),
		Credentials: make(map[string]Credentials),
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

func (u *User) SetCredentials(c Credentials) error {
	u.m.Lock()
	defer u.m.Unlock()

	u.Credentials[c.Name] = c
	return u.Upsert()
}

func (u *User) GetCredentials(name string) Credentials {
	u.m.RLock()
	defer u.m.RUnlock()

	return u.Credentials[name]
}

func (u *User) GetAllCredentials() []Credentials {
	u.m.RLock()
	defer u.m.RUnlock()

	credentials := []Credentials{}
	for _, credential := range u.Credentials {
		credentials = append(credentials, credential)
	}

	return credentials
}

func (u *User) DelCredentials(name string) error {
	u.m.Lock()
	defer u.m.Unlock()

	delete(u.Credentials, name)
	return u.Upsert()
}

func AtLeast(constraint, version string) bool {
	c, err := semver.NewConstraint(">= " + constraint)
	if err != nil {
		idl.Err(err)
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		idl.Debug(err)
		return false
	}

	return c.Check(v)
}
