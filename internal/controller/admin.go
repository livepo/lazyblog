package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lazyblog/internal/model"
	"lazyblog/pkg/config"
	"lazyblog/pkg/constant"
	"lazyblog/pkg/invoker"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

func AdminCreatePost(c *gin.Context) {
	token := c.GetHeader("X-Admin-Token")

	if token != config.Cfg.Auth.XAdminToken {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	filename := fileHeader.Filename
	log.Printf("Received file: %s", filename)

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 备份上传的原始文件到 posts/<filename>
	safeName := filepath.Base(filename)
	postsDir := "posts"
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "failed to create posts dir"})
		return
	}
	backupPath := filepath.Join(postsDir, safeName)
	if err := os.WriteFile(backupPath, buf.Bytes(), 0644); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to save backup: %v", err)})
		return
	}

	blog, err := parse(buf.String(), filename)
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

func parse(content string, filename string) (*blog, error) {
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
	if err := invoker.DB.Model(&model.Post{}).Where("file = ?", filename).First(&post).Error; err == nil {
		fmt.Println("Post already exists, updating...")
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
		post.File = filename
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(bodyPart), &buf); err != nil {
			return nil, fmt.Errorf("markdown conversion error: %w", err)
		}
		post.Content = buf.String()
		post.SID = model.GenerateSID()
		invoker.DB.Create(&post)
	}

	return blog, nil
}

func AdminUploadImage(c *gin.Context) {
	token := c.GetHeader("X-Admin-Token")

	if token != config.Cfg.Auth.XAdminToken {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	imageURL, err := uploadToImageHosting(buf.Bytes(), fileHeader.Filename)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message":  "image uploaded successfully",
		"imageUrl": imageURL,
	})

}

func uploadToImageHosting(imageData []byte, filename string) (string, error) {
	for _, hostConfig := range config.Cfg.ImageHostings {
		if hostConfig.Enable {
			switch hostConfig.Provider {
			case "imgurl":
				return uploadToImgurl(imageData, filename, hostConfig)
			default:
				return "", fmt.Errorf("unsupported image hosting provider: %s", hostConfig.Provider)
			}
		}
	}
	return "", fmt.Errorf("no image hosting provider enabled")
}

// uploadToImgurl uploads image to ImgURL (https://www.imgurl.org) and returns the image URL.
func uploadToImgurl(imageData []byte, filename string, hostConfig config.ImageHostingConfig) (string, error) {
	// Basic validation
	if hostConfig.ClientId == "" || hostConfig.ClientSecret == "" {
		return "", fmt.Errorf("imgurl clientId or clientSecret not configured")
	}

	// ImgURL upload endpoint (as in comment)
	endpoint := constant.ImgUrl

	// build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// file field
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file error: %w", err)
	}
	if _, err := fw.Write(imageData); err != nil {
		return "", fmt.Errorf("write file to form error: %w", err)
	}

	// required fields: uid, token
	if err := writer.WriteField("uid", hostConfig.ClientId); err != nil {
		return "", fmt.Errorf("write field uid error: %w", err)
	}
	if err := writer.WriteField("token", hostConfig.ClientSecret); err != nil {
		return "", fmt.Errorf("write field token error: %w", err)
	}
	if hostConfig.AlbumId != "" {
		if err := writer.WriteField("album_id", hostConfig.AlbumId); err != nil {
			return "", fmt.Errorf("write field album_id error: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer error: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("imgurl upload failed: status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	// parse JSON response
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			RelativePath string `json:"relative_path"`
			URL          string `json:"url"`
			ThumbnailURL string `json:"thumbnail_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return "", fmt.Errorf("parse response error: %w", err)
	}
	if res.Code != 200 {
		return "", fmt.Errorf("imgurl error: code=%d msg=%s", res.Code, res.Msg)
	}
	if res.Data.URL != "" {
		return res.Data.URL, nil
	}
	if res.Data.RelativePath != "" {
		return res.Data.RelativePath, nil
	}
	return "", fmt.Errorf("imgurl response missing url")
}
