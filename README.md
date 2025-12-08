# TUIstr (Nostr topical communities)

A lightweight terminal client for browsing open topical Nostr communities (kind `1111` posts with NIP-73 identifiers) and their NIP-22 comment threads. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and powered by [go-nostr](https://github.com/nbd-wtf/go-nostr). TUIstr is a fork of https://github.com/tonymajestro/reddit-tui. 

Open topical communities working spec: https://github.com/damus-io/dips/pull/2/files

<img width="387" height="134" alt="image" src="https://github.com/user-attachments/assets/d4581ac2-845e-493c-8759-fa5890dc33c0" />

<img width="612" height="434" alt="image" src="https://github.com/user-attachments/assets/32c6aa9b-5b83-4b4c-a20c-f7d3142883b8" />

<img width="567" height="440" alt="image" src="https://github.com/user-attachments/assets/b4e22d9d-77ea-4d49-b167-d00b6d03fdf0" />


## Features
- Featured feed across multiple relays and community identifiers.
- Jump directly to a specific community (`t:`, `u:`, or `g:` NIP-73 ids).
- View comment threads (NIP-22) with nested replies.
- Publish new posts to topic communities and reply to threads (requires a Nostr private key).
- Keyboard-driven navigation (vim-style) and modal search for communities.
- Configurable relays, timeouts, and featured communities via a TOML config.

## Installation

```bash
git clone https://github.com/tonymajestro/TUIstr.git tuistr
cd tuistr
./install.sh
```

To remove the binary:

```bash
./uninstall.sh
```

## Usage

```bash
# Open the featured communities feed
tuistr

# Jump to a specific community (NIP-73 id)
tuistr --community t:linux

# Open a specific event by ID (kind 1111)
tuistr --event <event_id>
```

## Keybindings
- Navigation: `h`, `j`, `k`, `l` or arrow keys
- Jump: `g` (top), `G` (bottom)
- Community search modal: `s`
- New post: `n` (from timelines)
- Load more posts: `L`
- Home: `H`
- Comments: `enter` on a post, `o` to open the event in a browser
- Reply to thread: `r` (while viewing comments)
- Collapse/expand replies: `c` while viewing a thread
- Back: `backspace` / `esc`
- Quit: `q` / `esc`

## Configuration

On first run, a config is created at `~/.config/tuistr/communities.toml`:

```toml
[core]
logLevel = "Warn"

[nostr]
relays = ["wss://relay.damus.io", "wss://nos.lol", "wss://relay.snort.social"]
timeoutSeconds = 10
limit = 50
# Provide a hex or nsec private key to publish posts/replies
# secretKey = ""

[communities]
# NIP-73 identifiers: topics (t:), relays (u:), geohashes (g:)
featured = ["t:nostr", "t:bitcoin", "t:linux"]
default = "t:nostr"
```

- **Featured feed**: Queries kind `1111` events tagged with any `I` value in `communities.featured`.
- **Community page**: Queries kind `1111` events with a root `I` tag matching the selected identifier.
- **Threads**: Fetches NIP-22 replies (kinds `1`/`1111`) referencing the root event (`e/E` tags).
- **Publishing**: Posts are kind `1111` with an `I` tag (topics only for now); replies are kind `1` with `e/E` tags back to the root.

## Notes
- No Reddit APIs or email logins remain—everything is fetched from Nostr relays via go-nostr.
- Kind `1111` post URLs are rendered as `https://nostr.eu/<nevent>` for easy sharing/opening.

## Why aren't you using NIP-29 relay based groups?
[NIP-29](https://github.com/nostr-protocol/nips/blob/master/29.md) is a great foundation to build a moderated discord alternative (see example implementation [flotilla](https://github.com/coracle-social/flotilla)). Open topical communities aims to solve a different problem in public townsquare forums. See a direct comparison below:

| Property | NIP-29 Relay-Based Groups | Open Topical Communities (DIP) |
|---------|----------------------------|--------------------------------|
| **Governance model** | Admin-owned, relay-enforced | Completely ownerless; no admins |
| **Moderation** | Admins can kick, ban, delete posts | No server-side moderation; only client filtering |
| **Structure** | Group metadata + event kinds (`40`, `41`, etc.) | Purely tag-based: NIP-73 + NIP-22 |
| **Posting rules** | Relay controls membership & posting | Anyone can post by using the correct `I` tag |
| **Relay dependence** | High — requires a specific group relay | Optional — events can come from any relay |
| **Censorship resistance** | Medium — admin/relay can censor or remove group | High — no group to censor; content exists across relays |
| **Threading** | Group-specific event structure | Standard NIP-22 threading |
| **Discoverability** | Join group → get group feed | Browse topics/hashtags → feed auto-aggregates |
| **Use-case fit** | Moderated group chats, project teams | Open-topic discovery, public-square conversations |
| **Dealing with spam** | Admin tools can remove content & ban users at the relay level | No centralized control → requires client-side spam filtering, heuristics, muting, reputation systems |
| **UX expectation** | Structured, rule-based group | Freeform Reddit/Twitter-like topic feeds |
| **Implementation complexity for clients** | Medium–high (new event kinds, admin flows, membership UI) | Low (reuse existing NIP-73/NIP-22) |
| **Complexity of implementing a Reddit-style Topic UI** | Not applicable; groups are not topic feeds | Moderate — requires topic directory, trending topics, per-topic timelines, but no new protocol requirements |
| **Community survivability** | Depends on group relay’s longevity | Very high — any relay can host posts; topics are global |
