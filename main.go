package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to JSON configuration file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger()

	bot, err := NewBotService(cfg.BotToken, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bot init error: %v\n", err)
		os.Exit(1)
	}

	groups := cfg.Groups
	if len(groups) == 0 {
		groups, err = bot.FetchGroupsFromUpdates()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch groups error: %v\n", err)
			os.Exit(1)
		}
	}

	msgCfg := MessageConfig{
		Message:   cfg.Message,
		ParseMode: cfg.ParseMode,
		MediaFile: cfg.MediaFile,
		MediaType: cfg.MediaType,
	}

	results := bot.Broadcast(groups, msgCfg)

	if hasFailures(results) {
		os.Exit(1)
	}
}

func hasFailures(results []SendResult) bool {
	for _, r := range results {
		if !r.Success {
			return true
		}
	}
	return false
}
