import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { setAuthToken, getAuthToken } from "../api";

describe("Billing API", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    localStorage.clear();
    setAuthToken("test-token");
  });

  afterEach(() => {
    global.fetch = originalFetch;
    localStorage.clear();
  });

  describe("getBilling", () => {
    it("fetches billing info successfully", async () => {
      const mockBillingInfo = {
        plan: "trial",
        status: "active",
        call_status: {
          can_receive: true,
          reason: "ok",
          calls_used: 5,
          calls_limit: 20,
          trial_days_left: 10,
          trial_calls_left: 15,
        },
        trial_ends_at: "2026-01-30T00:00:00Z",
        total_time_saved: 3600,
        current_usage: {
          calls_count: 5,
          minutes_used: 25,
          time_saved_seconds: 3600,
          spam_calls_blocked: 2,
          period_start: "2026-01-01",
          period_end: "2026-01-31",
        },
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockBillingInfo),
      });

      const { api } = await import("../api");
      const result = await api.getBilling();

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/billing"),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
            "Content-Type": "application/json",
          }),
        })
      );
      expect(result).toEqual(mockBillingInfo);
    });

    it("handles error response", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        text: () => Promise.resolve("Unauthorized"),
      });

      const { api } = await import("../api");
      await expect(api.getBilling()).rejects.toThrow();
    });
  });

  describe("createCheckout", () => {
    it("creates checkout session for basic monthly", async () => {
      const mockResponse = {
        checkout_url: "https://checkout.stripe.com/session_123",
        session_id: "cs_test_123",
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const { api } = await import("../api");
      const result = await api.createCheckout("basic", "monthly");

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/billing/checkout"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ plan: "basic", interval: "monthly" }),
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
            "Content-Type": "application/json",
          }),
        })
      );
      expect(result.checkout_url).toBe("https://checkout.stripe.com/session_123");
      expect(result.session_id).toBe("cs_test_123");
    });

    it("creates checkout session for pro annual", async () => {
      const mockResponse = {
        checkout_url: "https://checkout.stripe.com/session_456",
        session_id: "cs_test_456",
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const { api } = await import("../api");
      const result = await api.createCheckout("pro", "annual");

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/billing/checkout"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ plan: "pro", interval: "annual" }),
        })
      );
      expect(result).toEqual(mockResponse);
    });

    it("handles server error", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        text: () => Promise.resolve("Internal server error"),
      });

      const { api } = await import("../api");
      await expect(api.createCheckout("basic", "monthly")).rejects.toThrow();
    });

    it("handles no tenant error", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        text: () => Promise.resolve('{"error": "no tenant assigned"}'),
      });

      const { api } = await import("../api");
      await expect(api.createCheckout("basic", "monthly")).rejects.toThrow();
    });
  });

  describe("createPortal", () => {
    it("creates portal session successfully", async () => {
      const mockResponse = {
        portal_url: "https://billing.stripe.com/session_789",
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const { api } = await import("../api");
      const result = await api.createPortal();

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/billing/portal"),
        expect.objectContaining({
          method: "POST",
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
            "Content-Type": "application/json",
          }),
        })
      );
      expect(result.portal_url).toBe("https://billing.stripe.com/session_789");
    });

    it("handles unauthorized error", async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        text: () => Promise.resolve("Unauthorized"),
      });

      const { api } = await import("../api");
      await expect(api.createPortal()).rejects.toThrow();
    });
  });
});

describe("Billing API Authentication", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    global.fetch = originalFetch;
    localStorage.clear();
  });

  it("includes auth token in billing request headers when set", async () => {
    setAuthToken("my-auth-token");

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ plan: "trial", status: "active", call_status: { can_receive: true } }),
    });

    const { api } = await import("../api");
    await api.getBilling();

    expect(global.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer my-auth-token",
        }),
      })
    );
  });

  it("clears token on 401 from billing endpoint", async () => {
    setAuthToken("expired-token");

    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      text: () => Promise.resolve("Unauthorized"),
    });

    const { api } = await import("../api");
    await expect(api.getBilling()).rejects.toThrow();
    expect(getAuthToken()).toBeNull();
  });
});
