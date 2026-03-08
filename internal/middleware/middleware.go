package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/thomiceli/opengist/internal/config"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/session"
)

func Session(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get(c, "session")
		c.Set("session", sess)
		return next(c)
	}
}

func LoadUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess := c.Get("session").(*session.Session)
		if userId, ok := sess.Values["user"].(uint); ok {
			user, err := db.GetUserById(userId)
			if err == nil {
				c.Set("user", user)
			}
		}
		return next(c)
	}
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return c.Redirect(302, "/login")
		}
		return next(c)
	}
}

func SetData(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("config", config.C)
		user := c.Get("user")
		if user != nil {
			c.Set("currentUser", user.(*db.User))
		}
		return next(c)
	}
}
