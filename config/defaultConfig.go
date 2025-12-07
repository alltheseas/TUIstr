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

[communities]
#featured = ["t:nostr", "t:bitcoin", "t:linux"]
#default = "t:nostr"
`
