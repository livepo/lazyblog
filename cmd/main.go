package main

import (
	"fmt"
	"html/template"
	"lazyblog/internal/controller"
	"lazyblog/internal/model"
	"lazyblog/internal/view"
	"lazyblog/pkg/config"
	"lazyblog/pkg/invoker"
	"lazyblog/pkg/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	invoker.Init()
	pflag.Bool("initdb", false, "create db tables")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	if viper.GetBool("initdb") {
		fmt.Println("initdb...")
		invoker.DB.AutoMigrate(model.Post{}, model.Comment{}, model.FrendLink{})
		return
	}

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"formatAsDate":  view.FormatAsDate,
		"split":         strings.Split,
		"sub":           func(a, b int) int { return a - b },
		"relativeTime":  view.RelativeTime,
		"seq":           view.Seq,
		"add":           func(a, b int) int { return a + b },
		"truncate":      view.Truncate,
		"getFromConfig": func(k string) string { return viper.GetString(k) },
		"markdown":      view.Md2Html,
		"about":         view.AboutMe,
		"getLinks":      view.GetLinks,
		"getCategories": view.GetCategories,
		"getTags":       view.GetTags,
	})

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	if viper.GetBool("debug") {
		router.Use(middleware.Cors())
		gin.SetMode(gin.DebugMode)
	}
	router.LoadHTMLGlob("templates/**/*.tmpl")
	sitePrefix := router.Group(config.Cfg.Site.Prefix)
	sitePrefix.Static("/static", "./static")
	sitePrefix.GET("/", controller.Home)
	sitePrefix.GET("/posts", controller.ListPosts)
	sitePrefix.GET("/posts/:sid", controller.PostDetail)
	sitePrefix.POST("/posts/:sid/like", controller.LikePost)
	sitePrefix.POST("/posts/:sid/comment", controller.CreateComment)
	sitePrefix.GET("/posts/:sid/comments", controller.ListComments)
	sitePrefix.GET("/tags", controller.ListTags)
	sitePrefix.GET("/categories", controller.ListCategories)
	sitePrefix.GET("/archive", controller.ListArchive)
	sitePrefix.GET("/about", controller.About)
	sitePrefix.GET("/atom.xml", controller.AtomFeed)
	// router.POST("/posts", controller.CreatePost)
	router.POST("/admin/publish", controller.AdminCreatePost)
	router.POST("/admin/upload", controller.AdminUploadImage) // 选择合适的图床上传
	router.GET("/ready", func(c *gin.Context) {
		c.String(200, "ok")
	})
	// api.PUT("/posts/:sid", controller.UpdatePost)
	// api.DELETE("/posts/:sid", controller.DeletePost)

	router.Run()
}
