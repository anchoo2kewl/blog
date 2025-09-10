package gotests

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"anshumanbiswas.com/blog/models"
	_ "github.com/lib/pq"
)

func setupTestDB() (*sql.DB, error) {
	// Read environment variables (same as main app)
	dbHost := os.Getenv("PG_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("PG_PORT")
	if dbPort == "" {
		dbPort = "5433"
	}
	dbUser := os.Getenv("PG_USER")
	if dbUser == "" {
		dbUser = "blog"
	}
	dbPassword := os.Getenv("PG_PASSWORD")
	if dbPassword == "" {
		dbPassword = "1234" // fallback for tests
	}
	dbName := os.Getenv("PG_DB")
	if dbName == "" {
		dbName = "blog"
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test database connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func TestAPITokenCreation(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Initialize API token service
	apiTokenService := &models.APITokenService{DB: db}

	// Check if user ID 1 exists
	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", 1).Scan(&userExists)
	if err != nil {
		t.Fatalf("Failed to check if user exists: %v", err)
	}

	if !userExists {
		t.Skip("User ID 1 does not exist, skipping test")
	}

	// Test creating API token for user ID 1
	tokenName := "test-token-integration"
	t.Logf("Creating API token '%s' for user ID 1...", tokenName)
	
	token, err := apiTokenService.Create(1, tokenName, nil)
	if err != nil {
		t.Fatalf("Failed to create API token: %v", err)
	}

	// Verify token was created
	if token == nil {
		t.Fatal("Token is nil")
	}

	if token.ID == 0 {
		t.Error("Token ID should not be 0")
	}

	if token.Name != tokenName {
		t.Errorf("Expected token name '%s', got '%s'", tokenName, token.Name)
	}

	if token.UserID != 1 {
		t.Errorf("Expected user ID 1, got %d", token.UserID)
	}

	if token.Token == "" {
		t.Error("Token string should not be empty")
	}

	if !token.IsActive {
		t.Error("Token should be active")
	}

	t.Logf("✅ API Token created successfully!")
	t.Logf("   Token ID: %d", token.ID)
	t.Logf("   Token Name: %s", token.Name)
	t.Logf("   Token Length: %d characters", len(token.Token))
	t.Logf("   Created At: %s", token.CreatedAt)

	// Clean up: Delete the test token
	err = apiTokenService.Delete(token.ID, 1)
	if err != nil {
		t.Errorf("Failed to clean up test token: %v", err)
	}
}

func TestAPITokenValidation(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Initialize API token service
	apiTokenService := &models.APITokenService{DB: db}

	// Check if user ID 1 exists
	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", 1).Scan(&userExists)
	if err != nil {
		t.Fatalf("Failed to check if user exists: %v", err)
	}

	if !userExists {
		t.Skip("User ID 1 does not exist, skipping test")
	}

	// Create a test token
	tokenName := "test-validation-token"
	token, err := apiTokenService.Create(1, tokenName, nil)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// Test validation with the token
	user, err := apiTokenService.ValidateToken(token.Token)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if user == nil {
		t.Error("User should not be nil for valid token")
	} else {
		if user.UserID != 1 {
			t.Errorf("Expected user ID 1, got %d", user.UserID)
		}
	}

	// Test validation with invalid token
	_, err = apiTokenService.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Should fail to validate invalid token")
	}

	// Clean up
	err = apiTokenService.Delete(token.ID, 1)
	if err != nil {
		t.Errorf("Failed to clean up test token: %v", err)
	}

	t.Log("✅ API Token validation test completed successfully!")
}