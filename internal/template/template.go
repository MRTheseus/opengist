package template

import (
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thomiceli/opengist/internal/db"
)

type Template struct {
	templates *template.Template
}

func NewRenderer() *Template {
	tmpl := template.New("").Funcs(template.FuncMap{
		"formatTime": func(unix int64) string {
			return time.Unix(unix, 0).Local().Format("2006-01-02 15:04:05")
		},
		"formatDate": func(unix int64) string {
			return time.Unix(unix, 0).Local().Format("2006-01-02")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"toInt": func(v interface{}) int {
			switch val := v.(type) {
			case int:
				return val
			case int64:
				return int(val)
			case int32:
				return int(val)
			case uint:
				return int(val)
			case db.Visibility:
				return int(val)
			default:
				return 0
			}
		},
	})

	return &Template{templates: tmpl}
}

func (t *Template) AddFromFS(fsys fs.FS, patterns ...string) error {
	for _, pattern := range patterns {
		matches, err := fs.Glob(fsys, pattern)
		if err != nil {
			return err
		}
		for _, match := range matches {
			content, err := fs.ReadFile(fsys, match)
			if err != nil {
				return err
			}
			name := filepath.Base(match)
			tmpl := t.templates.New(name)
			if _, err := tmpl.Parse(string(content)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *Template) ParseGlob(pattern string) error {
	var err error
	t.templates, err = t.templates.ParseGlob(pattern)
	return err
}
