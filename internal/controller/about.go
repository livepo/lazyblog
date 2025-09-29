package controller

import (
	"bytes"
	"html/template"
	"lazyblog/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
)

func About(c *gin.Context) {
	var buf bytes.Buffer
	goldmark.Convert([]byte(config.Cfg.Site.About), &buf)
	c.HTML(200, "about.tmpl", gin.H{
		"Title":   "关于我",
		"Content": template.HTML(buf.String()),
	})
}
