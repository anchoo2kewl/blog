package controllers

// Slide PDF export.
//
// Browser-exact PDF of a deck. We do NOT re-implement reveal.js layout — we
// drive a headless Chrome (the `pdfsvc` sidecar, a stock chromedp/headless-shell
// image) at the deck's own `?print-pdf` view and capture Page.PrintToPDF with
// preferCSSPageSize — the same technique decktape uses. This is the minimal
// "ripped" core of gopdfsuit's HTML-conversion path, without the rest of that
// suite (signatures, barcodes, forms, …).
//
// The blog app holds NO Chrome binary; it only speaks the DevTools Protocol
// (chromedp is pure Go) to the sidecar over the internal docker network. A
// short-lived HMAC export token lets the sidecar's Chrome fetch the one deck
// being exported past the password gate; the public endpoint enforces the same
// view rules first.

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"anshumanbiswas.com/blog/models"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// pdfExportTTL bounds how long a minted export token (and the whole render) is
// valid. Generous enough for a cold Chrome tab + CDN asset fetch, short enough
// that a leaked URL is useless minutes later.
const pdfExportTTL = 90 * time.Second

// pdfRenderSem bounds concurrent Chrome tabs so a burst of export requests
// can't OOM the sidecar. Buffered to a small fixed fleet.
var pdfRenderSem = make(chan struct{}, 2)

// mintExportToken returns hex(HMAC-SHA256(secret, "slug|exp")).
func mintExportToken(secret, slug string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%s|%d", slug, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

// validExportToken constant-time-checks a token for slug and that exp is in the
// future. Returns false if the secret is unset (feature disabled / misconfig).
func validExportToken(secret, slug, token string, exp int64) bool {
	if secret == "" || token == "" {
		return false
	}
	if exp < time.Now().Unix() {
		return false
	}
	want := mintExportToken(secret, slug, exp)
	return hmac.Equal([]byte(want), []byte(token))
}

// hasValidExportToken reads export_token/exp from the request and validates them
// against this slug. Used by ViewSlide to let the sidecar's Chrome bypass the
// password gate and the published check for exactly the deck being exported.
func (s Slides) hasValidExportToken(r *http.Request, slug string) bool {
	q := r.URL.Query()
	exp, _ := strconv.ParseInt(q.Get("exp"), 10, 64)
	return validExportToken(s.ExportSecret, slug, q.Get("export_token"), exp)
}

// ExportSlidePDF handles GET /slides/{slug}/pdf — streams a browser-exact PDF.
func (s Slides) ExportSlidePDF(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	if s.PDFRemoteURL == "" || s.ExportSecret == "" {
		http.Error(w, "PDF export is not configured", http.StatusServiceUnavailable)
		return
	}

	slide, err := s.SlideService.GetBySlug(slug)
	if err != nil {
		http.Error(w, "Slide not found", http.StatusNotFound)
		return
	}

	// Access parity with ViewSlide: editors always; otherwise the deck must be
	// published, and if it is password-protected the caller must already hold a
	// granted access cookie (i.e. they got past the gate to view it).
	user, _ := s.isUserLoggedIn(r)
	isEditor := user != nil && models.CanEditSlides(user.Role)
	if !isEditor {
		if !slide.IsPublished {
			http.Error(w, "Slide not found", http.StatusNotFound)
			return
		}
		if slide.PasswordHash != "" {
			cookie, err := r.Cookie(fmt.Sprintf("slide_access_%s", slug))
			if err != nil || cookie.Value != "granted" {
				http.Error(w, "Unauthorized — open the presentation first", http.StatusUnauthorized)
				return
			}
		}
	}

	pdf, err := s.exportPDF(r.Context(), slug)
	if err != nil {
		log.Printf("slide pdf export (%s): %v", slug, err)
		http.Error(w, "Failed to render PDF", http.StatusBadGateway)
		return
	}
	writePDF(w, slug, pdf)
}

// APIExportSlidePDF handles GET/POST /api/slides/{slideID}/pdf — the
// Bearer-token-authenticated export path (the route group runs
// APIAuthMiddleware). For password-protected decks the caller must supply the
// correct deck password (query/form `password`), verified against the stored
// bcrypt hash, before we render — "only once the password has been provided".
func (s Slides) APIExportSlidePDF(w http.ResponseWriter, r *http.Request) {
	if s.PDFRemoteURL == "" || s.ExportSecret == "" {
		http.Error(w, "PDF export is not configured", http.StatusServiceUnavailable)
		return
	}

	slideID, err := strconv.Atoi(chi.URLParam(r, "slideID"))
	if err != nil {
		http.Error(w, "Invalid slide ID", http.StatusBadRequest)
		return
	}
	slide, err := s.SlideService.GetByID(slideID)
	if err != nil {
		http.Error(w, "Slide not found", http.StatusNotFound)
		return
	}

	// Password-protected decks: require the deck password even for token
	// holders. Accept it from query (?password=) or form body.
	if slide.PasswordHash != "" {
		_ = r.ParseForm()
		password := r.URL.Query().Get("password")
		if password == "" {
			password = r.FormValue("password")
		}
		if password == "" {
			http.Error(w, "This deck is password-protected — provide the deck password", http.StatusUnauthorized)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(slide.PasswordHash), []byte(password)) != nil {
			http.Error(w, "Incorrect deck password", http.StatusForbidden)
			return
		}
	}

	pdf, err := s.exportPDF(r.Context(), slide.Slug)
	if err != nil {
		log.Printf("api slide pdf export (%s): %v", slide.Slug, err)
		http.Error(w, "Failed to render PDF", http.StatusBadGateway)
		return
	}
	writePDF(w, slide.Slug, pdf)
}

// exportPDF mints a short-lived token, builds the internal print URL the
// sidecar's Chrome will load, and renders it. ExportBaseURL is how that Chrome
// reaches THIS app over the docker network (e.g. http://app:22222), not the
// public hostname. Concurrency is bounded so export bursts can't OOM the
// sidecar.
func (s Slides) exportPDF(parent context.Context, slug string) ([]byte, error) {
	exp := time.Now().Add(pdfExportTTL).Unix()
	tok := mintExportToken(s.ExportSecret, slug, exp)
	printURL := fmt.Sprintf("%s/slides/%s?print-pdf&export_token=%s&exp=%d",
		strings.TrimRight(s.ExportBaseURL, "/"), url.PathEscape(slug), tok, exp)

	pdfRenderSem <- struct{}{}
	defer func() { <-pdfRenderSem }()

	ctx, cancel := context.WithTimeout(parent, pdfExportTTL)
	defer cancel()
	return renderDeckPDF(ctx, s.PDFRemoteURL, printURL)
}

// writePDF streams PDF bytes as a download.
func writePDF(w http.ResponseWriter, slug string, pdf []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, slug))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(pdf)
}

