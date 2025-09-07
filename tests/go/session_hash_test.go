package gotests

import (
    "testing"
    "anshumanbiswas.com/blog/models"
    "golang.org/x/crypto/bcrypt"
)

func TestGenerateHashedToken_Verify(t *testing.T) {
    ss := models.SessionService{}
    token := "test-token-123"

    hash, err := ss.GenerateHashedToken(token)
    if err != nil {
        t.Fatalf("GenerateHashedToken returned error: %v", err)
    }
    if hash == "" {
        t.Fatalf("expected non-empty hash")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)); err != nil {
        t.Fatalf("hashed token does not verify against original: %v", err)
    }
}

func TestGenerateHashedToken_IsSalted(t *testing.T) {
    ss := models.SessionService{}
    token := "same-token"
    h1, err := ss.GenerateHashedToken(token)
    if err != nil { t.Fatal(err) }
    h2, err := ss.GenerateHashedToken(token)
    if err != nil { t.Fatal(err) }
    if h1 == h2 {
        t.Fatalf("expected different hashes for same token due to salting")
    }
}

