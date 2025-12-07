package utils

import (
	"testing"
	"time"
)

func TestNormalizeCommunity(t *testing.T) {
	if got := NormalizeCommunity(" T:NoStr "); got != "t:nostr" {
		t.Fatalf("expected normalized community id, got %s", got)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"abc", 3, "abc"},
		{"abcd", 4, "abcd"},
		{"abcde", 4, "a..."},
		{"abcdef", 5, "ab..."},
		{"abcdefg", 6, "abc..."},
	}

	for _, tt := range tests {
		got := TruncateString(tt.s, tt.width)
		if got != tt.want {
			t.Errorf("got %s, want %s with input %s", got, tt.want, tt.s)
		}
	}
}

func TestValidateTopic(t *testing.T) {
	valid := []string{"t:nostr", "t:linux", "t:city:nyc", "t:go-lang", "t:foo_bar"}
	for _, c := range valid {
		if !ValidateTopic(c) {
			t.Fatalf("expected valid topic for %s", c)
		}
	}

	invalid := []string{"", "nostr", "u:wss://relay", "t:", "t:!", "t: space", "t:ä"}
	for _, c := range invalid {
		if ValidateTopic(c) {
			t.Fatalf("expected invalid topic for %s", c)
		}
	}
}

func TestCopyToClipboardEmpty(t *testing.T) {
	if err := CopyToClipboard(""); err == nil {
		t.Fatalf("expected error copying empty text")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		min  int
		max  int
		val  int
		want int
	}{
		{0, 2, 0, 0},
		{0, 2, 1, 1},
		{0, 2, 2, 2},
		{0, 2, 3, 2},
		{0, 2, -1, 0},
		{0, 10, 5, 5},
		{0, 10, -5, 0},
		{0, 10, 15, 10},
	}

	for _, tt := range tests {
		got := Clamp(tt.min, tt.max, tt.val)
		if got != tt.want {
			t.Errorf("got %d, want %d with input: min %d, max %d, val %d", got, tt.want, tt.min, tt.max, tt.val)
		}
	}
}

func TestGetSingularPlural(t *testing.T) {
	tests := []struct {
		s        string
		singular string
		plural   string
		want     string
	}{
		{"0", "banana", "bananas", "0 bananas"},
		{"1", "banana", "bananas", "1 banana"},
		{"2", "banana", "bananas", "2 bananas"},
		{"3", "banana", "bananas", "3 bananas"},
	}

	for _, tt := range tests {
		got := GetSingularPlural(tt.s, tt.singular, tt.plural)
		if got != tt.want {
			t.Errorf("got %s, want %s with input: s %s, singular %s, plural %s", got, tt.want, tt.s, tt.singular, tt.plural)
		}
	}
}

func TestFriendlyTime(t *testing.T) {
	now := time.Now()
	if out := FriendlyTime(now.Add(-30 * time.Second)); out != "just now" {
		t.Fatalf("expected 'just now', got %s", out)
	}
	if out := FriendlyTime(now.Add(-2 * time.Hour)); out == "" || out == "just now" {
		t.Fatalf("expected friendly hours string, got %s", out)
	}
}

func TestShortenPubKey(t *testing.T) {
	short := ShortenPubKey("abcdef")
	if short != "abcdef" {
		t.Fatalf("expected unchanged short key, got %s", short)
	}

	long := ShortenPubKey("abcdef0123456789")
	if long == "abcdef0123456789" {
		t.Fatalf("expected shortened key, got %s", long)
	}
}
