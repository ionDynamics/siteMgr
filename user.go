package siteMgr //import "go.iondynamics.net/siteMgr"

import (
	"sort"
	"sync"

	"encoding/json"
)

type User struct {
	Name        string
	Password    string
	m           sync.RWMutex
	sites       map[string]Site
	credentials map[string]Credentials
}

type jsonUser struct {
	Name        string
	Password    string
	Sites       map[string]Site
	Credentials map[string]Credentials
}

func NewUser() *User {
	return &User{
		sites:       make(map[string]Site),
		credentials: make(map[string]Credentials),
	}
}

func (u *User) SetSite(s Site) {
	u.m.Lock()
	defer u.m.Unlock()

	if u.sites == nil {
		u.sites = make(map[string]Site)
	}

	u.sites[s.Name] = s
}

func (u *User) GetSite(s string) Site {
	u.m.RLock()
	defer u.m.RUnlock()

	return u.sites[s]
}

func (u *User) GetSites() []Site {
	u.m.RLock()
	defer u.m.RUnlock()

	sites := []Site{}
	for _, site := range u.sites {
		sites = append(sites, site)
	}

	sort.Sort(SiteSort(sites))

	return sites
}

func (u *User) DelSite(s string) {
	u.m.Lock()
	defer u.m.Unlock()

	delete(u.sites, s)
}

func (u *User) SetCredentials(c Credentials) {
	u.m.Lock()
	defer u.m.Unlock()

	if u.credentials == nil {
		u.credentials = make(map[string]Credentials)
	}

	u.credentials[c.Name] = c
}

func (u *User) GetCredentials(name string) Credentials {
	u.m.RLock()
	defer u.m.RUnlock()

	return u.credentials[name]
}

func (u *User) GetAllCredentials() []Credentials {
	u.m.RLock()
	defer u.m.RUnlock()

	credentials := []Credentials{}
	for _, credential := range u.credentials {
		credentials = append(credentials, credential)
	}

	sort.Sort(CredentialsSort(credentials))

	return credentials
}

func (u *User) DelCredentials(name string) {
	u.m.Lock()
	defer u.m.Unlock()

	delete(u.credentials, name)
}

func (u *User) MarshalJSON() ([]byte, error) {
	jso := &jsonUser{}
	jso.Name = u.Name
	jso.Password = u.Password
	jso.Sites = u.sites
	jso.Credentials = u.credentials

	return json.Marshal(jso)
}

func (u *User) UnmarshalJSON(data []byte) error {
	u.m.Lock()
	defer u.m.Unlock()

	jso := &jsonUser{}
	err := json.Unmarshal(data, jso)

	if err == nil {
		u.Name = jso.Name
		u.Password = jso.Password
		u.sites = jso.Sites
		u.credentials = jso.Credentials
	}

	return err
}
