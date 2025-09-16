package controller

import (
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateCommentRequest struct {
	Content  string `form:"comment"`
	Nickname string `form:"nickname"`
	Email    string `form:"email"`
	Website  string `form:"website"`
}

func CreateComment(c *gin.Context) {
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

	req := CreateCommentRequest{}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	comment := model.Comment{
		SID:      model.GenerateSID(),
		PostID:   post.ID,
		PostSID:  post.SID,
		Content:  req.Content,
		Nickname: req.Nickname,
		Email:    req.Email,
		Website:  req.Website,
		Approved: true,
		PubDate:  time.Now(),
	}
	invoker.DB.Create(&comment)
	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

type commentViewItem struct {
	Content  string `json:"content"`
	Nickname string `json:"nickname"`
	Website  string `json:"website"`
	PubDate  string `json:"pub_date"`
}

func ListComments(c *gin.Context) {
	sid := c.Param("sid")
	var post model.Post
	err := invoker.DB.Model(model.Post{}).Where("sid = ?", sid).First(&post).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	comments := make([]commentViewItem, 0)
	invoker.DB.Model(model.Comment{}).Where("post_id = ?", post.ID).Order("pub_date DESC").Find(&comments)

	c.JSON(http.StatusOK, comments)
}
