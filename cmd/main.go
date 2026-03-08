package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/thomiceli/opengist/internal/config"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/handler"
	custommw "github.com/thomiceli/opengist/internal/middleware"
	"github.com/thomiceli/opengist/internal/session"
	"github.com/thomiceli/opengist/internal/template"
)

//go:embed web/templates/*.html
var templatesFS embed.FS

//go:embed web/static/*.svg
var staticFS embed.FS

func main() {
	if err := config.Init(); err != nil {
		log.Fatal(err)
	}

	if err := db.Setup(config.C.DBPath); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	session.Init()

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(custommw.Session)
	e.Use(custommw.LoadUser)
	e.Use(custommw.SetData)

	tmpl := template.NewRenderer()
	templateContent, _ := fs.Sub(templatesFS, "web/templates")
	tmpl.AddFromFS(templateContent, "*.html")
	e.Renderer = tmpl

	// Static files
	staticContent, _ := fs.Sub(staticFS, "web/static")
	e.StaticFS("/static", staticContent)

	e.GET("/login", handler.Login)
	e.POST("/login", handler.ProcessLogin)
	e.GET("/logout", handler.Logout)
	e.GET("/settings/password", custommw.RequireAuth(handler.ChangePassword))
	e.POST("/settings/password", custommw.RequireAuth(handler.ProcessChangePassword))

	e.GET("/all", custommw.RequireAuth(handler.AllGists))
	e.GET("/new", custommw.RequireAuth(handler.NewGist))
	e.POST("/new", custommw.RequireAuth(handler.CreateGist))

	e.GET("/:title", handler.ViewGist)
	e.GET("/:title/detail", custommw.RequireAuth(handler.GistDetail))
	e.GET("/:title/edit", custommw.RequireAuth(handler.EditGist))
	e.POST("/:title/edit", custommw.RequireAuth(handler.UpdateGist))
	e.POST("/:title/delete", custommw.RequireAuth(handler.DeleteGist))
	e.GET("/:title/revisions", custommw.RequireAuth(handler.Revisions))
	e.GET("/:title/revisions/:version", custommw.RequireAuth(handler.ViewRevision))

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/all")
	})

	log.Printf("CloseGist starting on %s:%s", config.C.HttpHost, config.C.HttpPort)
	e.Start(config.C.HttpHost + ":" + config.C.HttpPort)
}
