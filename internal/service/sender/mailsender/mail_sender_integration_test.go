//go:build integration
// +build integration

package mailsend_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	mailsend "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/mailsender"
)

const (
	defaultAPIBase      = "http://localhost:8025"
	defaultMessagesPath = "/api/v2/messages"
	defaultTimeout      = 3 * time.Second

	// SMTP defaults for MailHog
	defaultSMTPHost = "localhost"
	defaultSMTPPort = 1025

	envAPIBase      = "MAILHOG_API_URL"       // base URL, e.g. http://localhost:8025
	envMessagesPath = "MAILHOG_MESSAGES_PATH" // optional override, e.g. /api/v2/messages
	envTimeout      = "MAILHOG_READY_TIMEOUT" // optional, duration like 3s

	// SMTP envs (для будущего теста отправки)
	envSMTPHost = "MAILHOG_SMTP_HOST" // optional, default localhost
	envSMTPPort = "MAILHOG_SMTP_PORT" // optional, default 1025
)

type mailhogCfg struct {
	apiBase      string
	messagesPath string
	endpoint     string
	timeout      time.Duration

	smtpHost string
	smtpPort int
}

func loadMailhogCfg(t *testing.T) mailhogCfg {
	t.Helper()

	// --- API base ---
	apiBase := strings.TrimSpace(os.Getenv(envAPIBase))
	if apiBase == "" {
		apiBase = defaultAPIBase
	}
	apiBase = strings.TrimRight(apiBase, "/")

	// --- messages path ---
	path := strings.TrimSpace(os.Getenv(envMessagesPath))
	if path == "" {
		path = defaultMessagesPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// --- timeout ---
	timeout := defaultTimeout
	if v := strings.TrimSpace(os.Getenv(envTimeout)); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("invalid %s=%q: %v", envTimeout, v, err)
		}
		if d <= 0 {
			t.Fatalf("invalid %s=%q: duration must be > 0", envTimeout, v)
		}
		timeout = d
	}

	// --- SMTP host ---
	smtpHost := strings.TrimSpace(os.Getenv(envSMTPHost))
	if smtpHost == "" {
		smtpHost = defaultSMTPHost
	}

	// --- SMTP port ---
	smtpPort := defaultSMTPPort
	if v := strings.TrimSpace(os.Getenv(envSMTPPort)); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("invalid %s=%q: %v", envSMTPPort, v, err)
		}
		if p <= 0 || p > 65535 {
			t.Fatalf("invalid %s=%q: port out of range", envSMTPPort, v)
		}
		smtpPort = p
	}

	return mailhogCfg{
		apiBase:      apiBase,
		messagesPath: path,
		endpoint:     apiBase + path,
		timeout:      timeout,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
	}
}

func TestMailhogReady(t *testing.T) {
	cfg := loadMailhogCfg(t)
	requireMailhogReady(t, cfg)
}

func requireMailhogReady(t *testing.T, cfg mailhogCfg) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.endpoint, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := (&http.Client{Timeout: cfg.timeout}).Do(req)
	if err != nil {
		t.Skipf("MailHog is not reachable (%s). Start it with docker and rerun. err=%v", cfg.endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("MailHog API not ready: GET %s returned %d (expected %d)", cfg.endpoint, resp.StatusCode, http.StatusOK)
	}
}

func sendTestEmail(t *testing.T, cfg mailhogCfg) (subject string, body string) {
	t.Helper()

	subject = "it-" + t.Name() + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	body = "integration test body: " + subject

	cfgMail := config.EmailSenderConfig{
		SMTPHost:     cfg.smtpHost,
		SMTPPort:     strconv.Itoa(cfg.smtpPort),
		SMTPUser:     "no-reply@example.com",
		SMTPPassword: "",
	}

	sender, err := mailsend.NewEmailSender(cfgMail)
	if err != nil {
		t.Fatalf("new email sender: %v", err)
	}

	record := &models.Record{
		SendChan: "mail",
		From:     "", // возьмётся из dialer.Username (SMTPUser)
		To:       []string{"test@example.com"},
		Subject:  subject,
		Data:     []byte(body),
	}

	if err := sender.SendMessage(record); err != nil {
		t.Fatalf("send message via %s:%d: %v", cfg.smtpHost, cfg.smtpPort, err)
	}

	return subject, body
}

