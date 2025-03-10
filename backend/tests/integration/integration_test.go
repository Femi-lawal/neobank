package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// Test configuration
const (
	identityURL = "http://localhost:8081"
	ledgerURL   = "http://localhost:8082"
	paymentURL  = "http://localhost:8083"
	productURL  = "http://localhost:8084"
	cardURL     = "http://localhost:8085"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// TestIdentityServiceHealth tests the identity service health endpoint
func TestIdentityServiceHealth(t *testing.T) {
	resp, err := httpClient.Get(identityURL + "/health")
	if err != nil {
		t.Skipf("Identity service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestLedgerServiceHealth tests the ledger service health endpoint
func TestLedgerServiceHealth(t *testing.T) {
	resp, err := httpClient.Get(ledgerURL + "/health")
	if err != nil {
		t.Skipf("Ledger service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestPaymentServiceHealth tests the payment service health endpoint
func TestPaymentServiceHealth(t *testing.T) {
	resp, err := httpClient.Get(paymentURL + "/health")
	if err != nil {
		t.Skipf("Payment service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestProductServiceHealth tests the product service health endpoint
func TestProductServiceHealth(t *testing.T) {
	resp, err := httpClient.Get(productURL + "/health")
	if err != nil {
		t.Skipf("Product service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestCardServiceHealth tests the card service health endpoint
func TestCardServiceHealth(t *testing.T) {
	resp, err := httpClient.Get(cardURL + "/health")
	if err != nil {
		t.Skipf("Card service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestUserRegistrationAndLogin tests the full auth flow
func TestUserRegistrationAndLogin(t *testing.T) {
	// Skip if service not available
	resp, err := httpClient.Get(identityURL + "/health")
	if err != nil {
		t.Skipf("Identity service not available: %v", err)
	}
	resp.Body.Close()

	// Register user
	registerBody := map[string]string{
		"email":      "integration-test@example.com",
		"password":   "TestPass123!",
		"first_name": "Integration",
		"last_name":  "Test",
	}
	body, _ := json.Marshal(registerBody)

	resp, err = httpClient.Post(identityURL+"/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	defer resp.Body.Close()

	// Accept both 201 (created) and 409 (already exists)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status 201 or 409, got %d", resp.StatusCode)
	}

	// Login
	loginBody := map[string]string{
		"email":    "integration-test@example.com",
		"password": "TestPass123!",
	}
	body, _ = json.Marshal(loginBody)

	resp, err = httpClient.Post(identityURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for login, got %d", resp.StatusCode)
	}

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	if loginResp["token"] == nil {
		t.Error("expected token in login response")
	}
}

// TestProductListEndpoint tests the product listing endpoint
func TestProductListEndpoint(t *testing.T) {
	resp, err := httpClient.Get(productURL + "/api/v1/products")
	if err != nil {
		t.Skipf("Product service not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
