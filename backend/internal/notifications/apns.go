package notifications

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// APNsConfig holds configuration for Apple Push Notification service
type APNsConfig struct {
	KeyPath    string // Path to .p8 key file
	KeyID      string // Key ID from Apple Developer Portal
	TeamID     string // Team ID from Apple Developer Portal
	BundleID   string // App bundle ID (e.g., cz.zvednu.app)
	Production bool   // Use production environment
}

// APNsClient sends push notifications via Apple Push Notification service
type APNsClient struct {
	client   *apns2.Client
	bundleID string
	logger   *log.Logger
	mu       sync.Mutex
}

// NewAPNsClient creates a new APNs client
func NewAPNsClient(cfg APNsConfig, logger *log.Logger) (*APNsClient, error) {
	if cfg.KeyPath == "" || cfg.KeyID == "" || cfg.TeamID == "" || cfg.BundleID == "" {
		logger.Println("APNs: missing configuration, push notifications disabled")
		return nil, nil
	}

	// Load the .p8 key
	keyBytes, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read APNs key file: %w", err)
	}

	// Parse the private key
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode APNs key PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse APNs key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("APNs key is not an ECDSA private key")
	}

	// Create the auth token
	authToken := &token.Token{
		AuthKey: ecdsaKey,
		KeyID:   cfg.KeyID,
		TeamID:  cfg.TeamID,
	}

	// Create the client
	var client *apns2.Client
	if cfg.Production {
		client = apns2.NewTokenClient(authToken).Production()
	} else {
		client = apns2.NewTokenClient(authToken).Development()
	}

	logger.Printf("APNs: client initialized (production=%v, bundle=%s)", cfg.Production, cfg.BundleID)

	return &APNsClient{
		client:   client,
		bundleID: cfg.BundleID,
		logger:   logger,
	}, nil
}

// CallNotification represents data for a call notification
type CallNotification struct {
	CallID          string
	FromNumber      string
	IntentSummary   string
	LegitimacyLabel string
}

// SendCallNotification sends a push notification about a completed call
func (c *APNsClient) SendCallNotification(deviceToken string, notif CallNotification) error {
	if c == nil || c.client == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Build the notification payload
	p := payload.NewPayload().
		AlertTitle(fmt.Sprintf("Novy hovor od %s", notif.FromNumber)).
		AlertBody(notif.IntentSummary).
		Sound("default").
		Custom("call_id", notif.CallID).
		Custom("legitimacy_label", notif.LegitimacyLabel)

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     p,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	res, err := c.client.Push(notification)
	if err != nil {
		c.logger.Printf("APNs: failed to send notification: %v", err)
		return err
	}

	if res.StatusCode != 200 {
		c.logger.Printf("APNs: notification rejected (status=%d, reason=%s)", res.StatusCode, res.Reason)
		return fmt.Errorf("APNs rejected notification: %s", res.Reason)
	}

	c.logger.Printf("APNs: notification sent successfully to %s...", deviceToken[:16])
	return nil
}

// SendTestNotification sends a test notification
func (c *APNsClient) SendTestNotification(deviceToken, message string) error {
	if c == nil || c.client == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	p := payload.NewPayload().
		AlertTitle("Zvednu Test").
		AlertBody(message).
		Sound("default")

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     p,
		Expiration:  time.Now().Add(1 * time.Hour),
	}

	res, err := c.client.Push(notification)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("APNs rejected notification: %s", res.Reason)
	}

	return nil
}

// UsageWarningType represents the type of usage warning
type UsageWarningType string

const (
	UsageWarning80Percent UsageWarningType = "80_percent"
	UsageWarningExpired   UsageWarningType = "expired"
)

// TrialDayType represents the type of trial day notification
type TrialDayType string

const (
	TrialDay10 TrialDayType = "day10"
	TrialDay12 TrialDayType = "day12"
)

// SendUsageWarning sends a push notification about usage limits
func (c *APNsClient) SendUsageWarning(deviceToken string, warningType UsageWarningType, callsUsed, callsLimit int) error {
	if c == nil || c.client == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var title, body string
	switch warningType {
	case UsageWarning80Percent:
		title = "Blížíš se k limitu"
		body = fmt.Sprintf("Využili jste %d z %d hovorů. Upgradujte pro více hovorů.", callsUsed, callsLimit)
	case UsageWarningExpired:
		title = "Trial vypršel"
		body = "Karen nebude přijímat hovory. Upgradujte pro pokračování."
	default:
		return nil
	}

	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Sound("default").
		Custom("warning_type", string(warningType))

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     p,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	res, err := c.client.Push(notification)
	if err != nil {
		c.logger.Printf("APNs: failed to send usage warning: %v", err)
		return err
	}

	if res.StatusCode != 200 {
		c.logger.Printf("APNs: usage warning rejected (status=%d, reason=%s)", res.StatusCode, res.Reason)
		return fmt.Errorf("APNs rejected notification: %s", res.Reason)
	}

	c.logger.Printf("APNs: usage warning sent successfully to %s...", deviceToken[:16])
	return nil
}

// SendTrialDayNotification sends a push notification about trial days remaining
func (c *APNsClient) SendTrialDayNotification(deviceToken string, dayType TrialDayType, timeSavedMinutes, callsHandled int) error {
	if c == nil || c.client == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var title, body string
	switch dayType {
	case TrialDay10:
		title = "Zbývají 4 dny trialu"
		body = fmt.Sprintf("Karen ti zatím ušetřila %d minut. Upgraduj na zvednu.cz", timeSavedMinutes)
	case TrialDay12:
		title = "Zbývají 2 dny trialu"
		body = fmt.Sprintf("Karen ti vyřídila %d hovorů. Upgraduj na zvednu.cz", callsHandled)
	default:
		return nil
	}

	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Sound("default").
		Custom("notification_type", "trial_reminder").
		Custom("trial_day", string(dayType))

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     p,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	res, err := c.client.Push(notification)
	if err != nil {
		c.logger.Printf("APNs: failed to send trial day notification: %v", err)
		return err
	}

	if res.StatusCode != 200 {
		c.logger.Printf("APNs: trial day notification rejected (status=%d, reason=%s)", res.StatusCode, res.Reason)
		return fmt.Errorf("APNs rejected notification: %s", res.Reason)
	}

	c.logger.Printf("APNs: trial day notification sent successfully to %s...", deviceToken[:16])
	return nil
}

// SendTrialGraceWarning sends a push notification that phone number will be released
func (c *APNsClient) SendTrialGraceWarning(deviceToken string, daysUntilRelease int, assignedNumber string) error {
	if c == nil || c.client == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	title := "Služba bude ukončena"
	body := fmt.Sprintf("Za %d dní bude vaše číslo %s odpojeno. Zrušte přesměrování nebo upgradujte.", daysUntilRelease, assignedNumber)

	p := payload.NewPayload().
		AlertTitle(title).
		AlertBody(body).
		Sound("default").
		Custom("notification_type", "trial_grace_warning").
		Custom("days_until_release", daysUntilRelease)

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     p,
		Expiration:  time.Now().Add(24 * time.Hour),
	}

	res, err := c.client.Push(notification)
	if err != nil {
		c.logger.Printf("APNs: failed to send grace warning: %v", err)
		return err
	}

	if res.StatusCode != 200 {
		c.logger.Printf("APNs: grace warning rejected (status=%d, reason=%s)", res.StatusCode, res.Reason)
		return fmt.Errorf("APNs rejected notification: %s", res.Reason)
	}

	c.logger.Printf("APNs: grace warning sent successfully to %s...", deviceToken[:16])
	return nil
}
