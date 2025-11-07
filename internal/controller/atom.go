package controller

import (
	"bytes"
	htmltmpl "html/template"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func AtomFeed(c *gin.Context) {
	posts := make([]model.Post, 0)
	invoker.DB.Model(model.Post{}).Order("pub_date desc").Find(&posts)
	// Build data and render template directly so we can set correct Content-Type
	// Create a template with minimal helpers used by atom.tmpl
	tpl, err := template.New("atom.tmpl").Funcs(template.FuncMap{
		"getFromConfig": func(k string) string { return viper.GetString(k) },
		"split":         strings.Split,
	}).ParseFiles("templates/pages/atom.tmpl")
	if err != nil {
		c.String(500, "template parse error: %v", err)
		return
	}

	// Convert posts so Content is treated as safe HTML
	type postForTpl struct {
		Title       string
		SID         int
		PubDate     time.Time
		Description string
		Content     htmltmpl.HTML
		Tags        string
		Author      string
	}
	feedPosts := make([]postForTpl, 0, len(posts))
	for _, p := range posts {
		feedPosts = append(feedPosts, postForTpl{
			Title:       p.Title,
			SID:         p.SID,
			PubDate:     p.PubDate,
			Description: p.Description,
			Content:     htmltmpl.HTML(p.Content),
			Tags:        p.Tags,
			Author:      p.Author,
		})
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, gin.H{"Posts": feedPosts}); err != nil {
		c.String(500, "template execute error: %v", err)
		return
	}
	c.Header("Content-Type", "application/atom+xml; charset=utf-8")
	c.String(200, buf.String())
}
