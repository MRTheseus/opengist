package session

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/thomiceli/opengist/internal/config"
)

type Session = sessions.Session

var store *sessions.CookieStore

func Init() {
	store = sessions.NewCookieStore([]byte(config.C.SecretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
		Secure:   false,
	}
}

func Get(c echo.Context, name string) (*Session, error) {
	return store.Get(c.Request(), name)
}

func Save(c echo.Context, sess *Session) error {
	return sess.Save(c.Request(), c.Response())
}

func Delete(c echo.Context, name string) error {
	sess, err := store.Get(c.Request(), name)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}
