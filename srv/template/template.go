package template //import "go.iondynamics.net/siteMgr/srv/template"

import (
	"github.com/GeertJohan/go.rice"
	"go.iondynamics.net/templice"
	"go.iondynamics.net/templiceEchoRenderer"
)

func New() (*templiceEchoRenderer.Renderer, error) {
	tpl := templice.New(rice.MustFindBox("files"))
	err := tpl.Load()
	return templiceEchoRenderer.New(tpl), err
}

func NewTpl() *templice.Template {
	tpl := templice.New(rice.MustFindBox("files"))
	return tpl
}

func Dev() (*templiceEchoRenderer.Renderer, error) {
	tpl := templice.New(rice.MustFindBox("files"))
	err := tpl.Dev().Load()
	return templiceEchoRenderer.New(tpl), err
}

func Renderer(tpl *templice.Template) (*templiceEchoRenderer.Renderer, error) {
	err := tpl.Load()
	return templiceEchoRenderer.New(tpl), err
}
