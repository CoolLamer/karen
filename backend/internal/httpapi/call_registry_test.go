package httpapi

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestCallRegistry_AddAndDone(t *testing.T) {
	cr := NewCallRegistry()

	if cr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0", cr.ActiveCount())
	}

	if !cr.Add() {
		t.Error("Add() should return true when not draining")
	}
	if cr.ActiveCount() != 1 {
		t.Errorf("ActiveCount() = %d, want 1", cr.ActiveCount())
	}

	if !cr.Add() {
		t.Error("Add() should return true when not draining")
	}
	if cr.ActiveCount() != 2 {
		t.Errorf("ActiveCount() = %d, want 2", cr.ActiveCount())
	}

	cr.Done()
	if cr.ActiveCount() != 1 {
		t.Errorf("ActiveCount() = %d, want 1 after one Done()", cr.ActiveCount())
	}

	cr.Done()
	if cr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0 after all Done()", cr.ActiveCount())
	}
}

func TestCallRegistry_Draining(t *testing.T) {
	cr := NewCallRegistry()

	if cr.IsDraining() {
		t.Error("IsDraining() should be false initially")
	}

	// Add a call before draining
	if !cr.Add() {
		t.Error("Add() should succeed before draining")
	}

	cr.StartDraining()

	if !cr.IsDraining() {
		t.Error("IsDraining() should be true after StartDraining()")
	}

	// New calls should be rejected
	if cr.Add() {
		t.Error("Add() should return false when draining")
	}

	// Active count should still be 1 (the pre-drain call)
	if cr.ActiveCount() != 1 {
		t.Errorf("ActiveCount() = %d, want 1", cr.ActiveCount())
	}

	// Complete the existing call
	cr.Done()
	if cr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0", cr.ActiveCount())
	}
}

func TestCallRegistry_WaitBlocksUntilDone(t *testing.T) {
	cr := NewCallRegistry()

	cr.Add()
	cr.Add()

	done := make(chan struct{})
	go func() {
		cr.Wait()
		close(done)
	}()

	// Wait should not complete yet
	select {
	case <-done:
		t.Error("Wait() should block while calls are active")
	default:
	}

	cr.Done()

	// Still one active
	select {
	case <-done:
		t.Error("Wait() should block while calls are active")
	default:
	}

	cr.Done()

	// Now Wait should complete
	<-done
}

func TestCallRegistry_ConcurrentAddAndDone(t *testing.T) {
	cr := NewCallRegistry()
	const n = 100

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if cr.Add() {
				defer cr.Done()
			}
		}()
	}

	wg.Wait()

	if cr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0 after all goroutines done", cr.ActiveCount())
	}
}

func TestCallRegistry_DrainDuringConcurrentAdds(t *testing.T) {
	cr := NewCallRegistry()
	const n = 100

	var wg sync.WaitGroup
	var accepted, rejected int64
	var mu sync.Mutex

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if cr.Add() {
				mu.Lock()
				accepted++
				mu.Unlock()
				defer cr.Done()
			} else {
				mu.Lock()
				rejected++
				mu.Unlock()
			}
		}()

		// Start draining midway
		if i == n/2 {
			cr.StartDraining()
		}
	}

	wg.Wait()

	if accepted+rejected != n {
		t.Errorf("accepted(%d) + rejected(%d) != %d", accepted, rejected, n)
	}
	if rejected == 0 {
		t.Error("expected some calls to be rejected after draining started")
	}
	if cr.ActiveCount() != 0 {
		t.Errorf("ActiveCount() = %d, want 0", cr.ActiveCount())
	}
}

func TestReadyzEndpoint(t *testing.T) {
	cr := NewCallRegistry()
	r := &Router{
		logger: log.New(io.Discard, "", 0),
		calls:  cr,
	}

	t.Run("returns 200 when not draining", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		r.handleReadyz(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if body := rec.Body.String(); body != "ok" {
			t.Errorf("body = %q, want %q", body, "ok")
		}
	})

	t.Run("returns 503 when draining", func(t *testing.T) {
		cr.StartDraining()

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		r.handleReadyz(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
		if body := rec.Body.String(); body != "draining" {
			t.Errorf("body = %q, want %q", body, "draining")
		}
	})
}

func TestInboundRejectsDuringDrain(t *testing.T) {
	cr := NewCallRegistry()
	cr.StartDraining()

	r := &Router{
		logger: log.New(io.Discard, "", 0),
		calls:  cr,
	}

	req := httptest.NewRequest(http.MethodPost, "/telephony/inbound", nil)
	rec := httptest.NewRecorder()
	r.handleTwilioInbound(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected TwiML response body")
	}
	// Should contain Reject with reason="busy"
	if !strings.Contains(body, "<Reject") {
		t.Error("response should contain <Reject> TwiML")
	}
	if !strings.Contains(body, `reason="busy"`) {
		t.Error("response should contain reason=\"busy\"")
	}
}
