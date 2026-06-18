package controllers

import (
	"testing"
	"time"
)

func TestExportToken_RoundTrip(t *testing.T) {
	const secret = "test-secret"
	exp := time.Now().Add(time.Minute).Unix()
	tok := mintExportToken(secret, "my-deck", exp)

	if !validExportToken(secret, "my-deck", tok, exp) {
		t.Fatal("freshly minted token should validate")
	}
}

func TestExportToken_Rejections(t *testing.T) {
	const secret = "test-secret"
	exp := time.Now().Add(time.Minute).Unix()
	tok := mintExportToken(secret, "my-deck", exp)

	cases := []struct {
		name              string
		secret, slug, tok string
		exp               int64
	}{
		{"wrong slug", secret, "other-deck", tok, exp},
		{"wrong secret", "nope", "my-deck", tok, exp},
		{"empty secret disables", "", "my-deck", tok, exp},
		{"empty token", secret, "my-deck", "", exp},
		{"expired", secret, "my-deck", mintExportToken(secret, "my-deck", time.Now().Add(-time.Second).Unix()), time.Now().Add(-time.Second).Unix()},
		{"tampered exp", secret, "my-deck", tok, exp + 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if validExportToken(c.secret, c.slug, c.tok, c.exp) {
				t.Errorf("%s: expected token to be rejected", c.name)
			}
		})
	}
}
