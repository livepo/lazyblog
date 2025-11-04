package main

import (
	"fmt"
	"html/template"
	"lazyblog/internal/controller"
	"lazyblog/internal/model"
	"lazyblog/pkg/config"
	"lazyblog/pkg/invoker"
	"lazyblog/pkg/middleware"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
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
		"getFromConfig": func(k string) string { return viper.GetString(k) },
		"markdown": func(s string) template.HTML {
			var buf strings.Builder
			if err := goldmark.Convert([]byte(s), &buf); err != nil {
				return template.HTML(s)
			}
			return template.HTML(buf.String())
		},
	})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	if viper.GetBool("debug") {
		router.Use(middleware.Cors())
		gin.SetMode(gin.DebugMode)
	}
	router.LoadHTMLGlob("templates/**/*.tmpl")
	daily := router.Group(config.Cfg.Site.Prefix)
	daily.Static("/static", "./static")
	daily.GET("/", controller.Home)
	daily.GET("/posts", controller.ListPosts)
	daily.GET("/posts/:sid", controller.PostDetail)
	daily.POST("/posts/:sid/like", controller.LikePost)
	daily.POST("/posts/:sid/comment", controller.CreateComment)
	daily.GET("/posts/:sid/comments", controller.ListComments)
	daily.GET("/tags", controller.ListTags)
	daily.GET("/categories", controller.ListCategories)
	daily.GET("/archive", controller.ListArchive)
	daily.GET("/about", controller.About)
	daily.GET("/atom.xml", controller.AtomFeed)
	// router.POST("/posts", controller.CreatePost)
	// router.GET("/admin/create", controller.AdminCreatePostPage)
	// api.PUT("/posts/:sid", controller.UpdatePost)
	// api.DELETE("/posts/:sid", controller.DeletePost)

	router.Run()
}
