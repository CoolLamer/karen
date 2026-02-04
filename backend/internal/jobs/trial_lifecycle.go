package jobs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lukasbauer/karen/internal/notifications"
	"github.com/lukasbauer/karen/internal/store"
)

// TrialLifecycleJob processes trial lifecycle notifications and phone number releases.
// It runs on a configurable interval (default: 1 hour) and:
// - Sends Day 10/12/14 conversion prompts
// - Sends grace period warnings when trial expires
// - Releases phone numbers after grace period ends
type TrialLifecycleJob struct {
	store    *store.Store
	sms      *notifications.SMSClient
	apns     *notifications.APNsClient
	logger   *log.Logger
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewTrialLifecycleJob creates a new trial lifecycle job.
func NewTrialLifecycleJob(s *store.Store, sms *notifications.SMSClient, apns *notifications.APNsClient, logger *log.Logger, interval time.Duration) *TrialLifecycleJob {
	if interval == 0 {
		interval = 1 * time.Hour
	}
	return &TrialLifecycleJob{
		store:    s,
		sms:      sms,
		apns:     apns,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the background job.
func (j *TrialLifecycleJob) Start() {
	j.wg.Add(1)
	go j.run()
	j.logger.Printf("TrialLifecycleJob: started (interval=%v)", j.interval)
}

// Stop gracefully stops the background job.
func (j *TrialLifecycleJob) Stop() {
	close(j.stopCh)
	j.wg.Wait()
	j.logger.Println("TrialLifecycleJob: stopped")
}

func (j *TrialLifecycleJob) run() {
	defer j.wg.Done()

	// Run immediately on start
	j.processAll()

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.processAll()
		case <-j.stopCh:
			return
		}
	}
}

func (j *TrialLifecycleJob) processAll() {
	ctx := context.Background()

	// Fetch SMS sender number from global_config before processing
	j.updateSMSSenderNumber(ctx)

	j.processDay10Notifications(ctx)
	j.processDay12Notifications(ctx)
	j.processDay14Notifications(ctx)
	j.processGraceNotifications(ctx)
	j.processPhoneNumberReleases(ctx)
}

// updateSMSSenderNumber fetches the SMS sender number from global_config
// and updates the SMS client. This allows runtime configuration changes.
func (j *TrialLifecycleJob) updateSMSSenderNumber(ctx context.Context) {
	if j.sms == nil {
		return
	}

	senderNumber, err := j.store.GetGlobalConfig(ctx, "sms_sender_number")
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get sms_sender_number from config: %v", err)
		return
	}

	if senderNumber == "" {
		j.logger.Println("TrialLifecycleJob: sms_sender_number not configured in global_config")
		return
	}

	// Only log if the number changed
	currentNumber := j.sms.GetSenderNumber()
	if currentNumber != senderNumber {
		j.sms.SetSenderNumber(senderNumber)
		j.logger.Printf("TrialLifecycleJob: updated SMS sender number to %s", senderNumber)
	}
}

func (j *TrialLifecycleJob) processDay10Notifications(ctx context.Context) {
	tenants, err := j.store.GetTenantsNeedingDay10Notification(ctx)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get day10 tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		j.sendDay10Notifications(ctx, tenant)
		if err := j.store.MarkTrialDay10NotificationSent(ctx, tenant.TenantID); err != nil {
			j.logger.Printf("TrialLifecycleJob: failed to mark day10 sent for tenant %s: %v", tenant.TenantID, err)
		}
	}

	if len(tenants) > 0 {
		j.logger.Printf("TrialLifecycleJob: processed %d day10 notifications", len(tenants))
	}
}

func (j *TrialLifecycleJob) processDay12Notifications(ctx context.Context) {
	tenants, err := j.store.GetTenantsNeedingDay12Notification(ctx)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get day12 tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		j.sendDay12Notifications(ctx, tenant)
		if err := j.store.MarkTrialDay12NotificationSent(ctx, tenant.TenantID); err != nil {
			j.logger.Printf("TrialLifecycleJob: failed to mark day12 sent for tenant %s: %v", tenant.TenantID, err)
		}
	}

	if len(tenants) > 0 {
		j.logger.Printf("TrialLifecycleJob: processed %d day12 notifications", len(tenants))
	}
}

func (j *TrialLifecycleJob) processDay14Notifications(ctx context.Context) {
	tenants, err := j.store.GetTenantsNeedingDay14Notification(ctx)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get day14 tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		j.sendDay14Notifications(ctx, tenant)
		if err := j.store.MarkTrialDay14NotificationSent(ctx, tenant.TenantID); err != nil {
			j.logger.Printf("TrialLifecycleJob: failed to mark day14 sent for tenant %s: %v", tenant.TenantID, err)
		}
	}

	if len(tenants) > 0 {
		j.logger.Printf("TrialLifecycleJob: processed %d day14 notifications", len(tenants))
	}
}

