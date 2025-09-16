package controller

import (
	"fmt"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type ListArchiveItem struct {
	Year  int
	Month int
	Posts []model.Post
}

type ListArchiveData struct {
	Title string
	Data  []ListArchiveItem
}

func ListArchive(c *gin.Context) {
	posts := make([]model.Post, 0)
	invoker.DB.Model(model.Post{}).Where("published = ?", true).Order("pub_date DESC").Find(&posts)
	results := make([]ListArchiveItem, 0)

	archiveMap := make(map[string][]model.Post)
	for _, post := range posts {
		year, month, _ := post.PubDate.Date()
		key := fmt.Sprintf("%04d-%02d", year, month)
		archiveMap[key] = append(archiveMap[key], post)
	}
	for key, postList := range archiveMap {
		var year, month int
		fmt.Sscanf(key, "%04d-%02d", &year, &month)
		results = append(results, ListArchiveItem{
			Year:  year,
			Month: month,
			Posts: postList,
		})
	}
	// 按时间降序排序
	sort.Slice(results, func(i, j int) bool {
		if results[i].Year != results[j].Year {
			return results[i].Year > results[j].Year
		}
		return results[i].Month > results[j].Month
	})

	c.HTML(http.StatusOK, "archive.tmpl", ListArchiveData{
		Title: "文章归档",
		Data:  results})
}
