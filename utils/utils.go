package utils

import (
	"fmt"
	"time"
)

func NormalizeCommunity(community string) string {
	return community
}

func TruncateString(s string, w int) string {
	if w <= 0 {
		return s
	} else if len(s) <= w || len(s) <= 3 {
		return s
	}

	return fmt.Sprintf("%s...", s[:w-3])
}

func Clamp(min, max, val int) int {
	if val < min {
		return min
	} else if val > max {
		return max
	}

	return val
}

func GetSingularPlural(s, singular, plural string) string {
	if s == "1" {
		return fmt.Sprintf("%s %s", s, singular)
	}

	return fmt.Sprintf("%s %s", s, plural)
}

func FriendlyTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}

func ShortenPubKey(pk string) string {
	if len(pk) <= 10 {
		return pk
	}

	return fmt.Sprintf("%s...%s", pk[:6], pk[len(pk)-4:])
}
