package controller

import (
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"

	"github.com/gin-gonic/gin"
)

func AtomFeed(c *gin.Context) {
	posts := make([]model.Post, 0)
	invoker.DB.Model(model.Post{}).Order("pub_date desc").Find(&posts)

	c.Header("Content-Type", "application/atom+xml; charset=utf-8")
	c.HTML(200, "atom.tmpl", gin.H{
		"Posts": posts,
	})
}
