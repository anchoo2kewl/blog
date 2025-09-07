package gotests

import (
    "net/http"
    "testing"
    u "anshumanbiswas.com/blog/utils"
)

func TestReadCookie_FindsCookie(t *testing.T) {
    r, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
    r.AddCookie(&http.Cookie{Name: u.CookieSession, Value: "tok123"})
    val, err := u.ReadCookie(r, u.CookieSession)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if val != "tok123" {
        t.Fatalf("expected 'tok123', got %q", val)
    }
}

func TestReadCookie_MissingCookie(t *testing.T) {
    r, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
    if _, err := u.ReadCookie(r, u.CookieUserEmail); err == nil {
        t.Fatalf("expected error for missing cookie")
    }
}

