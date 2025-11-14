package view

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"lazyblog/internal/model"
	"lazyblog/pkg/invoker"
	"os"
	"sort"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

func GetLinks() []model.FrendLink {
	var links []model.FrendLink
	invoker.DB.Model(&model.FrendLink{}).Where("enabled = ?", true).Find(&links)
	return links
}

type CategoryWithCount struct {
	Name      string
	PostCount int64
}

func GetCategories() []CategoryWithCount {
	var categories []CategoryWithCount
	var posts []model.Post

	invoker.DB.Model(model.Post{}).Find(&posts)
	categoryCountMap := make(map[string]int64)
	for _, post := range posts {
		categoryCountMap[post.Category]++
	}
	for category, count := range categoryCountMap {
		categories = append(categories, CategoryWithCount{
			Name:      category,
			PostCount: count,
		})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].PostCount < categories[j].PostCount
	})
	return categories
}

type TagWithCount struct {
	Name      string
	PostCount int64
}

func GetTags() []TagWithCount {
	var tags []TagWithCount
	var posts []model.Post

	invoker.DB.Model(model.Post{}).Find(&posts)
	tagCountMap := make(map[string]int64)
	for _, post := range posts {
		postTags := model.ParseTags(post.Tags)
		for _, tag := range postTags {
			tagCountMap[tag]++
		}
	}
	for tag, count := range tagCountMap {
		tags = append(tags, TagWithCount{
			Name:      tag,
			PostCount: count,
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PostCount < tags[j].PostCount
	})
	return tags
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

func RelativeTime(t time.Time) string {
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

func Seq(a, b int) []int {
	if a > b {
		return []int{}
	}
	result := make([]int, b-a+1)
	for i := range result {
		result[i] = a + i
	}
	return result
}

func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func Md2Html(md string) template.HTML {
	mdParser := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, Theme: "dark"},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf []byte
	writer := bytes.NewBuffer(buf)
	if err := mdParser.Convert([]byte(md), writer); err != nil {
		return ""
	}
	return template.HTML(writer.String())
}

func AboutMe() template.HTML {
	return Md2Html(viper.GetString("site.about"))
}

func CssEtag() string {
	path := "static/" + viper.GetString("site.css")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := md5.Sum(data)
	v := hex.EncodeToString(sum[:6])
	return v
}
