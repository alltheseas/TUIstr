package model

import (
	"fmt"
	"strings"
	"time"
)

type Comment struct {
	ID        string
	Author    string
	Text      string
	Timestamp string
	Depth     int
}

type Comments struct {
	PostID        string
	PostTitle     string
	PostAuthor    string
	Community     string
	PostText      string
	PostUrl       string
	PostTimestamp string
	Expiry        time.Time
	Comments      []Comment
}

func (c Comment) Title() string {
	return formatDepth(c.Text, c.Depth)
}

func (c Comment) Description() string {
	desc := fmt.Sprintf("by %s  %s", c.Author, c.Timestamp)
	return formatDepth(desc, c.Depth)
}

func (c Comment) FilterValue() string {
	return c.Author
}

func formatDepth(s string, depth int) string {
	var results strings.Builder
	for range depth {
		results.WriteString("  ")
	}
	results.WriteString(s)

	return results.String()
}
