//go:build chromedp_integration

// Integration smoke test for the PDF render core. Excluded from normal CI (no
// Chrome there); run locally against a real headless-shell sidecar:
//
//	docker run -d --name pdftest -p 9222:9222 chromedp/headless-shell:131.0.6778.86 \
//	  --no-sandbox --disable-gpu --disable-dev-shm-usage \
//	  --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222
//	go test ./controllers/ -tags chromedp_integration -run Integration -v
//
// Override target via PDFTEST_CHROME / PDFTEST_URL.
package controllers

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

func TestRenderDeckPDF_Integration(t *testing.T) {
	chrome := os.Getenv("PDFTEST_CHROME")
	if chrome == "" {
		chrome = "http://localhost:9222"
	}
	target := os.Getenv("PDFTEST_URL")
	if target == "" {
		// Classic public reveal.js deck — exercises print-pdf pagination.
		target = "https://lab.hakim.se/reveal-js/?print-pdf"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pdf, err := renderDeckPDF(ctx, chrome, target)
	if err != nil {
		t.Fatalf("renderDeckPDF: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("output is not a PDF (first bytes: %q)", pdf[:min(8, len(pdf))])
	}
	pages := bytes.Count(pdf, []byte("/Type /Page")) + bytes.Count(pdf, []byte("/Type/Page"))
	t.Logf("rendered %d bytes, ~%d page objects", len(pdf), pages)
	if len(pdf) < 2000 {
		t.Fatalf("PDF suspiciously small: %d bytes", len(pdf))
	}
}
