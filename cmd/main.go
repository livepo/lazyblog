package main

import (
	"fmt"
	"html/template"
	"lazyblog/internal/controller"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"lazyblog/pkg/middleware"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

func relativeTimeSimple(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		return "未来"
	}

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d秒前", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		return fmt.Sprintf("%d周前", int(duration.Hours()/(24*7)))
	case duration < 365*24*time.Hour:
		return fmt.Sprintf("%d月前", int(duration.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d年前", int(duration.Hours()/(24*365)))
	}
}

func seq(a, b int) []int {
	if a > b {
		return []int{}
	}
	result := make([]int, b-a+1)
	for i := range result {
		result[i] = a + i
	}
	return result
}

func add(i, j int) int {
	return i + j
}

func main() {
	invoker.Init()
	pflag.Bool("initdb", false, "create db tables")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	if viper.GetBool("initdb") {
		fmt.Println("initdb...")
		invoker.DB.AutoMigrate(model.Post{}, model.Comment{})
		return
	}

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"formatAsDate":       formatAsDate,
		"split":              strings.Split,
		"sub":                func(a, b int) int { return a - b },
		"relativeTimeSimple": relativeTimeSimple,
		"seq":                seq,
		"add":                add,
		"safeHTML":           func(s template.HTML) template.HTML { return s },
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "..."
		},
	})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	if viper.GetBool("debug") {
		router.Use(middleware.Cors())
		gin.SetMode(gin.DebugMode)
	}
	router.LoadHTMLGlob("templates/**/*.tmpl")
	router.Static("static", "./static")
	router.GET("/", controller.Home)
	router.GET("/posts", controller.ListPosts)
	router.GET("/posts/:sid", controller.PostDetail)
	router.POST("/posts/:sid/like", controller.LikePost)
	router.POST("/posts/:sid/comment", controller.CreateComment)
	router.GET("/posts/:sid/comments", controller.ListComments)
	router.GET("/tags", controller.ListTags)
	router.GET("/categories", controller.ListCategories)
	router.GET("/archive", controller.ListArchive)
	router.GET("/about", controller.About)
	// router.POST("/posts", controller.CreatePost)
	// router.GET("/admin/create", controller.AdminCreatePostPage)
	// api.PUT("/posts/:sid", controller.UpdatePost)
	// api.DELETE("/posts/:sid", controller.DeletePost)

	router.Run()
}
