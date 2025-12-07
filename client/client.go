package client

import (
	"context"
	"encoding/hex"
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
	ErrNoRelays         = errors.New("no relays configured")
	ErrNotFound         = errors.New("event not found")
	ErrNoPrivateKey     = errors.New("no nostr private key configured")
	ErrInvalidCommunity = errors.New("community must be a topic (t:...) for now")
	ErrInvalidThreadID  = errors.New("cannot reply: thread id is not a nostr event id (likely demo data)")
)

type NostrClient struct {
	pool        *nostr.SimplePool
	relays      []string
	timeout     time.Duration
	limit       int
	featured    []string
	community   string
	privKey     string
	pubKey      string
	postCache   *simpleCache[model.Posts]
	threadCache *simpleCache[model.Comments]
}

func NewNostrClient(cfg config.Config) (*NostrClient, error) {
	if len(cfg.Nostr.Relays) == 0 {
		return nil, ErrNoRelays
	}

	var (
		privKey string
		pubKey  string
		err     error
	)
	if strings.TrimSpace(cfg.Nostr.SecretKey) != "" {
		privKey, pubKey, err = parsePrivKey(cfg.Nostr.SecretKey)
		if err != nil {
			return nil, err
		}
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
		pool:        pool,
		relays:      cfg.Nostr.Relays,
		timeout:     timeout,
		limit:       limit,
		featured:    cfg.Communities.Featured,
		privKey:     privKey,
		pubKey:      pubKey,
		postCache:   newSimpleCache[model.Posts](),
		threadCache: newSimpleCache[model.Comments](),
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
	if cached, ok := c.threadCache.get(post.ThreadID); ok {
		return cached, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	replyMap := make(map[string]nostr.Event)

	threadFilters := []nostr.Filter{
		{Kinds: []int{1, 1111}, Tags: nostr.TagMap{"e": {post.ThreadID}}, Limit: c.limit},
		{Kinds: []int{1, 1111}, Tags: nostr.TagMap{"E": {post.ThreadID}}, Limit: c.limit},
	}

	for _, f := range threadFilters {
		events := c.collect(ctx, f)
		for _, evt := range events {
			if evt.ID == post.ThreadID {
				continue
			}
			replyMap[evt.ID] = evt
		}
	}

	thread := buildThread(post.ThreadID, replyMap)

	comments := make([]model.Comment, 0, len(thread))
	for _, evt := range thread {
		comments = append(comments, c.eventToComment(evt.Event, evt.Depth))
	}

	commentsModel := model.Comments{
		PostID:        post.ID,
		PostTitle:     post.PostTitle,
		PostAuthor:    post.Author,
		Community:     post.Community,
		PostText:      post.Content,
		PostUrl:       post.PostUrl,
		PostTimestamp: utils.FriendlyTime(post.CreatedAt),
		Comments:      comments,
		Expiry:        time.Now().Add(10 * time.Minute),
	}

	c.threadCache.set(post.ThreadID, commentsModel, commentsModel.Expiry)
	return commentsModel, nil
}

func (c *NostrClient) GetPostByID(id string) (model.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	events := c.collect(ctx, nostr.Filter{
		IDs:   []string{id},
		Limit: 1,
	})

	if len(events) == 0 {
		return model.Post{}, ErrNotFound
	}

	return c.eventToPost(events[0]), nil
}

func (c *NostrClient) fetchPosts(communities []string, until string, isHome bool) (model.Posts, error) {
	cacheKey := c.postsCacheKey(communities, until, isHome)
	if cached, ok := c.postCache.get(cacheKey); ok {
		return cached, nil
	}

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

	events := c.collect(ctx, filter)

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

	result := model.Posts{
		Description: description,
		Community:   communityLabel,
		IsHome:      isHome,
		Posts:       posts,
		After:       after,
		Expiry:      time.Now().Add(30 * time.Minute),
	}

	c.postCache.set(cacheKey, result, result.Expiry)
	return result, nil
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
	} else {
		community = utils.NormalizeCommunity(community)
	}

	postUrl := fmt.Sprintf("https://nostr.eu/%s", evt.ID)
	if nevent, err := nip19.EncodeEvent(evt.ID, c.relays, evt.PubKey); err == nil {
		postUrl = fmt.Sprintf("https://nostr.eu/%s", nevent)
	}

	return model.Post{
		ID:           evt.ID,
		PostTitle:    title,
		Content:      evt.Content,
		Author:       utils.ShortenPubKey(evt.PubKey),
		PubKey:       evt.PubKey,
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

func (c *NostrClient) PublishPost(community, content string) (model.Post, error) {
	if strings.TrimSpace(content) == "" {
		return model.Post{}, errors.New("content is required")
	}
	if c.privKey == "" {
		return model.Post{}, ErrNoPrivateKey
	}

	normalized := utils.NormalizeCommunity(community)
	if !utils.ValidateTopic(normalized) {
		return model.Post{}, ErrInvalidCommunity
	}

	evt := nostr.Event{
		Kind:    1111,
		Tags:    nostr.Tags{{"I", normalized}},
		Content: strings.TrimSpace(content),
	}

	if err := c.signAndPublish(&evt); err != nil {
		return model.Post{}, err
	}

	c.postCache.clear()
	c.threadCache.clear()

	return c.eventToPost(evt), nil
}

func (c *NostrClient) PublishReply(post model.Post, content string) (model.Comment, error) {
	if strings.TrimSpace(content) == "" {
		return model.Comment{}, errors.New("content is required")
	}
	if c.privKey == "" {
		return model.Comment{}, ErrNoPrivateKey
	}
	if !isValidEventID(post.ThreadID) {
		return model.Comment{}, ErrInvalidThreadID
	}

	tags := nostr.Tags{
		{"e", post.ThreadID},
		{"E", post.ThreadID},
	}

	if utils.ValidateTopic(post.Community) {
		tags = append(tags, nostr.Tag{"I", utils.NormalizeCommunity(post.Community)})
	}

	evt := nostr.Event{
		Kind:    1,
		Tags:    tags,
		Content: strings.TrimSpace(content),
	}

	if err := c.signAndPublish(&evt); err != nil {
		return model.Comment{}, err
	}

	c.threadCache.clear()
	c.postCache.clear()

	return c.eventToComment(evt, 0), nil
}

// EncodeNevent returns a nip19 nevent for the given post.
func (c *NostrClient) EncodeNevent(post model.Post) (string, error) {
	if !isValidEventID(post.ID) {
		return "", ErrInvalidThreadID
	}
	// Prefer relays we already queried so links resolve predictably.
	return nip19.EncodeEvent(post.ID, c.relays, post.PubKey)
}

func (c *NostrClient) signAndPublish(evt *nostr.Event) error {
	if c.privKey == "" {
		return ErrNoPrivateKey
	}

	evt.CreatedAt = nostr.Now()
	evt.PubKey = c.pubKey

	if err := evt.Sign(c.privKey); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	results := c.pool.PublishMany(ctx, c.relays, *evt)
	var (
		success    bool
		errorsSeen []string
	)

	for res := range results {
		if res.Error == nil {
			success = true
			continue
		}
		errorsSeen = append(errorsSeen, fmt.Sprintf("%s: %v", res.RelayURL, res.Error))
	}

	if success {
		return nil
	}

	if len(errorsSeen) > 0 {
		return fmt.Errorf("publish failed: %s", strings.Join(errorsSeen, "; "))
	}

	return errors.New("publish failed")
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

func (c *NostrClient) postsCacheKey(communities []string, cursor string, isHome bool) string {
	ids := append([]string{}, communities...)
	sort.Strings(ids)
	prefix := "community"
	if isHome {
		prefix = "home"
	}
	return fmt.Sprintf("%s:%s:%s:%d", prefix, strings.Join(ids, ","), cursor, c.limit)
}

func parsePrivKey(secret string) (string, string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", "", nil
	}

	if strings.HasPrefix(secret, "nsec") {
		if _, data, err := nip19.Decode(secret); err == nil {
			if sk, ok := data.(string); ok {
				secret = sk
			} else {
				return "", "", errors.New("could not decode nsec key payload")
			}
		} else {
			return "", "", err
		}
	}

	if len(secret) != 64 {
		return "", "", fmt.Errorf("private key must be 64 hex chars or nsec, got %d chars", len(secret))
	}

	if _, err := hex.DecodeString(secret); err != nil {
		return "", "", err
	}

	pubKey, err := nostr.GetPublicKey(secret)
	if err != nil {
		return "", "", err
	}

	return secret, pubKey, nil
}

func (c *NostrClient) collect(ctx context.Context, filter nostr.Filter) []nostr.Event {
	ch := c.pool.FetchMany(ctx, c.relays, filter)
	var events []nostr.Event
	for ev := range ch {
		if ev.Event == nil {
			continue
		}
		events = append(events, *ev.Event)
	}
	return events
}

func isValidEventID(id string) bool {
	if len(id) != 64 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
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
				parentID = p.AsTagReference()
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
