package siteMgr //import "go.iondynamics.net/siteMgr"

import (
	semver "github.com/hashicorp/go-version"
	idl "go.iondynamics.net/iDlogger"
)

type Site struct {
	Name     string
	Login    string
	Version  string
	Template string
	Email    string
}

type SiteSort []Site

func (a SiteSort) Len() int           { return len(a) }
func (a SiteSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SiteSort) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Credentials struct {
	Name     string
	Login    string
	Email    string
	Password string
	Version  string
}

type CredentialsSort []Credentials

func (a CredentialsSort) Len() int           { return len(a) }
func (a CredentialsSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CredentialsSort) Less(i, j int) bool { return a[i].Name < a[j].Name }

type ConnectionInfo struct {
	ProtocolVersion    string
	ProtocolConstraint string
	RemoteAddress      string
	ClientVendor       string
	ClientName         string
	ClientVariant      string
	ClientVersion      string
	IdenticonHash      []byte
}

func AtLeast(constraint, version string) bool {
	return Constraint(">= "+constraint, version)
}

func Constraint(constraint, version string) bool {
	c, err := semver.NewConstraint(constraint)
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
