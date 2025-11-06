package controller

import (
	"bytes"
	"fmt"
	"lazyblog/internal/model"
	"lazyblog/pkg/config"
	"lazyblog/pkg/invoker"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

type AdminPostRequest struct {
	Content string `json:"content" binding:"required"`
}

func AdminCreatePost(c *gin.Context) {
	token := c.GetHeader("X-Admin-Token")

	if token != config.Cfg.Auth.XAdminToken {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	var req AdminPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	blog, err := parse(req.Content)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message": "post published successfully",
		"title":   blog.Title,
	})
}

type blog struct {
	Title       string `yaml:"title"`
	Markdown    string `yaml:"markdown"` // Markdown content
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Published   bool   `yaml:"published"`
	PubDate     string `yaml:"pubdate"`
	Tags        string `yaml:"tags"`     // Comma-separated tags
	Category    string `yaml:"category"` // Category of the post
}

func parse(content string) (*blog, error) {
	// 找到 front-matter
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("file missing front-matter")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid front-matter format")
	}

	yamlPart := parts[1]
	bodyPart := strings.TrimSpace(parts[2])

	blog := &blog{}
	if err := yaml.Unmarshal([]byte(yamlPart), blog); err != nil {
		return nil, fmt.Errorf("yaml parse error: %w", err)
	}
	blog.Markdown = bodyPart

	var post model.Post
	if err := invoker.DB.Model(&model.Post{}).Where("title = ?", blog.Title).First(&post).Error; err == nil {
		fmt.Println("Post already exists, updating...")
		// post.Title = blog.Title
		post.Description = blog.Description
		post.Author = blog.Author
		post.Published = blog.Published
		post.PubDate, _ = time.Parse("2006-01-02", blog.PubDate)
		post.Tags = blog.Tags
		post.Category = blog.Category
		post.Markdown = bodyPart
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(bodyPart), &buf); err != nil {
			return nil, fmt.Errorf("markdown conversion error: %w", err)
		}
		post.Content = buf.String()
		invoker.DB.Save(&post)
	} else {
		// create new post
		fmt.Println("Creating new post...")
		post.Title = blog.Title
		post.Description = blog.Description
		post.Author = blog.Author
		post.Published = blog.Published
		post.PubDate, _ = time.Parse("2006-01-02", blog.PubDate)
		post.Tags = blog.Tags
		post.Category = blog.Category
		post.Markdown = bodyPart
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(bodyPart), &buf); err != nil {
			return nil, fmt.Errorf("markdown conversion error: %w", err)
		}
		post.Content = buf.String()
		post.SID = model.GenerateSID()
		post.File = "create by api"
		invoker.DB.Create(&post)
	}

	return blog, nil
}