func (j *TrialLifecycleJob) processGraceNotifications(ctx context.Context) {
	tenants, err := j.store.GetTenantsNeedingGraceNotification(ctx)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get grace tenants: %v", err)
		return
	}

	// Get grace period from config
	gracePeriodDays := j.getGracePeriodDays(ctx)

	for _, tenant := range tenants {
		j.sendGraceNotifications(ctx, tenant, gracePeriodDays)
		if err := j.store.MarkTrialGraceNotificationSent(ctx, tenant.TenantID); err != nil {
			j.logger.Printf("TrialLifecycleJob: failed to mark grace sent for tenant %s: %v", tenant.TenantID, err)
		}
	}

	if len(tenants) > 0 {
		j.logger.Printf("TrialLifecycleJob: processed %d grace notifications", len(tenants))
	}
}

func (j *TrialLifecycleJob) processPhoneNumberReleases(ctx context.Context) {
	gracePeriodDays := j.getGracePeriodDays(ctx)

	tenants, err := j.store.GetTenantsForPhoneNumberRelease(ctx, gracePeriodDays)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get release tenants: %v", err)
		return
	}

	for _, tenant := range tenants {
		// Get phone number before releasing
		phoneNumber := ""
		if tenant.PhoneNumber != nil {
			phoneNumber = *tenant.PhoneNumber
		}

		// Release the phone number
		if err := j.store.ReleaseExpiredTrialPhoneNumber(ctx, tenant.TenantID); err != nil {
			j.logger.Printf("TrialLifecycleJob: failed to release phone for tenant %s: %v", tenant.TenantID, err)
			continue
		}

		// Send release notifications
		j.sendReleaseNotifications(ctx, tenant, phoneNumber)

		j.logger.Printf("TrialLifecycleJob: released phone number %s from tenant %s", phoneNumber, tenant.TenantID)
	}

	if len(tenants) > 0 {
		j.logger.Printf("TrialLifecycleJob: released %d phone numbers", len(tenants))
	}
}

func (j *TrialLifecycleJob) getGracePeriodDays(ctx context.Context) int {
	val, err := j.store.GetGlobalConfig(ctx, "trial_grace_period_days")
	if err != nil {
		return 7 // default
	}
	days, err := strconv.Atoi(val)
	if err != nil {
		return 7 // default
	}
	return days
}

func (j *TrialLifecycleJob) logNotification(ctx context.Context, channel, notifType, recipient string, tenantID string, body string, err error) {
	status := "sent"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}
	if logErr := j.store.InsertNotificationLog(ctx, channel, notifType, recipient, &tenantID, body, status, errMsg); logErr != nil {
		j.logger.Printf("TrialLifecycleJob: failed to log notification: %v", logErr)
	}
}

func (j *TrialLifecycleJob) sendDay10Notifications(ctx context.Context, tenant store.TrialTenantInfo) {
	users, err := j.store.GetTenantUsers(ctx, tenant.TenantID)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get users for tenant %s: %v", tenant.TenantID, err)
		return
	}

	timeSavedMinutes := tenant.TimeSavedTotal / 60
	smsBody := fmt.Sprintf("Zvednu: Zbývají ti 4 dny trialu. Karen ti zatím ušetřila %d minut. Upgraduj na zvednu.cz", timeSavedMinutes)
	pushBody := fmt.Sprintf("Karen ti zatím ušetřila %d minut. Upgraduj na zvednu.cz", timeSavedMinutes)

	for _, user := range users {
		// Send SMS
		if j.sms != nil {
			smsErr := j.sms.SendTrialDay10Notification(ctx, user.Phone, timeSavedMinutes)
			if smsErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day10 SMS to %s: %v", user.Phone, smsErr)
			}
			j.logNotification(ctx, "sms", "trial_day10", user.Phone, tenant.TenantID, smsBody, smsErr)
		}

		// Send APNs push
		if j.apns != nil && user.PushToken != nil {
			pushErr := j.apns.SendTrialDayNotification(*user.PushToken, notifications.TrialDay10, timeSavedMinutes, tenant.CallsHandled)
			if pushErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day10 push: %v", pushErr)
			}
			j.logNotification(ctx, "apns", "trial_day10", (*user.PushToken)[:min(16, len(*user.PushToken))]+"...", tenant.TenantID, pushBody, pushErr)
		}
	}
}

func (j *TrialLifecycleJob) sendDay12Notifications(ctx context.Context, tenant store.TrialTenantInfo) {
	users, err := j.store.GetTenantUsers(ctx, tenant.TenantID)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get users for tenant %s: %v", tenant.TenantID, err)
		return
	}

	smsBody := fmt.Sprintf("Zvednu: Zbývají ti 2 dny trialu. Karen ti vyřídila %d hovorů. Upgraduj na zvednu.cz", tenant.CallsHandled)
	pushBody := fmt.Sprintf("Karen ti vyřídila %d hovorů. Upgraduj na zvednu.cz", tenant.CallsHandled)

	for _, user := range users {
		// Send SMS
		if j.sms != nil {
			smsErr := j.sms.SendTrialDay12Notification(ctx, user.Phone, tenant.CallsHandled)
			if smsErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day12 SMS to %s: %v", user.Phone, smsErr)
			}
			j.logNotification(ctx, "sms", "trial_day12", user.Phone, tenant.TenantID, smsBody, smsErr)
		}

		// Send APNs push
		if j.apns != nil && user.PushToken != nil {
			pushErr := j.apns.SendTrialDayNotification(*user.PushToken, notifications.TrialDay12, tenant.TimeSavedTotal/60, tenant.CallsHandled)
			if pushErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day12 push: %v", pushErr)
			}
			j.logNotification(ctx, "apns", "trial_day12", (*user.PushToken)[:min(16, len(*user.PushToken))]+"...", tenant.TenantID, pushBody, pushErr)
		}
	}
}

