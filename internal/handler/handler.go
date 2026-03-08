package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/session"
)

type Flash struct {
	Message string
	Type    string
}

func getFlash(c echo.Context) []Flash {
	sess := c.Get("session").(*session.Session)
	flashes := sess.Flashes()
	sess.Save(c.Request(), c.Response())
	
	var result []Flash
	for _, f := range flashes {
		if m, ok := f.(map[string]string); ok {
			result = append(result, Flash{Message: m["message"], Type: m["type"]})
		}
	}
	return result
}

func addFlash(c echo.Context, message string, flashType string) {
	sess := c.Get("session").(*session.Session)
	sess.AddFlash(map[string]string{"message": message, "type": flashType})
	sess.Save(c.Request(), c.Response())
}

func Login(c echo.Context) error {
	user := c.Get("currentUser")
	if user != nil {
		return c.Redirect(302, "/all")
	}
	
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"config": c.Get("config"),
		"flashes": getFlash(c),
	})
}

func ProcessLogin(c echo.Context) error {
	password := c.FormValue("password")

	if password == "" {
		addFlash(c, "密码不能为空", "error")
		return c.Redirect(302, "/login")
	}

	user, err := db.GetUserByUsername("admin")
	if err != nil {
		addFlash(c, "系统错误", "error")
		return c.Redirect(302, "/login")
	}

	if !user.VerifyPassword(password) {
		addFlash(c, "密码错误", "error")
		return c.Redirect(302, "/login")
	}

	sess := c.Get("session").(*session.Session)
	sess.Values["user"] = user.ID
	sess.Save(c.Request(), c.Response())

	return c.Redirect(302, "/all")
}

func Logout(c echo.Context) error {
	session.Delete(c, "session")
	return c.Redirect(302, "/all")
}

func ChangePassword(c echo.Context) error {
	return c.Render(http.StatusOK, "password.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"flashes": getFlash(c),
	})
}

func ProcessChangePassword(c echo.Context) error {
	user := c.Get("currentUser").(*db.User)
	oldPassword := c.FormValue("old_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if !user.VerifyPassword(oldPassword) {
		addFlash(c, "原密码错误", "error")
		return c.Redirect(302, "/settings/password")
	}

	if newPassword != confirmPassword {
		addFlash(c, "两次输入的新密码不一致", "error")
		return c.Redirect(302, "/settings/password")
	}

	if len(newPassword) < 6 {
		addFlash(c, "密码长度至少6位", "error")
		return c.Redirect(302, "/settings/password")
	}

	if err := user.SetPassword(newPassword); err != nil {
		addFlash(c, "密码加密失败", "error")
		return c.Redirect(302, "/settings/password")
	}

	if err := user.Update(); err != nil {
		addFlash(c, "密码保存失败", "error")
		return c.Redirect(302, "/settings/password")
	}

	addFlash(c, "密码修改成功", "success")
	return c.Redirect(302, "/all")
}

func AllGists(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	sort := c.QueryParam("sort")
	if sort != "created" && sort != "updated" {
		sort = "created"
	}
	order := c.QueryParam("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	query := c.QueryParam("q")

	var gists []*db.Gist
	var total int64
	var err error

	if query != "" {
		gists, err = db.SearchGists(query, page-1, sort, order)
		total, _ = db.CountSearchGists(query)
	} else {
		gists, err = db.GetAllGists(page-1, sort, order)
		total, _ = db.CountAllGists()
	}

	if err != nil {
		return err
	}

	hasMore := len(gists) > 10
	if hasMore {
		gists = gists[:10]
	}

	return c.Render(http.StatusOK, "all.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"gists": gists,
		"page": page,
		"hasMore": hasMore,
		"sort": sort,
		"order": order,
		"query": query,
		"total": total,
		"flashes": getFlash(c),
	})
}

func NewGist(c echo.Context) error {
	return c.Render(http.StatusOK, "new.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"flashes": getFlash(c),
		"form": map[string]string{
			"title": "",
			"content": "",
			"visibility": "public",
		},
	})
}

