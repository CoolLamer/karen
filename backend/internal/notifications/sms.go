package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SMSConfig holds configuration for SMS notifications via Twilio
type SMSConfig struct {
	AccountSID   string // Twilio Account SID
	AuthToken    string // Twilio Auth Token
	SenderNumber string // Twilio phone number to send from (E.164 format) - optional, can be set later via SetSenderNumber
}

// SMSClient sends SMS notifications via Twilio Programmable Messaging
type SMSClient struct {
	accountSID   string
	authToken    string
	senderNumber string
	logger       *log.Logger
	mu           sync.Mutex
}

// NewSMSClient creates a new SMS client for sending notifications.
// SenderNumber is optional at initialization - it can be set later via SetSenderNumber
// or fetched from global_config by the caller.
func NewSMSClient(cfg SMSConfig, logger *log.Logger) (*SMSClient, error) {
	if cfg.AccountSID == "" || cfg.AuthToken == "" {
		logger.Println("SMS: missing Twilio credentials, SMS notifications disabled")
		return nil, nil
	}

	if cfg.SenderNumber != "" {
		logger.Printf("SMS: client initialized (sender=%s)", cfg.SenderNumber)
	} else {
		logger.Println("SMS: client initialized (sender number will be fetched from config)")
	}

	return &SMSClient{
		accountSID:   cfg.AccountSID,
		authToken:    cfg.AuthToken,
		senderNumber: cfg.SenderNumber,
		logger:       logger,
	}, nil
}

// SetSenderNumber updates the sender phone number used for outgoing SMS.
// This allows dynamic configuration from global_config.
func (c *SMSClient) SetSenderNumber(number string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.senderNumber = number
}

// GetSenderNumber returns the current sender phone number.
func (c *SMSClient) GetSenderNumber() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.senderNumber
}

// twilioMessageResponse represents a Twilio Messages API response
type twilioMessageResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SendSMS sends an SMS message to the specified phone number
func (c *SMSClient) SendSMS(ctx context.Context, to, body string) error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.senderNumber == "" {
		c.logger.Println("SMS: cannot send - sender number not configured (set sms_sender_number in global_config)")
		return fmt.Errorf("SMS sender number not configured")
	}

	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		c.accountSID,
	)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", c.senderNumber)
	data.Set("Body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Printf("SMS: failed to send to %s: %v", to, err)
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	var msgResp twilioMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Printf("SMS: Twilio error (code=%d, msg=%s)", msgResp.ErrorCode, msgResp.ErrorMessage)
		return fmt.Errorf("Twilio API error: %d - %s", msgResp.ErrorCode, msgResp.ErrorMessage)
	}

	c.logger.Printf("SMS: sent successfully to %s (sid=%s, status=%s)", to, msgResp.SID, msgResp.Status)
	return nil
}

// TrialNotificationType represents the type of trial notification
type TrialNotificationType string

const (
	TrialNotificationDay10         TrialNotificationType = "day10"
	TrialNotificationDay12         TrialNotificationType = "day12"
	TrialNotificationDay14         TrialNotificationType = "day14"
	TrialNotificationGraceWarning  TrialNotificationType = "grace_warning"
	TrialNotificationPhoneReleased TrialNotificationType = "phone_released"
)

// SendTrialDay10Notification sends the Day 10 trial reminder (4 days left)
func (c *SMSClient) SendTrialDay10Notification(ctx context.Context, to string, timeSavedMinutes int) error {
	body := fmt.Sprintf("Zvednu: Zbývají ti 4 dny trialu. Karen ti zatím ušetřila %d minut. Upgraduj na zvednu.cz", timeSavedMinutes)
	return c.SendSMS(ctx, to, body)
}

// SendTrialDay12Notification sends the Day 12 trial reminder (2 days left)
func (c *SMSClient) SendTrialDay12Notification(ctx context.Context, to string, callsHandled int) error {
	body := fmt.Sprintf("Zvednu: Zbývají ti 2 dny trialu. Karen ti vyřídila %d hovorů. Upgraduj na zvednu.cz", callsHandled)
	return c.SendSMS(ctx, to, body)
}

// SendTrialExpiredNotification sends the Day 14 trial expired notification
func (c *SMSClient) SendTrialExpiredNotification(ctx context.Context, to string) error {
	body := "Zvednu: Trial skončil. Karen nebude přijímat hovory. Upgraduj na zvednu.cz"
	return c.SendSMS(ctx, to, body)
}

// SendTrialGraceWarningNotification sends a warning that phone number will be released
func (c *SMSClient) SendTrialGraceWarningNotification(ctx context.Context, to, assignedNumber string, daysUntilRelease int) error {
	body := fmt.Sprintf("Zvednu: Za %d dní bude vaše číslo %s odpojeno. Zrušte přesměrování nebo upgradujte na zvednu.cz", daysUntilRelease, assignedNumber)
	return c.SendSMS(ctx, to, body)
}

// SendPhoneNumberReleasedNotification sends notification that phone number has been released
func (c *SMSClient) SendPhoneNumberReleasedNotification(ctx context.Context, to, releasedNumber string) error {
	body := fmt.Sprintf("Zvednu: Číslo %s odpojeno. Prosím zrušte přesměrování hovorů. Pro obnovení: zvednu.cz", releasedNumber)
	return c.SendSMS(ctx, to, body)
}
