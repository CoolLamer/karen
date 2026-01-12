package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Discord is a simple Discord webhook notifier.
type Discord struct {
	webhookURL string
	logger     *log.Logger
	client     *http.Client
}

// NewDiscord creates a new Discord notifier. If webhookURL is empty,
// notifications are silently skipped.
func NewDiscord(webhookURL string, logger *log.Logger) *Discord {
	return &Discord{
		webhookURL: webhookURL,
		logger:     logger,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled returns true if the webhook is configured.
func (d *Discord) Enabled() bool {
	return d.webhookURL != ""
}

// discordMessage is the payload for Discord webhook.
type discordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

type discordEmbed struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Color       int           `json:"color,omitempty"`
	Fields      []embedField  `json:"fields,omitempty"`
	Timestamp   string        `json:"timestamp,omitempty"`
}

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// send posts a message to Discord webhook asynchronously.
// Errors are logged but don't affect caller.
func (d *Discord) send(ctx context.Context, msg discordMessage) {
	if !d.Enabled() {
		return
	}

	go func() {
		body, err := json.Marshal(msg)
		if err != nil {
			d.logger.Printf("discord: failed to marshal message: %v", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewReader(body))
		if err != nil {
			d.logger.Printf("discord: failed to create request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := d.client.Do(req)
		if err != nil {
			d.logger.Printf("discord: failed to send webhook: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			d.logger.Printf("discord: webhook returned status %d", resp.StatusCode)
		}
	}()
}

// NotifyNewUser sends a notification when a new user registers.
func (d *Discord) NotifyNewUser(ctx context.Context, phone string) {
	msg := discordMessage{
		Embeds: []discordEmbed{{
			Title:       "Nový uživatel",
			Description: fmt.Sprintf("Registroval se nový uživatel s číslem `%s`", phone),
			Color:       0x00FF00, // Green
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}},
	}
	d.send(ctx, msg)
}

// NotifyPhoneNumbersExhausted sends a notification when phone numbers run out.
func (d *Discord) NotifyPhoneNumbersExhausted(ctx context.Context, tenantID, tenantName string) {
	msg := discordMessage{
		Content: "@here", // Ping everyone
		Embeds: []discordEmbed{{
			Title:       "Došla telefonní čísla!",
			Description: "Nový tenant nedostal přidělené číslo - pool je prázdný.",
			Color:       0xFF0000, // Red
			Fields: []embedField{
				{Name: "Tenant ID", Value: fmt.Sprintf("`%s`", tenantID), Inline: true},
				{Name: "Název", Value: tenantName, Inline: true},
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}},
	}
	d.send(ctx, msg)
}
