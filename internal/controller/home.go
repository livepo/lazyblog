package controller

import (
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HomeData struct {
	Title    string
	Posts    []model.Post
	Comments []model.Comment
}

func Home(c *gin.Context) {
	var posts []model.Post
	invoker.DB.Model(model.Post{}).Where("published = ?", true).Order("pub_date desc").Limit(6).Find(&posts)
	var comments []model.Comment
	invoker.DB.Model(model.Comment{}).Where("approved = ?", true).Order("pub_date desc").Limit(5).Find(&comments)
	c.HTML(http.StatusOK, "index.tmpl", HomeData{Title: "首页", Posts: posts, Comments: comments})
}
