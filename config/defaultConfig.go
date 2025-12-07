package config

const defaultConfiguration = `
#
# Default configuration for the Nostr communities TUI.
# Uncomment to customize.

[core]
#logLevel = "Warn"

[nostr]
#relays = ["wss://relay.damus.io", "wss://nos.lol", "wss://relay.snort.social"]
#timeoutSeconds = 10
#limit = 50
#secretKey = ""  # nsec or hex, required to publish

[communities]
#featured = ["t:nostr", "t:farmstr", "t:foodstr"]
#default = ""  # leave empty to start on the featured feed
`