// renderDeckPDF connects to the remote Chrome (chromedp/headless-shell) at
// remoteBase (e.g. http://pdfsvc:9222), navigates to printURL, waits for
// reveal.js to finish its print-pdf layout, and returns the printed PDF bytes.
func renderDeckPDF(ctx context.Context, remoteBase, printURL string) ([]byte, error) {
	wsURL, err := resolveChromeWS(ctx, remoteBase)
	if err != nil {
		return nil, fmt.Errorf("resolve chrome ws: %w", err)
	}

	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(ctx, wsURL)
	defer cancelAlloc()
	// Fresh tab per export; sharing the remote browser is fine.
	tabCtx, cancelTab := chromedp.NewContext(allocCtx)
	defer cancelTab()

	var pdf []byte
	err = chromedp.Run(tabCtx,
		// A landscape-ish viewport so reveal lays the print pages out sensibly;
		// preferCSSPageSize below still honors reveal's own @page size.
		chromedp.EmulateViewport(1280, 720),
		chromedp.Navigate(printURL),
		// reveal.js sets document.title and toggles the print-pdf body class once
		// its print layout is ready; wait for that, then settle for fonts/CDN CSS.
		chromedp.ActionFunc(waitRevealPrintReady),
		chromedp.ActionFunc(func(ctx context.Context) error {
			data, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return err
			}
			pdf = data
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	if len(pdf) == 0 {
		return nil, fmt.Errorf("empty pdf")
	}
	return pdf, nil
}

// waitRevealPrintReady polls until reveal.js reports ready (or a short deadline),
// then pauses briefly so web fonts and CDN stylesheets are painted before print.
func waitRevealPrintReady(ctx context.Context) error {
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		var ready bool
		// Reveal exposes isReady(); guard for it not yet existing.
		if err := chromedp.Evaluate(
			`!!(window.Reveal && Reveal.isReady && Reveal.isReady())`, &ready,
		).Do(ctx); err == nil && ready {
			break
		}
		if err := chromedp.Sleep(200 * time.Millisecond).Do(ctx); err != nil {
			return err
		}
	}
	// Final settle for font swap + print stylesheet application.
	return chromedp.Sleep(1500 * time.Millisecond).Do(ctx)
}

// resolveChromeWS turns an http base (http://host:9222) into the browser
// webSocketDebuggerUrl that NewRemoteAllocator needs. headless-shell binds to
// its container hostname, so we rewrite the ws host back to the one we dialed.
func resolveChromeWS(ctx context.Context, base string) (string, error) {
	base = strings.TrimRight(base, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/json/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", err
	}
	var v struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &v); err != nil || v.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("no webSocketDebuggerUrl from %s: %s", base, strings.TrimSpace(string(body)))
	}
	// The ws URL Chrome reports uses ITS view of the host (often 127.0.0.1 or the
	// container id). Rewrite host:port to what we actually dialed so the app can
	// reach it across the docker network.
	dialed, err := url.Parse(base)
	if err != nil {
		return v.WebSocketDebuggerURL, nil
	}
	ws, err := url.Parse(v.WebSocketDebuggerURL)
	if err != nil {
		return v.WebSocketDebuggerURL, nil
	}
	ws.Host = dialed.Host
	return ws.String(), nil
}
