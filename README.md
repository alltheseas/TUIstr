# TUIstr (Nostr topical communities)

A lightweight terminal client for browsing open topical Nostr communities (kind `1111` posts with NIP-73 identifiers) and their NIP-22 comment threads. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and powered by [go-nostr](https://github.com/nbd-wtf/go-nostr). 

Open topical communities working spec: https://github.com/damus-io/dips/pull/2/files

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
