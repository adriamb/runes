package web

import (
	"github.com/amassanet/gopad/model"
	"github.com/amassanet/gopad/server"
	"github.com/amassanet/gopad/store"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	UNDEF = "undefined"
)

func InitWeb() {

	server.Srv.Engine.GET("/", HtmlMain)
	server.Srv.Engine.GET("/entry/:id", HtmlGetEntry)
	server.Srv.Engine.POST("/entry/:id", HtmlPostEntry)

}

type HtmlEntry struct {
	*model.Entry
}

func NewHtmlEntry(entry *model.Entry) *HtmlEntry {
	return &HtmlEntry{entry}
}

func HtmlMain(c *gin.Context) {

	resp := <-server.Srv.Store.Entry.List()
	if resp.Err != nil {
		HtmlDumpError(c, resp.Err)
		return
	}

	entries := resp.Data.([]*model.Entry)

	htmlEntries := []HtmlEntry{}
	for _, entry := range entries {
		htmlEntries = append(htmlEntries, HtmlEntry{entry})
	}

	var err error = nil
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":   "Main website",
		"prefix":  server.Srv.Config.Prefix,
		"entries": htmlEntries,
		"error":   err,
	})
}

func ButtonPressed(c *gin.Context, name string) bool {
	return c.DefaultPostForm(name, UNDEF) != UNDEF
}

func HtmlDumpError(c *gin.Context, err error) {
	c.HTML(http.StatusOK, "500.tmpl", gin.H{
		"prefix":  server.Srv.Config.Prefix,
		"message": err.Error(),
	})
}

func HtmlGetEntry(c *gin.Context) {

	id := c.Param("id")

	var entry *model.Entry

	if id != "new" {
		result := <-server.Srv.Store.Entry.Get(id)
		if result.Err != nil {
			HtmlDumpError(c, result.Err)
			return
		}
		entry = result.Data.(*model.Entry)
	} else {
		entry = &model.Entry{Id: "new"}
	}

	c.HTML(http.StatusOK, "entry.tmpl", gin.H{
		"prefix": server.Srv.Config.Prefix,
		"entry":  entry,
	})

}

func HtmlPostEntry(c *gin.Context) {

	if ButtonPressed(c, "btnsave") {

		entry := model.Entry{
			Id:       c.Param("id"),
			Title:    c.DefaultPostForm("Title", "undefined"),
			Markdown: c.DefaultPostForm("Markdown", "undefined"),
		}

		var result store.StoreResult
		if entry.Id == "new" {
			result = <-server.Srv.Store.Entry.Add(&entry)
		} else {
			result = <-server.Srv.Store.Entry.Update(&entry)
		}
		if result.Err != nil {
			HtmlDumpError(c, result.Err)
			return
		}
		c.Redirect(301, server.Srv.Config.Prefix+"/")
		return
	}

	if ButtonPressed(c, "btnback") {
		c.Redirect(301, server.Srv.Config.Prefix+"/")
		return
	}

}
