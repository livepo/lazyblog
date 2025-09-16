package main

import (
	"bytes"
	"fmt"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

type Blog struct {
	Title       string `yaml:"title"`
	Markdown    string `yaml:"markdown"` // Markdown content
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Published   bool   `yaml:"published"`
	PubDate     string `yaml:"pubdate"`
	Tags        string `yaml:"tags"`     // Comma-separated tags
	Category    string `yaml:"category"` // Category of the post
}

func parseFile(filename string) (*Blog, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// 将文件内容转为字符串
	content := string(data)

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

	blog := &Blog{}
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
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(bodyPart), &buf); err != nil {
			return nil, fmt.Errorf("markdown conversion error: %w", err)
		}
		post.Content = buf.String()
		post.SID = model.GenerateSID()
		post.File = filename
		invoker.DB.Create(&post)
	}

	return blog, nil
}

func main() {
	invoker.Init()
	var rootCmd = &cobra.Command{
		Use:   "blogparser [file]",
		Short: "Parse blog file with front-matter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			blog, err := parseFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse file: %w", err)
			}

			fmt.Printf("Title: %s\n", blog.Title)
			fmt.Printf("Author: %s\n", blog.Author)
			fmt.Printf("Published: %v\n", blog.Published)
			fmt.Printf("Description: %s\n", blog.Description)
			fmt.Printf("Tags: %v\n", blog.Tags)
			fmt.Printf("Category: %s\n", blog.Category)
			fmt.Printf("PubDate: %s\n", blog.PubDate)
			fmt.Printf("Markdown:\n%s\n", blog.Markdown)

			return nil
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
