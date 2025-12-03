package model

import (
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID           string
	PostTitle    string
	Content      string
	Author       string
	Community    string
	FriendlyDate string
	CreatedAt    time.Time
	PostUrl      string
	ThreadID     string
}

type Posts struct {
	Description string
	Community   string
	IsHome      bool
	Posts       []Post
	After       string
	Expiry      time.Time
}

func (p Post) Title() string {
	return p.PostTitle
}

func (p Post) Description() string {
	var sb strings.Builder
	if strings.TrimSpace(p.Community) != "" {
		sb.WriteString(p.Community)
		sb.WriteString("  ")
	}

	fmt.Fprintf(&sb, "posted %s by %s", p.FriendlyDate, p.Author)
	return sb.String()
}

func (p Post) FilterValue() string {
	return p.PostTitle
}
