# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DelayedNotifier is a Go web application for sending delayed notifications via email and Telegram. The application uses the Gin web framework (wrapped through the `github.com/wb-go/wbf` package) and provides a REST API for scheduling and managing delayed messages.

## Build and Run Commands

### Building the Application
```bash
go build ./cmd
```

### Running the Application
```bash
go run ./cmd
```
The server starts on the address specified in `.env` file (default: localhost:8080).

### Development Commands
```bash
# Download dependencies
go mod download

# Clean up dependencies
go mod tidy

# Run tests (when implemented)
go test ./...

# Run tests with coverage (when implemented)
go test -cover ./...
```

## Architecture

### Project Structure
- `cmd/main.go` - Application entry point
- `config/` - Configuration management from environment variables
- `internal/` - Internal application packages
  - `models/` - Data models and validation logic
  - `service/` - Business logic services
    - `sender/` - Notification sending services (email, telegram)
    - `timeControl/` - Time-based scheduling logic
  - `web/` - HTTP server and routing
- `tools/` - Utility functions for validation

### Key Components

#### Configuration
- Uses `.env` file for environment variables
- Required variables: `httpHost`, `httpPort`, `staticFilesPath`
- SMTP configuration for email sending (from environment)

#### Web Server
- Built on Gin framework through `github.com/wb-go/wbf/ginext` wrapper
- Serves static files from `internal/web/static`
- Fallback to `index.html` for unmatched routes (SPA support)
- Default debug mode - switch to release mode in production

#### Data Models
- `Record` struct represents notification messages with:
  - ID, data payload, send time
  - Status (waiting/sended/redused)
  - Channel (tg/mail)
  - From/To addresses and subject
- Custom validation for email addresses and Telegram usernames

#### Notification Services
- `EmailSender` - SMTP-based email sending (partially implemented)
- `TgSender` - Telegram bot integration (placeholder)
- `TimeController` - Scheduling logic (placeholder)

### Dependencies
- `github.com/gin-gonic/gin` - Web framework
- `github.com/wb-go/wbf` - WB framework wrapper for Gin
- `github.com/joho/godotenv` - Environment variable loading
- `gopkg.in/gomail.v2` - Email sending
- `github.com/go-playground/validator/v10` - Struct validation

## Development Notes

### Environment Setup
1. Copy and configure `.env` file with required variables
2. Ensure SMTP settings are configured for email functionality
3. Static files should be placed in `internal/web/static/`

### Code Status
- Web server and basic routing are functional
- Model validation is implemented
- Email and Telegram senders are partially implemented (need completion)
- Time scheduling service needs implementation
- No database layer currently present

### Validation Logic
- Custom validators for email addresses and Telegram usernames
- Telegram usernames must start with @, contain 4-32 characters of alphanumeric/underscore
- Email validation uses Go's standard `mail.ParseAddress` function