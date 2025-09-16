package model

import (
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Post struct {
	gorm.Model            // This will add ID, CreatedAt, UpdatedAt, DeletedAt fields
	ID          int       `gorm:"primaryKey;autoIncrement"`
	SID         int       `gorm:"column:sid;not null;unique" json:"sid"` // Unique identifier for the post
	Title       string    `gorm:"type:varchar(255);not null" json:"title" yaml:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Markdown    string    `gorm:"type:text;not null" json:"markdown" yaml:"markdown"` // Markdown content
	Description string    `gorm:"type:varchar(500)" json:"description" yaml:"description"`
	Author      string    `gorm:"type:varchar(100)" json:"author" yaml:"author"`
	Published   bool      `gorm:"default:false" json:"published" yaml:"published"`
	PubDate     time.Time `gorm:"type:datetime" json:"pub_date" yaml:"pubdate"`
	Tags        string    `gorm:"type:varchar(255)" json:"tags" yaml:"tags"`         // Comma-separated tags
	Category    string    `gorm:"type:varchar(100)" json:"category" yaml:"category"` // Category of the post
	LikesCount  int64     `gorm:"default:0" json:"likes_count"`                      // Number of likes
	File        string    `gorm:"type:varchar(255)" json:"file"`                     // File path if applicable
}

type Comment struct {
	gorm.Model           // This will add ID, CreatedAt, UpdatedAt, DeletedAt fields
	ID         int       `gorm:"primaryKey;autoIncrement" json:"-"`
	SID        int       `gorm:"column:sid;not null" json:"-"`
	PostID     int       `gorm:"column:post_id;not null" json:"-"` // Foreign key to Post
	PostSID    int       `gorm:"column:post_sid" json:"post_sid"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Nickname   string    `gorm:"type:varchar(100);not null" json:"nickname"`
	Email      string    `gorm:"type:varchar(100);not null" json:"email"`
	Website    string    `gorm:"type:varchar(100)" json:"website"`
	Approved   bool      `gorm:"default:false" json:"-"` // Whether the comment is approved
	PubDate    time.Time `gorm:"type:datetime" json:"pub_date"`
}

func GenerateSID() int {
	// This function should generate a unique SID for each post/comment/like.
	return int(time.Now().Unix()-time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC).Unix())*100 + rand.Intn(100)
}

func ParseTags(tagStr string) []string {
	tags := make([]string, 0)
	for _, tag := range SplitAndTrim(tagStr, ",") {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
func SplitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		parts = append(parts, trimmed)
	}
	return parts
}
