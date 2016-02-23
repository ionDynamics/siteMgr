package route //import "go.iondynamics.net/siteMgr/srv/route"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	idl "go.iondynamics.net/iDlogger"

	"go.iondynamics.net/siteMgr"
	"go.iondynamics.net/siteMgr/encoder"
	"go.iondynamics.net/siteMgr/srv/registry"
	"go.iondynamics.net/siteMgr/srv/session"
)

func Init(e *echo.Echo, fn func(string) string) {
	e.Get("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, "/site/list")
	})

	e.Get("/register", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "registerGet.tpl", nil)
	})

	e.Post("/register", func(c *echo.Context) error {
		if c.Form("login-pass") != c.Form("login-pass2") {
			return fmt.Errorf("%s", "passwords don't match")
		}
		usr := siteMgr.NewUser()
		usr.Name = c.Form("login-name")
		usr.Password = c.Form("login-pass")

		err := usr.Register()
		if err != nil {
			idl.Debug(err)
			return err
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	})

	e.Get("/login", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "loginGet.tpl", nil)
	})

	e.Post("/login", func(c *echo.Context) error {
		usr := siteMgr.NewUser()
		usr.Name = c.Form("login-name")
		usr.Password = c.Form("login-pass")

		if err := usr.Login(); err != nil {
			idl.Debug(err)
			return err
		}

		expiration := time.Now().Add(10 * time.Hour)
		cookie := http.Cookie{Name: "token", Value: session.Start(usr), Expires: expiration}
		http.SetCookie(c.Response(), &cookie)

		return c.Redirect(http.StatusFound, "/site/list")
	})

	e.Get("/site/list", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		if registry.Get(usr.Name) == nil {
			return c.Render(http.StatusOK, "clientGet.tpl", usr)
		}

		return c.Render(http.StatusOK, "siteListGet.tpl", usr)
	})

	e.Post("/site/send", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		ch := registry.Get(usr.Name)
		if ch != nil {
			idl.Debug(ch)
			site := usr.GetSite(c.Form("site-name"))
			idl.Debug(site)

			idl.Debug("sending site to client: ", usr.Name)
			msg, err := encoder.Do(site)
			if err != nil {
				return err
			}
			ch <- msg
		} else {
			idl.Debug("nil channel")
		}
		return c.Redirect(http.StatusFound, "/site/list")
	})

	e.Post("/site/set", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		site := siteMgr.Site{
			Name:     c.Form("site-name"),
			Version:  c.Form("site-version"),
			Template: c.Form("site-template"),
			Login:    c.Form("site-login"),
			Email:    c.Form("site-email"),
		}
		usr.SetSite(site)
		return c.Redirect(http.StatusFound, "/site/list")
	})

	e.Post("/site/del", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		err := usr.DelSite(c.Form("site-name"))
		if err != nil {
			idl.Debug(err)
			return err
		}

		return c.Redirect(http.StatusFound, "/site/list")
	})

	e.Get("/backup/mgr", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		return c.Render(http.StatusOK, "backupMgrGet.tpl", usr)
	})

	e.Get("/backup/get", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+usr.Name+"_"+time.Now().Format("06_01_02_15_04")+".json\"")
		c.Response().Header().Set("Content-type", "application/json")
		enc := json.NewEncoder(c.Response())
		err := enc.Encode(usr.GetSites())
		if err != nil {
			return err
		}
		c.Response().Flush()

		return nil
	})

	e.Post("/backup/recover", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		file, _, err := c.Request().FormFile("recover-json")
		if err != nil {
			return err
		}

		sites := []siteMgr.Site{}
		dec := json.NewDecoder(file)
		err = dec.Decode(&sites)
		if err != nil {
			return err
		}

		idl.Debug(sites)

		for _, site := range sites {
			err = usr.SetSite(site)
			if err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusFound, "/backup/mgr")
	})

	e.Get("/logout", func(c *echo.Context) error {
		usr := getUser(c)
		if usr == nil {
			return c.Redirect(http.StatusFound, "/login")
		}

		token, err := c.Request().Cookie("token")
		if err != nil {
			idl.Debug(err)
			return nil
		}

		session.Del(token.Value)

		expiration := time.Now().Add(-48 * time.Hour)
		cookie := http.Cookie{Name: "token", Value: "invalid", Expires: expiration}
		http.SetCookie(c.Response(), &cookie)
		return c.Redirect(http.StatusFound, "/login")
	})

	//Session Middleware
	e.Use(func(c *echo.Context) error {
		token, err := c.Request().Cookie("token")
		if err != nil {
			idl.Debug(err)
			return nil
		}
		c.Set("usr", session.Get(token.Value))
		return nil
	})
}

func getUser(c *echo.Context) *siteMgr.User {
	usr := c.Get("usr")
	if usr == nil {
		return nil
	}
	switch usr.(type) {
	case *siteMgr.User:
		return usr.(*siteMgr.User)
	default:
		idl.Err("context.Get(\"usr\") is not a *siteMgr.User")
	}
	return nil
}
