package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reddittui/config"
	"reddittui/model"
	"reddittui/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/nbd-wtf/go-nostr/nip22"
)

var (
	ErrNoRelays = errors.New("no relays configured")
	ErrNotFound = errors.New("event not found")
)

type NostrClient struct {
	pool      *nostr.SimplePool
	relays    []string
	timeout   time.Duration
	limit     int
	featured  []string
	community string
}

func NewNostrClient(cfg config.Config) (*NostrClient, error) {
	if len(cfg.Nostr.Relays) == 0 {
		return nil, ErrNoRelays
	}

	ctx := context.Background()
	pool := nostr.NewSimplePool(ctx)

	for _, relay := range cfg.Nostr.Relays {
		if _, err := pool.EnsureRelay(relay); err != nil {
			slog.Warn("Could not connect to relay", "relay", relay, "error", err)
		}
	}

	timeout := time.Duration(cfg.Nostr.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	limit := cfg.Nostr.Limit
	if limit <= 0 {
		limit = 50
	}

	return &NostrClient{
		pool:     pool,
		relays:   cfg.Nostr.Relays,
		timeout:  timeout,
		limit:    limit,
		featured: cfg.Communities.Featured,
	}, nil
}

func (c *NostrClient) Close() {
	c.pool.Close("shutdown")
}

func (c *NostrClient) GetFeaturedPosts(until string) (model.Posts, error) {
	return c.fetchPosts(c.featured, until, true)
}

func (c *NostrClient) GetCommunityPosts(community, until string) (model.Posts, error) {
	return c.fetchPosts([]string{community}, until, false)
}

func (c *NostrClient) GetThread(post model.Post) (model.Comments, error) {
	// Fetch replies referencing the root event
	filters := nostr.Filters{
		{
			Kinds: []int{1, 1111},
			Tags:  nostr.TagMap{"e": {post.ThreadID}},
			Limit: c.limit,
		},
		{
			Kinds: []int{1, 1111},
			Tags:  nostr.TagMap{"E": {post.ThreadID}},
			Limit: c.limit,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	events := c.pool.FetchMany(ctx, c.relays, filters)
	replyMap := make(map[string]nostr.Event)
	for _, evt := range events {
		if evt.ID == post.ThreadID {
			continue
		}
		replyMap[evt.ID] = evt
	}

	thread := buildThread(post.ThreadID, replyMap)

	comments := make([]model.Comment, 0, len(thread))
	for _, evt := range thread {
		comments = append(comments, c.eventToComment(evt.Event, evt.Depth))
	}

	return model.Comments{
		PostID:        post.ID,
		PostTitle:     post.PostTitle,
		PostAuthor:    post.Author,
		Community:     post.Community,
		PostText:      post.Content,
		PostUrl:       post.PostUrl,
		PostTimestamp: utils.FriendlyTime(post.CreatedAt),
		Comments:      comments,
	}, nil
}

func (c *NostrClient) GetPostByID(id string) (model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	events := c.pool.FetchMany(ctx, c.relays, nostr.Filters{
		{
			IDs:   []string{id},
			Limit: 1,
		},
	})

	if len(events) == 0 {
		return model.Post{}, ErrNotFound
	}

	return c.eventToPost(events[0]), nil
}

func (c *NostrClient) fetchPosts(communities []string, until string, isHome bool) (model.Posts, error) {
	filter := nostr.Filter{
		Kinds: []int{1111},
		Limit: c.limit,
	}

	if len(communities) > 0 {
		filter.Tags = nostr.TagMap{"I": communities}
	}

	if cursor := parseCursor(until); cursor != nil {
		filter.Until = cursor
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	events := c.pool.FetchMany(ctx, c.relays, nostr.Filters{filter})
	if len(events) == 0 {
		return model.Posts{IsHome: isHome, Community: strings.Join(communities, ",")}, nil
	}

	dedup := make(map[string]nostr.Event)
	for _, evt := range events {
		dedup[evt.ID] = evt
	}

	uniqueEvents := make([]nostr.Event, 0, len(dedup))
	for _, evt := range dedup {
		uniqueEvents = append(uniqueEvents, evt)
	}

	sort.Slice(uniqueEvents, func(i, j int) bool {
		return uniqueEvents[i].CreatedAt > uniqueEvents[j].CreatedAt
	})

	posts := make([]model.Post, 0, len(uniqueEvents))
	for _, evt := range uniqueEvents {
		posts = append(posts, c.eventToPost(evt))
	}

	after := ""
	if len(uniqueEvents) > 0 {
		oldest := uniqueEvents[len(uniqueEvents)-1].CreatedAt
		after = strconv.FormatInt(int64(oldest), 10)
	}

	description := "Open community posts"
	communityLabel := "Communities"
	if !isHome && len(communities) == 1 {
		communityLabel = communities[0]
		description = fmt.Sprintf("Posts tagged %s", communities[0])
	} else if isHome {
		description = "Featured communities timeline"
	}

	return model.Posts{
		Description: description,
		Community:   communityLabel,
		IsHome:      isHome,
		Posts:       posts,
		After:       after,
		Expiry:      time.Now().Add(30 * time.Minute),
	}, nil
}

func (c *NostrClient) eventToPost(evt nostr.Event) model.Post {
	created := time.Unix(int64(evt.CreatedAt), 0)
	title := evt.Content
	if subject := evt.Tags.GetFirst([]string{"subject"}); subject != nil && len(*subject) > 1 && strings.TrimSpace((*subject)[1]) != "" {
		title = (*subject)[1]
	}

	title = strings.TrimSpace(title)
	title = firstLine(title)
	if title == "" {
		title = "(untitled)"
	}

	community := extractCommunity(evt.Tags)
	if community == "" {
		community = "untagged"
	}

	postUrl := fmt.Sprintf("https://njump.me/%s", evt.ID)
	if nevent, err := nip19.EncodeEvent(evt.ID, evt.PubKey, "", nil); err == nil {
		postUrl = fmt.Sprintf("https://njump.me/%s", nevent)
	}

	return model.Post{
		ID:           evt.ID,
		PostTitle:    title,
		Content:      evt.Content,
		Author:       utils.ShortenPubKey(evt.PubKey),
		Community:    community,
		FriendlyDate: utils.FriendlyTime(created),
		CreatedAt:    created,
		PostUrl:      postUrl,
		ThreadID:     evt.ID,
	}
}

func (c *NostrClient) eventToComment(evt nostr.Event, depth int) model.Comment {
	created := time.Unix(int64(evt.CreatedAt), 0)
	return model.Comment{
		ID:        evt.ID,
		Author:    utils.ShortenPubKey(evt.PubKey),
		Text:      evt.Content,
		Timestamp: utils.FriendlyTime(created),
		Depth:     depth,
	}
}

func extractCommunity(tags nostr.Tags) string {
	for _, tag := range tags {
		if len(tag) < 2 {
			continue
		}

		if tag[0] == "I" || tag[0] == "i" {
			return tag[1]
		}
	}

	return ""
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}

func parseCursor(cursor string) *nostr.Timestamp {
	if cursor == "" {
		return nil
	}

	ts, err := strconv.ParseInt(cursor, 10, 64)
	if err != nil {
		return nil
	}

	t := nostr.Timestamp(ts)
	return &t
}

type threadNode struct {
	Event nostr.Event
	Depth int
}

func buildThread(rootID string, events map[string]nostr.Event) []threadNode {
	children := make(map[string][]nostr.Event)
	for _, evt := range events {
		parentID := rootID
		if ptr := nip22.GetImmediateParent(evt.Tags); ptr != nil {
			switch p := ptr.(type) {
			case nostr.EventPointer:
				parentID = p.ID
			case nostr.EntityPointer:
				parentID = p.ID
			case nostr.Pointer:
				if val := p.AsTagReference(); val != "" {
					parentID = val
				}
			}
		}

		// orphaned replies go back to root
		if _, ok := events[parentID]; !ok && parentID != rootID {
			parentID = rootID
		}

		children[parentID] = append(children[parentID], evt)
	}

	for parent := range children {
		sort.Slice(children[parent], func(i, j int) bool {
			return children[parent][i].CreatedAt < children[parent][j].CreatedAt
		})
	}

	var ordered []threadNode
	var walk func(parent string, depth int)
	walk = func(parent string, depth int) {
		for _, child := range children[parent] {
			ordered = append(ordered, threadNode{Event: child, Depth: depth})
			walk(child.ID, depth+1)
		}
	}

	walk(rootID, 0)
	return ordered
}