func CreateGist(c echo.Context) error {
	title := c.FormValue("title")
	content := c.FormValue("content")
	filename := c.FormValue("filename")
	visibility := c.FormValue("visibility")

	formData := map[string]string{
		"title": title,
		"content": content,
		"visibility": visibility,
	}

	if title == "" || content == "" {
		return c.Render(http.StatusOK, "new.html", map[string]interface{}{
			"config": c.Get("config"),
			"user": c.Get("currentUser"),
			"flashes": []Flash{{Message: "标题和内容不能为空", Type: "error"}},
			"form": formData,
		})
	}

	if !db.IsValidTitle(title) {
		return c.Render(http.StatusOK, "new.html", map[string]interface{}{
			"config": c.Get("config"),
			"user": c.Get("currentUser"),
			"flashes": []Flash{{Message: "标题只能包含大小写字母和数字", Type: "error"}},
			"form": formData,
		})
	}

	exists, _ := db.TitleExists(title, 0)
	if exists {
		return c.Render(http.StatusOK, "new.html", map[string]interface{}{
			"config": c.Get("config"),
			"user": c.Get("currentUser"),
			"flashes": []Flash{{Message: "标题已存在，请更换", Type: "error"}},
			"form": formData,
		})
	}

	if filename == "" {
		filename = title + ".txt"
	}

	vis := db.PublicVisibility
	if visibility == "private" {
		vis = db.PrivateVisibility
	}

	user := c.Get("currentUser").(*db.User)
	gist := &db.Gist{
		Title:      title,
		Filename:   filename,
		Content:    content,
		Visibility: vis,
		UserID:     user.ID,
	}

	if err := gist.Create(); err != nil {
		return c.Render(http.StatusOK, "new.html", map[string]interface{}{
			"config": c.Get("config"),
			"user": c.Get("currentUser"),
			"flashes": []Flash{{Message: "创建失败", Type: "error"}},
			"form": formData,
		})
	}

	addFlash(c, "创建成功", "success")
	return c.Redirect(302, "/all")
}

func ViewGist(c echo.Context) error {
	title := c.Param("title")
	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.String(http.StatusNotFound, "Gist不存在")
	}

	user := c.Get("currentUser")
	if gist.Visibility == db.PrivateVisibility && user == nil {
		return c.String(http.StatusForbidden, "此Gist为私有，请登录查看")
	}

	return c.Blob(http.StatusOK, "text/plain; charset=utf-8", []byte(gist.Content))
}

func GistDetail(c echo.Context) error {
	title := c.Param("title")
	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	user := c.Get("currentUser")

	return c.Render(http.StatusOK, "gist.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": user,
		"gist": gist,
		"flashes": getFlash(c),
	})
}

func EditGist(c echo.Context) error {
	title := c.Param("title")
	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	return c.Render(http.StatusOK, "edit.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"gist": gist,
		"flashes": getFlash(c),
	})
}

func UpdateGist(c echo.Context) error {
	oldTitle := c.Param("title")
	gist, err := db.GetGistByTitle(oldTitle)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	title := c.FormValue("title")
	content := c.FormValue("content")
	filename := c.FormValue("filename")
	visibility := c.FormValue("visibility")

	if title == "" || content == "" {
		addFlash(c, "标题和内容不能为空", "error")
		return c.Redirect(302, "/"+oldTitle+"/edit")
	}

	if !db.IsValidTitle(title) {
		addFlash(c, "标题只能包含大小写字母和数字", "error")
		return c.Redirect(302, "/"+oldTitle+"/edit")
	}

	if title != oldTitle {
		exists, _ := db.TitleExists(title, gist.ID)
		if exists {
			addFlash(c, "标题已存在，请更换", "error")
			return c.Redirect(302, "/"+oldTitle+"/edit")
		}
	}

	if filename == "" {
		filename = title + ".txt"
	}

	vis := db.PublicVisibility
	if visibility == "private" {
		vis = db.PrivateVisibility
	}

	version, _ := db.GetNextVersion(gist.ID)
	revision := &db.Revision{
		GistID:     gist.ID,
		Version:    version,
		Title:      gist.Title,
		Filename:   gist.Filename,
		Content:    gist.Content,
		Visibility: gist.Visibility,
	}
	revision.Create()

	gist.Title = title
	gist.Filename = filename
	gist.Content = content
	gist.Visibility = vis

	if err := gist.Update(); err != nil {
		addFlash(c, "更新失败", "error")
		return c.Redirect(302, "/"+oldTitle+"/edit")
	}

	addFlash(c, "更新成功", "success")
	return c.Redirect(302, "/"+title+"/detail")
}

func DeleteGist(c echo.Context) error {
	title := c.Param("title")
	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	db.DeleteRevisionsByGistId(gist.ID)
	gist.Delete()

	addFlash(c, "删除成功", "success")
	return c.Redirect(302, "/all")
}

func Revisions(c echo.Context) error {
	title := c.Param("title")
	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	revisions, _ := db.GetRevisionsByGistId(gist.ID)

	return c.Render(http.StatusOK, "revisions.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"gist": gist,
		"revisions": revisions,
	})
}

func ViewRevision(c echo.Context) error {
	title := c.Param("title")
	version, _ := strconv.Atoi(c.Param("version"))

	gist, err := db.GetGistByTitle(title)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "Gist不存在",
		})
	}

	revision, err := db.GetRevisionByGistIdAndVersion(gist.ID, version)
	if err != nil {
		return c.Render(http.StatusNotFound, "error.html", map[string]interface{}{
			"config": c.Get("config"),
			"message": "版本不存在",
		})
	}

	return c.Render(http.StatusOK, "revision.html", map[string]interface{}{
		"config": c.Get("config"),
		"user": c.Get("currentUser"),
		"gist": gist,
		"revision": revision,
	})
}
