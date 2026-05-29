package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mymmrac/telego"
)

// BotService wraps telego client and handles message delivery with logging.
type BotService struct {
	client telego.Bot
	logger *slog.Logger
}

// MessageConfig bundles all parameters needed for a broadcast.
type MessageConfig struct {
	Message   string
	ParseMode string
	MediaFile string
	MediaType string
}

// NewBotService creates a BotService using the provided bot token.
// Returns an error if the bot cannot initialize (e.g., invalid token).
func NewBotService(token string, logger *slog.Logger) (*BotService, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, fmt.Errorf("init telego bot: %w", err)
	}
	return &BotService{client: *bot, logger: logger}, nil
}

// SendResult captures the outcome of a single message delivery attempt.
type SendResult struct {
	GroupID   int64
	GroupName string
	Success   bool
	Error     string
	SentAt    time.Time
}

// Broadcast sends a message to all provided groups.
// If MediaFile is set, sends media with the message as caption.
// Logs every attempt and returns a slice of results for external inspection.
// Processing continues even if individual sends fail.
func (s *BotService) Broadcast(groups []Group, cfg MessageConfig) []SendResult {
	results := make([]SendResult, 0, len(groups))

	for _, group := range groups {
		result := s.sendToGroup(group, cfg)
		results = append(results, result)
	}

	s.logSummary(results)
	return results
}

func (s *BotService) sendToGroup(group Group, cfg MessageConfig) SendResult {
	result := SendResult{GroupID: group.ID, GroupName: group.Name, SentAt: time.Now().UTC()}

	if cfg.MediaFile != "" {
		return s.sendMediaToGroup(group, cfg, result)
	}
	return s.sendTextToGroup(group, cfg, result)
}

func (s *BotService) sendTextToGroup(group Group, cfg MessageConfig, result SendResult) SendResult {
	params := telego.SendMessageParams{
		ChatID: telego.ChatID{ID: group.ID},
		Text:   cfg.Message,
	}
	applyParseMode(&params.ParseMode, cfg.ParseMode)

	groupLabel := groupLogLabel(group)

	_, err := s.client.SendMessage(context.Background(), &params)
	if err != nil {
		result.Error = err.Error()
		s.logger.Error("send failed",
			groupLabel,
			slog.String("error", result.Error),
		)
		return result
	}

	result.Success = true
	s.logger.Info("send succeeded",
		groupLabel,
		slog.Time("sent_at", result.SentAt),
	)
	return result
}

func (s *BotService) sendMediaToGroup(group Group, cfg MessageConfig, result SendResult) SendResult {
	inputFile, fileHandle, err := resolveInputFile(cfg.MediaFile)
	if err != nil {
		result.Error = err.Error()
		s.logger.Error("send failed",
			groupLogLabel(group),
			slog.String("error", result.Error),
		)
		return result
	}
	if fileHandle != nil {
		defer fileHandle.Close()
	}

	groupLabel := groupLogLabel(group)
	mediaType := cfg.MediaType
	if mediaType == "" {
		mediaType = "photo"
	}

	var sendErr error
	switch mediaType {
	case "photo":
		params := telego.SendPhotoParams{
			ChatID:  telego.ChatID{ID: group.ID},
			Photo:   inputFile,
			Caption: cfg.Message,
		}
		applyParseMode(&params.ParseMode, cfg.ParseMode)
		_, sendErr = s.client.SendPhoto(context.Background(), &params)
	case "video":
		params := telego.SendVideoParams{
			ChatID:  telego.ChatID{ID: group.ID},
			Video:   inputFile,
			Caption: cfg.Message,
		}
		applyParseMode(&params.ParseMode, cfg.ParseMode)
		_, sendErr = s.client.SendVideo(context.Background(), &params)
	case "document":
		params := telego.SendDocumentParams{
			ChatID:   telego.ChatID{ID: group.ID},
			Document: inputFile,
			Caption:  cfg.Message,
		}
		applyParseMode(&params.ParseMode, cfg.ParseMode)
		_, sendErr = s.client.SendDocument(context.Background(), &params)
	default:
		sendErr = fmt.Errorf("unsupported media_type: %s", mediaType)
	}

	if sendErr != nil {
		result.Error = sendErr.Error()
		s.logger.Error("send failed",
			groupLabel,
			slog.String("media_type", mediaType),
			slog.String("error", result.Error),
		)
		return result
	}

	result.Success = true
	s.logger.Info("send succeeded",
		groupLabel,
		slog.String("media_type", mediaType),
		slog.Time("sent_at", result.SentAt),
	)
	return result
}

func resolveInputFile(source string) (telego.InputFile, *os.File, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return telego.InputFile{URL: source}, nil, nil
	}

	file, err := os.Open(source)
	if err == nil {
		return telego.InputFile{File: file}, file, nil
	}

	if os.IsNotExist(err) {
		return telego.InputFile{FileID: source}, nil, nil
	}

	return telego.InputFile{}, nil, fmt.Errorf("open media file: %w", err)
}

func applyParseMode(target *string, parseMode string) {
	switch parseMode {
	case "Markdown":
		*target = telego.ModeMarkdown
	case "MarkdownV2":
		*target = telego.ModeMarkdownV2
	}
}

func groupLogLabel(group Group) slog.Attr {
	if group.Name != "" {
		return slog.String("group", fmt.Sprintf("%s (%d)", group.Name, group.ID))
	}
	return slog.Int64("group_id", group.ID)
}

func (s *BotService) logSummary(results []SendResult) {
	var successCount, failCount int
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	s.logger.Info("broadcast complete",
		slog.Int("total", len(results)),
		slog.Int("succeeded", successCount),
		slog.Int("failed", failCount),
	)
}

func (s *BotService) FetchGroupsFromUpdates() ([]Group, error) {
	updates, err := s.client.GetUpdates(context.Background(), &telego.GetUpdatesParams{})
	if err != nil {
		return nil, fmt.Errorf("fetch updates: %w", err)
	}

	seen := make(map[int64]bool)
	var groups []Group

	for _, update := range updates {
		if update.Message == nil {
			continue
		}

		chat := update.Message.Chat
		if chat.Type != "group" && chat.Type != "supergroup" && chat.Type != "channel" {
			continue
		}

		if seen[chat.ID] {
			continue
		}
		seen[chat.ID] = true

		name := chat.Title
		if name == "" {
			name = fmt.Sprintf("Chat %d", chat.ID)
		}
		groups = append(groups, Group{ID: chat.ID, Name: name})
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups found in recent updates")
	}

	s.logger.Info("groups discovered", slog.Int("count", len(groups)))
	return groups, nil
}

func newLogger() *slog.Logger {
	logFile, err := os.OpenFile("sf.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	session := time.Now().UTC().Format(time.RFC3339)
	separator := strings.Repeat("=", 60)
	header := fmt.Sprintf("\n%s\nSession: %s\n%s\n", separator, session, separator)
	logFile.WriteString(header)

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	return slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
