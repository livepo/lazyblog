package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminCreatePostPage(c *gin.Context) {
	// auth := c.GetHeader("X-Admin-Auth")
	// if auth != config.Cfg.Auth.Username {
	// 	c.AbortWithStatus(http.StatusUnauthorized)
	// 	return
	// }

	c.HTML(http.StatusOK, "create.tmpl", gin.H{
		"Title": "发表文章",
	})
}
