# Telegram Broadcast Bot

A simple, cross-platform Go application that broadcasts messages to multiple Telegram groups using the [telego](https://github.com/mymmrac/telego) library. Configuration is externalized to a JSON file — no hardcoded tokens or group IDs.

---

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Building](#building)
- [Running](#running)
- [Error Handling & Logging](#error-handling--logging)
- [Project Structure](#project-structure)
- [Architecture](#architecture)

---

## Features

- **JSON-based configuration** — bot token, group IDs, message text, and parse mode in one file
- **Markdown support** — send formatted messages using Telegram Markdown or MarkdownV2
- **Per-group error tracking** — knows exactly which groups failed and why
- **Structured logging** — every send attempt and a final summary are logged
- **Cross-platform builds** — compile for Linux and Windows from Ubuntu without extra tools
- **Graceful failure handling** — continues broadcasting to remaining groups even if one fails
- **Non-zero exit codes** — returns `1` when any group fails, making it easy to detect issues in scripts or CI/CD

---

## Prerequisites

- **Go 1.21+** (for `log/slog` support)
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- Group IDs where your bot is a member

---

## Installation

```bash
# Clone or create the project directory
cd sf

# Download dependencies
go mod tidy
```

---

## Configuration

Create a `config.json` file in the project root (or copy from the example):

```bash
cp config.example.json config.json
```

### Config Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `bot_token` | string | Yes | Your Telegram bot token from @BotFather |
| `groups` | array of objects | No | List of target groups. When empty or omitted, the bot auto-discovers groups from recent Telegram updates |
| `message` | string | Yes | The message to broadcast. When `media_file` is set, this becomes the media caption |
| `parse_mode` | string | No | `"MarkdownV2"` or `"Markdown"`. Defaults to `"MarkdownV2"` if empty |
| `media_file` | string | No | Path, URL, or Telegram file ID of media to send. When set, `message` becomes the caption |
| `media_type` | string | No | `"photo"` or `"document"`. Defaults to `"photo"` if empty. Only used when `media_file` is set |

### Example `config.json`

Copy from `config.example.json` and fill in your values:

```json
{
  "bot_token": "YOUR_BOT_TOKEN_HERE",
  "message": "Hello *world* this is **bold** and `code`",
  "parse_mode": "MarkdownV2",
  "groups": [
    {
      "id": -1001234567890,
      "name": "Engineering Team"
    }
  ]
}
```

**Two modes, one config file:**

- **Manual mode:** Include `groups` array to send to specific groups
- **Auto-discovery:** Remove `groups` field entirely to discover groups from recent Telegram updates

> **Note:** The `name` field is optional but recommended. It appears in logs so you can identify which group succeeded or failed without memorizing IDs.

### Why Add Group Names?

Group IDs like `-1001234567890` are meaningless at a glance. Adding a `name` makes logs self-documenting:

**Without names (hard to read):**
```
send failed  group_id=-1001234567890  error="Forbidden..."
send failed  group_id=-1009876543210  error="Forbidden..."
```

**With names (immediately clear):**
```
send failed  group="Engineering Team (-1001234567890)"  error="Forbidden..."
send failed  group="Marketing Alerts (-1009876543210)"  error="Forbidden..."
```

The name is purely decorative — it does not affect Telegram delivery. Leave it empty or omit it entirely if you prefer.

### Sending Media

You can optionally attach a media file (image, video, or document) to your broadcast. When `media_file` is set, the `message` text becomes the media caption instead of a standalone text message.

**Supported media types:**

| Type | `media_type` | Supported Formats | Notes |
|---|---|---|---|
| **Image** | `"photo"` | jpg, png, gif | Best for photos and screenshots |
| **Video** | `"video"` | mp4, mov | Best for video clips |
| **Document** | `"document"` | pdf, zip, any | Best for files that shouldn't be compressed |

**Supported file sources:**

| Source | Example | Notes |
|---|---|---|
| **Local file** | `"image.jpg"` | Relative or absolute path on disk |
| **URL** | `"https://example.com/image.jpg"` | Telegram downloads the file |
| **Telegram file ID** | `"AgADBAAD..."` | Reuses an already-uploaded file (fastest) |

**Example with image:**

```json
{
  "bot_token": "123456:ABC-DEF...",
  "groups": [{"id": -1001234567890, "name": "Engineering Team"}],
  "message": "Check out this **awesome** photo!",
  "parse_mode": "MarkdownV2",
  "media_file": "assets/screenshot.png",
  "media_type": "photo"
}
```

**Example with video:**

```json
{
  "media_file": "assets/demo.mp4",
  "media_type": "video"
}
```

**Example with document:**

```json
{
  "media_file": "assets/report.pdf",
  "media_type": "document"
}
```

**Example with URL:**

```json
{
  "media_file": "https://example.com/banner.jpg",
  "media_type": "photo"
}
```

> **Note:** `media_type` defaults to `"photo"`. Always specify `"video"` for video files and `"document"` for PDFs/archives.

### How to Get Your Group ID

1. Add your bot to the group
2. Send a message in the group
3. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for `"chat":{"id":-100...` — that number is your group ID

> **Note:** Group IDs for supergroups and channels always start with `-100`.

---

## Building

### Native Build (Current OS)

```bash
make build
# Creates: sf
```

### Cross-Compilation

```bash
# Linux AMD64
make build-linux
# Creates: sf-linux

# Windows AMD64
make build-windows
# Creates: sf.exe
```

### Manual Build (Without Make)

```bash
# Native
go build -o sf .

# Linux
go GOOS=linux GOARCH=amd64 build -o sf-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o sf.exe .
```

### Clean Build Artifacts

```bash
make clean
```

---

## Running

### Using Make

```bash
# Uses config.json in current directory
make run
```

### Using Go Directly

```bash
# Default config path
go run .

# Custom config path
go run . -config=/path/to/config.json

# Using the compiled binary
./sf -config=/path/to/config.json
```

### Command-Line Flags

| Flag | Default | Description |
|---|---|---|
| `-config` | `config.json` | Path to the JSON configuration file |

### Automatic Group Discovery

When `groups` is empty or omitted from `config.json`, the bot automatically discovers target groups from recent Telegram updates:

```bash
./sf -config=config.json
```

With this `config.json`:

```json
{
  "bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
  "message": "Hello *world* this is **bold** and `code`",
  "parse_mode": "MarkdownV2"
}
```

**How it works:**
1. The bot fetches recent updates via the Telegram Bot API (`getUpdates`)
2. Extracts unique group/supergroup/channel IDs where the bot has received messages
3. Deduplicates them automatically
4. Sends the message once per discovered group

**Requirements:**
- The bot must have received at least one message in each target group recently
- Groups without recent activity won't appear
- You can verify available updates manually: `https://api.telegram.org/bot<TOKEN>/getUpdates`

**When to use manual groups:**
- You need precise control over which groups receive messages
- You want human-readable names in logs
- You need to send to groups without recent updates

---

## Error Handling & Logging

The application uses Go's structured logging (`log/slog`) to provide clear output.

### Success Log Example

When a group has a `name` in config:
```
time=2024-05-21T10:30:00.000Z level=INFO msg="send succeeded" group="Engineering Team (-1001234567890)" sent_at=2024-05-21T10:30:00.000Z
time=2024-05-21T10:30:01.000Z level=INFO msg="send succeeded" group="Marketing Alerts (-1009876543210)" sent_at=2024-05-21T10:30:01.000Z
time=2024-05-21T10:30:01.000Z level=INFO msg="broadcast complete" total=2 succeeded=2 failed=0
```

When a group has no `name`:
```
time=2024-05-21T10:30:00.000Z level=INFO msg="send succeeded" group_id=-1001234567890 sent_at=2024-05-21T10:30:00.000Z
```

### Media Log Example

When `media_file` is set, logs include the media type:
```
time=2024-05-21T10:30:00.000Z level=INFO msg="send succeeded" group="Engineering Team (-1001234567890)" media_type=photo sent_at=2024-05-21T10:30:00.000Z
time=2024-05-21T10:30:01.000Z level=ERROR msg="send failed" group="Marketing Alerts (-1009876543210)" media_type=document error="Bad Request: file is too big"
time=2024-05-21T10:30:01.000Z level=INFO msg="broadcast complete" total=2 succeeded=1 failed=1
```

### Failure Log Example

```
time=2024-05-21T10:30:00.000Z level=ERROR msg="send failed" group="Engineering Team (-1001234567890)" error="Forbidden: bot was kicked from the group chat"
time=2024-05-21T10:30:01.000Z level=INFO msg="send succeeded" group="Marketing Alerts (-1009876543210)" sent_at=2024-05-21T10:30:01.000Z
time=2024-05-21T10:30:01.000Z level=INFO msg="broadcast complete" total=2 succeeded=1 failed=1
```

### Common Errors

| Error | Meaning | Action |
|---|---|---|
| `Forbidden: bot was kicked from the group chat` | Bot was removed/banned from the group | Re-add the bot or remove the group ID from config |
| `Bad Request: chat not found` | Group no longer exists or ID is wrong | Verify the group ID |
| `Forbidden: bot is not a member of the group chat` | Bot was never added or was removed | Add the bot to the group |
| `Bad Request: can't parse entities` | Markdown formatting is invalid | Check your message for unescaped characters |
| `Bad Request: file is too big` | Media file exceeds Telegram limits (10 MB for photos, 50 MB for documents/videos) | Compress the file or use a smaller version |
| `resolve media file: ...` | Failed to open local media file | Check the file path exists and is readable |

### Exit Codes

| Code | Meaning |
|---|---|
| `0` | All messages sent successfully |
| `1` | Config error, bot initialization failed, or one/more groups failed |

### Log File Output

In addition to stdout, every run appends logs to `sf.log` in the current directory with a session separator:

```
============================================================
Session: 2024-05-21T10:30:00Z
============================================================
time=2024-05-21T10:30:00.000Z level=INFO msg="send succeeded" group="Engineering Team (-1001234567890)" sent_at=2024-05-21T10:30:00.000Z
time=2024-05-21T10:30:01.000Z level=INFO msg="broadcast complete" total=2 succeeded=2 failed=0

============================================================
Session: 2024-05-21T11:15:00Z
============================================================
time=2024-05-21T11:15:00.000Z level=ERROR msg="send failed" group="Marketing Alerts (-1009876543210)" error="Forbidden: bot was kicked from the group chat"
```

> **Note:** The file is created automatically if it doesn't exist. Old sessions are never overwritten — logs accumulate for audit purposes.

---

## Project Structure

```
.
├── config.go           # JSON config loader with validation
├── bot.go              # Telegram bot service (telego wrapper + logging)
├── main.go             # Entry point, flag parsing, orchestration
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation
├── config.example.json # Example configuration
└── README.md           # This file
```

---

## Architecture

Built with **SOLID**, **KISS**, **DRY**, and **YAGNI** principles:

- **Single Responsibility** — each file does one thing: config loads, bot sends, main orchestrates
- **Interface Segregation** — `Messenger` interface decouples business logic from telego implementation
- **No over-engineering** — no database, no web server, no env file parser. JSON config is sufficient for this scope
- **Fail-fast validation** — config is validated immediately; bot token is checked at initialization
- **Graceful degradation** — one group failure does not block the rest

### Data Flow

```
config.json → LoadConfig() → validate() → NewBotService() → Broadcast()
                                                              ↓
                                                    ┌─────────┴─────────┐
                                                    ↓                   ↓
                                             sendToGroup()          sendToGroup()
                                                    ↓                   ↓
                                             log success/error    log success/error
                                                    ↓                   ↓
                                             ┌──────┴──────┐
                                             ↓             ↓
                                        logSummary()   return results
```

---

## Markdown Formatting Reference

When using `parse_mode: MarkdownV2`, escape these characters with a backslash: `_`, `*`, `[`, `]`, `(`, `)`, `~`, `` ` ``, `>`, `#`, `+`, `-`, `=`, `|`, `{`, `}`, `.`, `!`.

| Format | Syntax | Example |
|---|---|---|
| Bold | `**text**` | `**bold**` |
| Italic | `__text__` | `__italic__` |
| Code inline | `` `code` `` | `` `inline code` `` |
| Code block | ` ```code``` ` | ` ```block``` ` |
| Strikethrough | `~~text~~` | `~~strikethrough~~` |
| Link | `[text](URL)` | `[Google](https://google.com)` |

---

## License

This is a personal starter project. Use and modify as needed.
