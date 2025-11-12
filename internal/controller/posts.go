package controller

import (
	"html/template"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type ListPostsData struct {
	Title     string
	Posts     []model.Post
	Page      int
	Size      int
	TotalPage int
}

func ListPosts(c *gin.Context) {
	pageStr := c.Query("page")
	page := cast.ToInt(pageStr)
	if page <= 0 {
		page = 1
	}
	sizeStr := c.Query("size")
	size := cast.ToInt(sizeStr)
	if size <= 0 {
		size = 10
	}
	tag := c.Query("tags")
	category := c.Query("category")

	posts := make([]model.Post, 0)

	query := invoker.DB.Model(model.Post{}).Where("published = ?", true)
	if tag != "" {
		query = query.Where("FIND_IN_SET(?, tags)", tag)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query = query.Order("pub_date DESC")
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * size).Limit(size).Find(&posts)

	totalPage := int(total) / size
	if int(total)%size != 0 {
		totalPage += 1
	}
	c.HTML(http.StatusOK, "posts.tmpl", ListPostsData{
		Title:     "文章列表",
		Posts:     posts,
		Page:      page,
		Size:      size,
		TotalPage: totalPage,
	})
}

type PostDetailData struct {
	Post     model.Post
	Comments []model.Comment
	Content  template.HTML
}

func PostDetail(c *gin.Context) {
	sid := c.Param("sid")
	if sid == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var post model.Post
	err := invoker.DB.Model(model.Post{}).Where("sid = ?", sid).First(&post).Error
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	comments := make([]model.Comment, 0)
	invoker.DB.Model(model.Comment{}).Where("post_id = ?", post.ID).Order("pub_date DESC").Find(&comments)

	c.HTML(http.StatusOK, "detail.tmpl", PostDetailData{Post: post, Comments: comments, Content: template.HTML(post.Content)})
}

func LikePost(c *gin.Context) {
	sid := c.Param("sid")
	if sid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var post model.Post
	err := invoker.DB.Model(model.Post{}).Where("sid = ?", sid).First(&post).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	post.LikesCount += 1
	invoker.DB.Updates(&post)
	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}
