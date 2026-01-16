package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lukasbauer/karen/internal/store"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/webhook"
)

// Stripe price IDs (set via environment variables)
var (
	stripePriceBasicMonthly = os.Getenv("STRIPE_PRICE_BASIC_MONTHLY")
	stripePriceBasicAnnual  = os.Getenv("STRIPE_PRICE_BASIC_ANNUAL")
	stripePriceProMonthly   = os.Getenv("STRIPE_PRICE_PRO_MONTHLY")
	stripePriceProAnnual    = os.Getenv("STRIPE_PRICE_PRO_ANNUAL")
	stripeWebhookSecret     = os.Getenv("STRIPE_WEBHOOK_SECRET")
	stripeSuccessURL        = os.Getenv("STRIPE_SUCCESS_URL")
	stripeCancelURL         = os.Getenv("STRIPE_CANCEL_URL")
)

func init() {
	// Set Stripe API key from environment
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// handleCreateCheckout creates a Stripe Checkout session for upgrading
func (r *Router) handleCreateCheckout(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil || authUser.TenantID == nil {
		http.Error(w, `{"error": "no tenant assigned"}`, http.StatusForbidden)
		return
	}

	var body struct {
		Plan     string `json:"plan"`     // "basic" or "pro"
		Interval string `json:"interval"` // "monthly" or "annual"
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Get the price ID based on plan and interval
	priceID := getPriceID(body.Plan, body.Interval)
	if priceID == "" {
		http.Error(w, `{"error": "invalid plan or interval"}`, http.StatusBadRequest)
		return
	}

	// Get tenant to check/create Stripe customer
	tenant, err := r.store.GetTenantByID(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
		return
	}

	// Get or create Stripe customer
	customerID, err := r.getOrCreateStripeCustomer(req.Context(), tenant, authUser.Phone)
	if err != nil {
		r.logger.Printf("billing: failed to get/create customer: %v", err)
		http.Error(w, `{"error": "failed to create customer"}`, http.StatusInternalServerError)
		return
	}

	// Create Checkout session
	successURL := stripeSuccessURL
	if successURL == "" {
		successURL = r.cfg.PublicBaseURL + "/billing/success?session_id={CHECKOUT_SESSION_ID}"
	}
	cancelURL := stripeCancelURL
	if cancelURL == "" {
		cancelURL = r.cfg.PublicBaseURL + "/billing/cancel"
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"tenant_id": tenant.ID,
			"plan":      body.Plan,
		},
	}

	s, err := checkoutsession.New(params)
	if err != nil {
		r.logger.Printf("billing: failed to create checkout session: %v", err)
		http.Error(w, `{"error": "failed to create checkout session"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("billing: created checkout session %s for tenant %s", s.ID, tenant.ID)

	writeJSON(w, http.StatusOK, map[string]string{
		"checkout_url": s.URL,
		"session_id":   s.ID,
	})
}

// handleCreatePortal creates a Stripe Customer Portal session
func (r *Router) handleCreatePortal(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil || authUser.TenantID == nil {
		http.Error(w, `{"error": "no tenant assigned"}`, http.StatusForbidden)
		return
	}

	// Get tenant
	tenant, err := r.store.GetTenantByID(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
		return
	}

	// Get billing info to get Stripe customer ID
	billingInfo, err := r.store.GetTenantBillingInfo(req.Context(), tenant.ID)
	if err != nil || billingInfo.StripeCustomerID == nil {
		http.Error(w, `{"error": "no subscription found"}`, http.StatusNotFound)
		return
	}

	returnURL := r.cfg.PublicBaseURL + "/settings"

	params := &stripe.BillingPortalSessionParams{
		Customer:  billingInfo.StripeCustomerID,
		ReturnURL: stripe.String(returnURL),
	}

	s, err := session.New(params)
	if err != nil {
		r.logger.Printf("billing: failed to create portal session: %v", err)
		http.Error(w, `{"error": "failed to create portal session"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"portal_url": s.URL,
	})
}

// handleStripeWebhook processes Stripe webhook events
func (r *Router) handleStripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Printf("billing webhook: failed to read body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	sigHeader := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, sigHeader, stripeWebhookSecret)
	if err != nil {
		r.logger.Printf("billing webhook: signature verification failed: %v", err)
		http.Error(w, "signature verification failed", http.StatusBadRequest)
		return
	}

	r.logger.Printf("billing webhook: received event %s (type=%s)", event.ID, event.Type)

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			r.logger.Printf("billing webhook: failed to parse session: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		r.handleCheckoutCompleted(&session)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			r.logger.Printf("billing webhook: failed to parse subscription: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		r.handleSubscriptionUpdated(&subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			r.logger.Printf("billing webhook: failed to parse subscription: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		r.handleSubscriptionDeleted(&subscription)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			r.logger.Printf("billing webhook: failed to parse invoice: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		r.handlePaymentSucceeded(&invoice)

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			r.logger.Printf("billing webhook: failed to parse invoice: %v", err)
			http.Error(w, "failed to parse event", http.StatusBadRequest)
			return
		}
		r.handlePaymentFailed(&invoice)
	}

	w.WriteHeader(http.StatusOK)
}

// handleCheckoutCompleted processes a completed checkout session
func (r *Router) handleCheckoutCompleted(session *stripe.CheckoutSession) {
	tenantID, ok := session.Metadata["tenant_id"]
	if !ok {
		r.logger.Printf("billing webhook: checkout session missing tenant_id")
		return
	}

	plan := session.Metadata["plan"]
	if plan == "" {
		plan = "basic" // Default to basic if not specified
	}

	// Update tenant with Stripe info
	// Use background context since webhooks are async
	ctx := context.Background()
	err := r.store.UpdateTenantBilling(ctx, tenantID, map[string]any{
		"stripe_customer_id":     session.Customer.ID,
		"stripe_subscription_id": session.Subscription.ID,
		"plan":                   plan,
		"status":                 "active",
		"current_period_start":   time.Now(),
		"current_period_calls":   0, // Reset calls for new billing period
	})

	if err != nil {
		r.logger.Printf("billing webhook: failed to update tenant %s: %v", tenantID, err)
		return
	}

	r.logger.Printf("billing webhook: upgraded tenant %s to plan %s", tenantID, plan)
}

// handleSubscriptionUpdated processes subscription updates
func (r *Router) handleSubscriptionUpdated(subscription *stripe.Subscription) {
	ctx := context.Background()
	// Find tenant by Stripe customer ID
	tenantID, err := r.store.GetTenantIDByStripeCustomer(ctx, subscription.Customer.ID)
	if err != nil {
		r.logger.Printf("billing webhook: tenant not found for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Determine plan from price ID
	plan := getPlanFromPriceID(subscription.Items.Data[0].Price.ID)

	// Update tenant
	status := "active"
	if subscription.Status == stripe.SubscriptionStatusCanceled {
		status = "cancelled"
	} else if subscription.Status == stripe.SubscriptionStatusPastDue {
		status = "past_due"
	}

	err = r.store.UpdateTenantBilling(ctx, tenantID, map[string]any{
		"plan":   plan,
		"status": status,
	})

	if err != nil {
		r.logger.Printf("billing webhook: failed to update tenant %s: %v", tenantID, err)
		return
	}

	r.logger.Printf("billing webhook: updated subscription for tenant %s (plan=%s, status=%s)", tenantID, plan, status)
}

// handleSubscriptionDeleted processes subscription cancellations
func (r *Router) handleSubscriptionDeleted(subscription *stripe.Subscription) {
	ctx := context.Background()
	// Find tenant by Stripe customer ID
	tenantID, err := r.store.GetTenantIDByStripeCustomer(ctx, subscription.Customer.ID)
	if err != nil {
		r.logger.Printf("billing webhook: tenant not found for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Downgrade to trial (but expired)
	err = r.store.UpdateTenantBilling(ctx, tenantID, map[string]any{
		"plan":                   "trial",
		"status":                 "cancelled",
		"stripe_subscription_id": nil,
	})

	if err != nil {
		r.logger.Printf("billing webhook: failed to update tenant %s: %v", tenantID, err)
		return
	}

	r.logger.Printf("billing webhook: subscription cancelled for tenant %s", tenantID)
}

// handlePaymentSucceeded processes successful payments
func (r *Router) handlePaymentSucceeded(invoice *stripe.Invoice) {
	if invoice.Subscription == nil {
		return // Not a subscription invoice
	}

	ctx := context.Background()
	// Find tenant by Stripe customer ID
	tenantID, err := r.store.GetTenantIDByStripeCustomer(ctx, invoice.Customer.ID)
	if err != nil {
		r.logger.Printf("billing webhook: tenant not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Reset period calls for new billing period
	err = r.store.ResetTenantPeriodCalls(ctx, tenantID)
	if err != nil {
		r.logger.Printf("billing webhook: failed to reset period calls for tenant %s: %v", tenantID, err)
		return
	}

	r.logger.Printf("billing webhook: payment succeeded for tenant %s, reset period calls", tenantID)
}

// handlePaymentFailed processes failed payments
func (r *Router) handlePaymentFailed(invoice *stripe.Invoice) {
	if invoice.Subscription == nil {
		return // Not a subscription invoice
	}

	ctx := context.Background()
	// Find tenant by Stripe customer ID
	tenantID, err := r.store.GetTenantIDByStripeCustomer(ctx, invoice.Customer.ID)
	if err != nil {
		r.logger.Printf("billing webhook: tenant not found for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Update status to past_due
	err = r.store.UpdateTenantBilling(ctx, tenantID, map[string]any{
		"status": "past_due",
	})

	if err != nil {
		r.logger.Printf("billing webhook: failed to update tenant %s: %v", tenantID, err)
		return
	}

	r.logger.Printf("billing webhook: payment failed for tenant %s", tenantID)

	// TODO: Send SMS notification about failed payment
}

// getOrCreateStripeCustomer gets an existing Stripe customer or creates a new one
func (r *Router) getOrCreateStripeCustomer(ctx context.Context, tenant *store.Tenant, phone string) (string, error) {
	// Check if tenant already has a Stripe customer ID
	billingInfo, err := r.store.GetTenantBillingInfo(ctx, tenant.ID)
	if err == nil && billingInfo.StripeCustomerID != nil && *billingInfo.StripeCustomerID != "" {
		return *billingInfo.StripeCustomerID, nil
	}

	// Create new customer
	params := &stripe.CustomerParams{
		Phone: stripe.String(phone),
		Name:  stripe.String(tenant.Name),
		Metadata: map[string]string{
			"tenant_id": tenant.ID,
		},
	}

	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Save the customer ID to the tenant
	_ = r.store.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
		"stripe_customer_id": c.ID,
	})

	return c.ID, nil
}

// getPriceID returns the Stripe price ID for a plan and interval
func getPriceID(plan, interval string) string {
	switch plan {
	case "basic":
		if interval == "annual" {
			return stripePriceBasicAnnual
		}
		return stripePriceBasicMonthly
	case "pro":
		if interval == "annual" {
			return stripePriceProAnnual
		}
		return stripePriceProMonthly
	default:
		return ""
	}
}

// getPlanFromPriceID determines the plan name from a Stripe price ID
func getPlanFromPriceID(priceID string) string {
	switch priceID {
	case stripePriceBasicMonthly, stripePriceBasicAnnual:
		return "basic"
	case stripePriceProMonthly, stripePriceProAnnual:
		return "pro"
	default:
		return "basic"
	}
}
