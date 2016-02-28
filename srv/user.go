package srv //import "go.iondynamics.net/siteMgr/srv"

import (
	"fmt"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/srv/persistence"
)

type User struct {
	*siteMgr.User
}

func NewUser() *User {
	return &User{siteMgr.NewUser()}
}

func (u *User) Update() error {
	u2, err := persistence.ReadUser(u.Name, u.Password)
	if err != nil {
		return err
	}
	*u = User{u2}
	return nil
}

func (u *User) Register() error {
	exists, err := persistence.UserExists(u.Name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s", "already exists")
	}
	return u.Upsert()
}

func (u *User) Upsert() error {
	return persistence.UpsertUser(u.User)
}

func (u *User) Login() error {
	u2, err := persistence.ReadUser(u.Name, u.Password)
	if err != nil {
		return err
	}
	if u2.Password == u.Password && u2.Name == u.Name {
		*u = User{u2}
		return nil
	}
	return fmt.Errorf("%s", "invalid login")

}

func (u *User) SetSite(s siteMgr.Site) error {
	u.User.SetSite(s)
	return u.Upsert()
}

func (u *User) DelSite(s string) error {
	u.User.DelSite(s)
	return u.Upsert()
}

func (u *User) SetCredentials(c siteMgr.Credentials) error {
	u.User.SetCredentials(c)
	return u.Upsert()
}

func (u *User) DelCredentials(name string) error {
	u.User.DelCredentials(name)
	return u.Upsert()
}
