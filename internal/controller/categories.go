package controller

import (
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListCategoriesItem struct {
	Category string
	Count    int64
}

type ListCategoriesData struct {
	Title string
	Data  []ListCategoriesItem
}

func ListCategories(c *gin.Context) {
	results := make([]ListCategoriesItem, 0)
	invoker.DB.Model(model.Post{}).Select("category, count(*) as count").Where("published = ?", true).Group("category").Find(&results)

	c.HTML(http.StatusOK, "categories.tmpl", ListCategoriesData{
		Title: "文章分类",
		Data:  results})
}
