package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"tuistr/utils"

	"github.com/BurntSushi/toml"
)

const (
	configFilename = "communities.toml"
)

type Config struct {
	Core        CoreConfig        `toml:"core"`
	Nostr       NostrConfig       `toml:"nostr"`
	Communities CommunitiesConfig `toml:"communities"`
}

type CoreConfig struct {
	LogLevel string
}

type NostrConfig struct {
	Relays         []string
	TimeoutSeconds int
	Limit          int
	SecretKey      string
}

type CommunitiesConfig struct {
	Featured []string
	Default  string
}

func NewConfig() Config {
	return Config{
		Core: CoreConfig{
			LogLevel: "Warn",
		},
		Nostr: NostrConfig{
			Relays:         []string{"wss://relay.damus.io", "wss://nos.lol", "wss://relay.snort.social"},
			TimeoutSeconds: 10,
			Limit:          50,
			SecretKey:      "",
		},
		Communities: CommunitiesConfig{
			Featured: []string{"t:nostr", "t:farmstr", "t:foodstr"},
			Default:  "",
		},
	}
}

func LoadConfig() (Config, error) {
	defaultConfig := NewConfig()

	configDir, err := utils.GetConfigDir()
	if err != nil {
		slog.Warn("Could not get config directory", "error", err)
		return defaultConfig, err
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		slog.Warn("Could not make config directory", "error", err)
		return defaultConfig, err
	}

	configPath := filepath.Join(configDir, configFilename)
	configFile, err := os.Open(configPath)
	if os.IsNotExist(err) {
		createConfigFile(configPath)
		return defaultConfig, err
	} else if err != nil {
		slog.Warn("Could not open config file", "error", err)
		return defaultConfig, err
	}

	defer configFile.Close()

	var configFromFile Config
	decoder := toml.NewDecoder(configFile)
	meta, err := decoder.Decode(&configFromFile)
	if err != nil {
		slog.Warn("Could not decode config file", "error", err)
		return defaultConfig, err
	}

	mergedConfig := mergeConfig(defaultConfig, configFromFile, meta)
	return mergedConfig, err
}

// Merge right config into left
func mergeConfig(left, right Config, meta toml.MetaData) Config {
	if meta.IsDefined("core", "logLevel") {
		left.Core.LogLevel = right.Core.LogLevel
	}

	if meta.IsDefined("nostr", "relays") {
		left.Nostr.Relays = right.Nostr.Relays
	}

	if meta.IsDefined("nostr", "timeoutSeconds") {
		left.Nostr.TimeoutSeconds = right.Nostr.TimeoutSeconds
	}

	if meta.IsDefined("nostr", "limit") {
		left.Nostr.Limit = right.Nostr.Limit
	}

	if meta.IsDefined("nostr", "secretKey") {
		left.Nostr.SecretKey = right.Nostr.SecretKey
	}

	if meta.IsDefined("communities", "featured") {
		left.Communities.Featured = right.Communities.Featured
	}

	if meta.IsDefined("communities", "default") {
		left.Communities.Default = right.Communities.Default
	}

	return left
}

func createConfigFile(configFilePath string) error {
	configFile, err := os.Create(configFilePath)
	if err != nil {
		return err
	}

	_, err = configFile.WriteString(defaultConfiguration)
	return err
}