func TestEmailSender_SendMessage_Integration(t *testing.T) {
	cfg := loadMailhogCfg(t)
	requireMailhogReady(t, cfg)

	testMessageNum := 3

	// expected[subject] = body
	expected := make(map[string]string, testMessageNum)

	for i := 0; i < testMessageNum; i++ {
		subject, body := sendTestEmail(t, cfg)
		expected[subject] = body
	}

	mhMessages, err := fetchMessages(t, cfg)
	if err != nil {
		t.Fatalf("fetch messages from mailhog: %v", err)
	}

	if len(mhMessages) < testMessageNum {
		t.Fatalf("expected at least %d messages in MailHog, got %d", testMessageNum, len(mhMessages))
	}

	found := make(map[string]bool, testMessageNum)

	for _, msg := range mhMessages {
		subjects, ok := msg.Content.Headers["Subject"]
		if !ok || len(subjects) == 0 { // пропускаем пустые сообщения
			continue
		}

		subj := subjects[0] // тк subjects это []stirng
		expectedBody, isOurMessage := expected[subj]
		if !isOurMessage {
			continue // это не наше письмо (например старое)
		}

		// Проверка To (по SMTP envelope)
		if len(msg.To) == 0 {
			t.Errorf("message %q has empty To", subj)
			continue
		}
		gotTo := msg.To[0].Mailbox + "@" + msg.To[0].Domain
		if gotTo != "test@example.com" {
			t.Errorf("message %q: unexpected To: got %q want %q", subj, gotTo, "test@example.com")
		}

		// Проверка Subject
		if subj != "" && subj != subjects[0] {
			t.Errorf("message subject mismatch: got %q want %q", subjects[0], subj)
		}

		// Проверяем корректность тестовых данных (expectedBody),
		// чтобы при изменении генерации body ошибка была понятнее.
		if !strings.Contains(expectedBody, "integration test body:") {
			t.Fatalf("test bug: expected body has unexpected format: %q", expectedBody)
		}

		// Проверка Body:
		// MailHog может хранить quoted-printable с переносами (=\\r\\n),
		// поэтому точное == ненадёжно. Проверяем устойчивый фрагмент.
		if !strings.Contains(msg.Content.Body, "integration test body:") {
			t.Errorf("message %q body does not contain expected prefix; body=%q", subj, msg.Content.Body)
		}

		found[subj] = true
	}

	// Убедимся, что нашли все отправленные нами письма
	for subj := range expected {
		if !found[subj] {
			t.Errorf("sent message not found in MailHog: subject=%q", subj)
		}
	}
}

// это респонс
type mailhogMessagesResponse struct {
	Total int64            `json:"total"`
	Count int64            `json:"count"`
	Start int64            `json:"start"`
	Items []mailhogMessage `json:"items"`
}

// это список сообщений
type mailhogMessage struct {
	ID      string         `json:"ID"`
	From    mailhogPath    `json:"From"`
	To      []mailhogPath  `json:"To"`
	Content mailhogContent `json:"Content"`
	Created time.Time      `json:"Created"`
}

type mailhogPath struct {
	Mailbox string `json:"Mailbox"`
	Domain  string `json:"Domain"`
}

// это типо payload каждого сообщения
type mailhogContent struct {
	Headers map[string][]string `json:"Headers"`
	Body    string              `json:"Body"`
}

func fetchMessages(t *testing.T, cfg mailhogCfg) ([]mailhogMessage, error) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, cfg.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request error: %w", err)
	}
	resp, err := (&http.Client{Timeout: cfg.timeout}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch messages error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response mailhogMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return response.Items, nil
}
