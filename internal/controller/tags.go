package controller

import (
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListTagsItem struct {
	Tag   string
	Count int64
}

type ListTagsData struct {
	Title string
	Data  []ListTagsItem
}

func ListTags(c *gin.Context) {
	results := make([]ListTagsItem, 0)
	tags := make([]string, 0)
	invoker.DB.Model(model.Post{}).Select("tags").Where("published = ?", true).Find(&tags)
	tagCountMap := make(map[string]int64)
	for _, tagStr := range tags {
		tagList := model.ParseTags(tagStr)
		for _, tag := range tagList {
			tagCountMap[tag]++
		}
	}
	for tag, count := range tagCountMap {
		results = append(results, ListTagsItem{Tag: tag, Count: count})
	}
	c.HTML(http.StatusOK, "tags.tmpl", ListTagsData{
		Title: "文章标签",
		Data:  results})
}
