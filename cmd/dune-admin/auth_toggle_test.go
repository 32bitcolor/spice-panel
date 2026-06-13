package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClearSessionOnAuthToggle(t *testing.T) {
	hasClearCookie := func(w *httptest.ResponseRecorder) bool {
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName && c.MaxAge < 0 {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name        string
		wasEnabled  bool
		nowEnabled  bool
		wantCleared bool
	}{
		{"off → on clears", false, true, true},
		{"on → off clears", true, false, true},
		{"on → on no change", true, true, false},
		{"off → off no change", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/v1/config", nil)
			clearSessionOnAuthToggle(w, r, tt.wasEnabled, tt.nowEnabled)
			if got := hasClearCookie(w); got != tt.wantCleared {
				t.Errorf("cleared = %v, want %v", got, tt.wantCleared)
			}
		})
	}
}