func (j *TrialLifecycleJob) sendDay14Notifications(ctx context.Context, tenant store.TrialTenantInfo) {
	users, err := j.store.GetTenantUsers(ctx, tenant.TenantID)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get users for tenant %s: %v", tenant.TenantID, err)
		return
	}

	smsBody := "Zvednu: Trial skončil. Karen nebude přijímat hovory. Upgraduj na zvednu.cz"
	pushBody := "Trial skončil. Karen nebude přijímat hovory. Upgraduj na zvednu.cz"

	for _, user := range users {
		// Send SMS
		if j.sms != nil {
			smsErr := j.sms.SendTrialExpiredNotification(ctx, user.Phone)
			if smsErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day14 SMS to %s: %v", user.Phone, smsErr)
			}
			j.logNotification(ctx, "sms", "trial_expired", user.Phone, tenant.TenantID, smsBody, smsErr)
		}

		// Send APNs push (using existing UsageWarningExpired)
		if j.apns != nil && user.PushToken != nil {
			pushErr := j.apns.SendUsageWarning(*user.PushToken, notifications.UsageWarningExpired, tenant.CallsHandled, 20)
			if pushErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send day14 push: %v", pushErr)
			}
			j.logNotification(ctx, "apns", "trial_expired", (*user.PushToken)[:min(16, len(*user.PushToken))]+"...", tenant.TenantID, pushBody, pushErr)
		}
	}
}

func (j *TrialLifecycleJob) sendGraceNotifications(ctx context.Context, tenant store.TrialTenantInfo, gracePeriodDays int) {
	users, err := j.store.GetTenantUsers(ctx, tenant.TenantID)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get users for tenant %s: %v", tenant.TenantID, err)
		return
	}

	phoneNumber := ""
	if tenant.PhoneNumber != nil {
		phoneNumber = *tenant.PhoneNumber
	}

	smsBody := fmt.Sprintf("Zvednu: Za %d dní bude vaše číslo %s odpojeno. Zrušte přesměrování nebo upgradujte na zvednu.cz", gracePeriodDays, phoneNumber)
	pushBody := fmt.Sprintf("Za %d dní bude vaše číslo %s odpojeno. Zrušte přesměrování nebo upgradujte.", gracePeriodDays, phoneNumber)

	for _, user := range users {
		// Send SMS
		if j.sms != nil {
			smsErr := j.sms.SendTrialGraceWarningNotification(ctx, user.Phone, phoneNumber, gracePeriodDays)
			if smsErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send grace SMS to %s: %v", user.Phone, smsErr)
			}
			j.logNotification(ctx, "sms", "grace_warning", user.Phone, tenant.TenantID, smsBody, smsErr)
		}

		// Send APNs push
		if j.apns != nil && user.PushToken != nil {
			pushErr := j.apns.SendTrialGraceWarning(*user.PushToken, gracePeriodDays, phoneNumber)
			if pushErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send grace push: %v", pushErr)
			}
			j.logNotification(ctx, "apns", "grace_warning", (*user.PushToken)[:min(16, len(*user.PushToken))]+"...", tenant.TenantID, pushBody, pushErr)
		}
	}
}

func (j *TrialLifecycleJob) sendReleaseNotifications(ctx context.Context, tenant store.TrialTenantInfo, releasedNumber string) {
	users, err := j.store.GetTenantUsers(ctx, tenant.TenantID)
	if err != nil {
		j.logger.Printf("TrialLifecycleJob: failed to get users for tenant %s: %v", tenant.TenantID, err)
		return
	}

	smsBody := fmt.Sprintf("Zvednu: Číslo %s odpojeno. Prosím zrušte přesměrování hovorů. Pro obnovení: zvednu.cz", releasedNumber)

	for _, user := range users {
		// Send SMS only (push might not work if app uninstalled)
		if j.sms != nil {
			smsErr := j.sms.SendPhoneNumberReleasedNotification(ctx, user.Phone, releasedNumber)
			if smsErr != nil {
				j.logger.Printf("TrialLifecycleJob: failed to send release SMS to %s: %v", user.Phone, smsErr)
			}
			j.logNotification(ctx, "sms", "phone_released", user.Phone, tenant.TenantID, smsBody, smsErr)
		}
	}
}
