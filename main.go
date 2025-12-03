package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"reddittui/components"
	"reddittui/config"
	"reddittui/utils"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "v0.3.9"

type CliArgs struct {
	community   string
	postId      string
	showVersion bool
}

func main() {
	configuration, _ := config.LoadConfig()

	logFile, err := utils.InitLogger(configuration.Core.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open logfile: %v\n", err)
	}

	defer logFile.Close()

	var args CliArgs
	flag.StringVar(&args.postId, "event", "", "Event id")
	flag.StringVar(&args.community, "community", "", "Community identifier (NIP-73)")
	flag.BoolVar(&args.showVersion, "version", false, "Version")
	flag.Parse()

	if args.showVersion {
		fmt.Printf("communities-tui version %s\n", version)
		os.Exit(0)
	}

	communities, err := components.NewCommunitiesTui(configuration, args.community, args.postId)
	if err != nil {
		slog.Error("Error initializing communities tui", "error", err)
		os.Exit(1)
	}

	p := tea.NewProgram(communities, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		slog.Error("Error running communities tui, see logfile for details", "error", err)
		os.Exit(1)
	}
}
