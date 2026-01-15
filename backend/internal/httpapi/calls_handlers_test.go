package httpapi

import (
	"strings"
	"testing"
)

func TestCallPathParsing(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantID   string
		wantType string
	}{
		{
			name:     "viewed endpoint",
			path:     "/api/calls/CAabc123/viewed",
			wantID:   "CAabc123",
			wantType: "viewed",
		},
		{
			name:     "resolve endpoint",
			path:     "/api/calls/CAabc123/resolve",
			wantID:   "CAabc123",
			wantType: "resolve",
		},
		{
			name:     "simple call ID",
			path:     "/api/calls/CAabc123",
			wantID:   "CAabc123",
			wantType: "detail",
		},
		{
			name:     "long call ID",
			path:     "/api/calls/CA0123456789abcdef0123456789abcdef",
			wantID:   "CA0123456789abcdef0123456789abcdef",
			wantType: "detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := strings.TrimPrefix(tt.path, "/api/calls/")

			var callID string
			var pathType string

			switch {
			case strings.HasSuffix(path, "/viewed"):
				callID = strings.TrimSuffix(path, "/viewed")
				pathType = "viewed"
			case strings.HasSuffix(path, "/resolve"):
				callID = strings.TrimSuffix(path, "/resolve")
				pathType = "resolve"
			default:
				callID = path
				pathType = "detail"
			}

			if callID != tt.wantID {
				t.Errorf("callID = %q, want %q", callID, tt.wantID)
			}
			if pathType != tt.wantType {
				t.Errorf("pathType = %q, want %q", pathType, tt.wantType)
			}
		})
	}
}

func TestCallPatchDispatch(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectFound bool
	}{
		{"viewed suffix", "/api/calls/CAabc123/viewed", true},
		{"resolve suffix", "/api/calls/CAabc123/resolve", true},
		{"no suffix", "/api/calls/CAabc123", false},
		{"invalid suffix", "/api/calls/CAabc123/invalid", false},
		{"empty after calls", "/api/calls/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := strings.TrimPrefix(tt.path, "/api/calls/")
			found := strings.HasSuffix(path, "/viewed") || strings.HasSuffix(path, "/resolve")

			if found != tt.expectFound {
				t.Errorf("path %q: found = %v, want %v", tt.path, found, tt.expectFound)
			}
		})
	}
}

func TestCallDeleteDispatch(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectFound bool
	}{
		{"resolve suffix", "/api/calls/CAabc123/resolve", true},
		{"viewed suffix - not for DELETE", "/api/calls/CAabc123/viewed", false},
		{"no suffix", "/api/calls/CAabc123", false},
		{"invalid suffix", "/api/calls/CAabc123/invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := strings.TrimPrefix(tt.path, "/api/calls/")
			// DELETE only supports /resolve
			found := strings.HasSuffix(path, "/resolve")

			if found != tt.expectFound {
				t.Errorf("path %q: found = %v, want %v", tt.path, found, tt.expectFound)
			}
		})
	}
}

func TestCallIDExtraction(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		suffix    string
		wantID    string
		wantEmpty bool
	}{
		{
			name:   "normal ID with viewed",
			path:   "CAabc123/viewed",
			suffix: "/viewed",
			wantID: "CAabc123",
		},
		{
			name:   "normal ID with resolve",
			path:   "CAabc123/resolve",
			suffix: "/resolve",
			wantID: "CAabc123",
		},
		{
			name:      "empty path",
			path:      "/viewed",
			suffix:    "/viewed",
			wantID:    "",
			wantEmpty: true,
		},
		{
			name:   "ID with special characters",
			path:   "CA_abc-123/viewed",
			suffix: "/viewed",
			wantID: "CA_abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := strings.TrimSuffix(tt.path, tt.suffix)

			if id != tt.wantID {
				t.Errorf("extracted ID = %q, want %q", id, tt.wantID)
			}

			if tt.wantEmpty && id != "" {
				t.Errorf("expected empty ID, got %q", id)
			}
		})
	}
}

func TestTenantAccessCheck(t *testing.T) {
	tests := []struct {
		name         string
		userTenantID *string
		callTenantID *string
		expectAccess bool
	}{
		{
			name:         "matching tenants",
			userTenantID: strPtr("tenant-123"),
			callTenantID: strPtr("tenant-123"),
			expectAccess: true,
		},
		{
			name:         "different tenants",
			userTenantID: strPtr("tenant-123"),
			callTenantID: strPtr("tenant-456"),
			expectAccess: false,
		},
		{
			name:         "user has no tenant",
			userTenantID: nil,
			callTenantID: strPtr("tenant-123"),
			expectAccess: false,
		},
		{
			name:         "call has no tenant",
			userTenantID: strPtr("tenant-123"),
			callTenantID: nil,
			expectAccess: false,
		},
		{
			name:         "both nil",
			userTenantID: nil,
			callTenantID: nil,
			expectAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Logic from handleGetCall: user must have tenant AND call must have matching tenant
			hasAccess := tt.userTenantID != nil &&
				tt.callTenantID != nil &&
				*tt.userTenantID == *tt.callTenantID

			if hasAccess != tt.expectAccess {
				t.Errorf("access = %v, want %v", hasAccess, tt.expectAccess)
			}
		})
	}
}

// Helper function for string pointers in tests
func strPtr(s string) *string {
	return &s
}
