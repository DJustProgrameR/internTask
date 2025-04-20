// Package integration это интеграционное тестирование
package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// FullFlowTest полный сценарий интеграционного теста
func FullFlowTest(t *testing.T, app *fiber.App) {
	t.Run("full flow", func(t *testing.T) {

		registerModer(t, app)
		moderToken := loginModer(t, app)
		employeeToken := dummyLoginEmployee(t, app)

		t.Logf("got tokens")

		pvzID := createPVZ(t, app, moderToken)

		t.Logf("created pvz")

		_ = createReception(t, app, employeeToken, pvzID)

		t.Logf("created reception")

		for i := 0; i < 50; i++ {
			createProduct(t, app, employeeToken, pvzID)
		}

		t.Logf("added products")

		closeReception(t, app, employeeToken, pvzID)

		t.Logf("closed reception")
	})
}

func dummyLoginModer(t *testing.T, app *fiber.App) string {
	loginReq := map[string]string{
		"role": "moderator",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	token := string(bodyBytes)
	return token
}

func registerModer(t *testing.T, app *fiber.App) {
	loginReq := map[string]string{
		"email":    "vl@mail.ru",
		"password": "123456789",
		"role":     "moderator",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	var result struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	assert.Equal(t, "vl@mail.ru", result.Email)

	assert.Equal(t, "moderator", result.Role)
}

func loginModer(t *testing.T, app *fiber.App) string {
	loginReq := map[string]string{
		"email":    "vl@mail.ru",
		"password": "123456789",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	token := string(bodyBytes)
	return token
}

func dummyLoginEmployee(t *testing.T, app *fiber.App) string {
	loginReq := map[string]string{
		"role": "employee",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	token := string(bodyBytes)
	return token
}

func createPVZ(t *testing.T, app *fiber.App, token string) string {
	reqBody := map[string]string{
		"city": "Москва",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to create PVZ: %v", err)
	}

	var result struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	assert.Equal(t, "", result.Message)

	return result.ID
}

func createReception(t *testing.T, app *fiber.App, token string, pvzID string) string {
	reqBody := map[string]string{
		"pvzId": pvzID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/receptions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	assert.Equal(t, "", result.Message)

	return result.ID
}

func createProduct(t *testing.T, app *fiber.App, token string, pvzID string) string {
	productTypes := []string{"электроника", "одежда", "обувь"}
	productType := productTypes[time.Now().UnixNano()%3]

	reqBody := map[string]string{
		"type":  productType,
		"pvzId": pvzID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	assert.Equal(t, "", result.Message)

	return result.ID
}

func closeReception(t *testing.T, app *fiber.App, token string, pvzID string) {
	req := httptest.NewRequest("POST", "/pvz/"+pvzID+"/close_last_reception", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	assert.Equal(t, "", result.Message)
}
