package e2e

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func uniqueName(base string) string {
	return fmt.Sprintf("%s-%d", base, rand.Intn(999999))
}

func TestAlerts(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AlertsTenant"))

	// Create a driver with CNH expiring in 10 days → should trigger alert
	expiryDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name":            "João Alerta",
		"cnh_number":      "99887766",
		"cnh_category":    "B",
		"cnh_expiry_date": expiryDate,
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	var driver map[string]interface{}
	decodeJSON(t, resp, &driver)
	driverID := driver["id"].(string)

	t.Run("list all alerts returns CNH alert", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/alerts", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		items := result["data"].([]interface{})
		if len(items) == 0 {
			t.Fatal("expected at least one alert, got none")
		}

		found := false
		for _, item := range items {
			a := item.(map[string]interface{})
			if a["type"] == "cnh_expiry" {
				found = true
				if a["read"] != false {
					t.Error("new alert should be unread")
				}
			}
		}
		if !found {
			t.Error("expected cnh_expiry alert, not found")
		}
	})

	t.Run("filter unread-only returns the alert", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/alerts?unread=true", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) == 0 {
			t.Fatal("expected unread alert, got none")
		}
	})

	t.Run("mark alert as read", func(t *testing.T) {
		// Get alert id
		resp := doRequest(t, "GET", "/v1/alerts", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		alertID := items[0].(map[string]interface{})["id"].(string)

		// Mark read
		resp = doRequest(t, "PATCH", fmt.Sprintf("/v1/alerts/%s/read", alertID), nil, token)
		mustStatus(t, resp, http.StatusOK)
		var readResult map[string]interface{}
		decodeJSON(t, resp, &readResult)
		if readResult["read"] != true {
			t.Error("expected read=true")
		}

		// After marking read, unread filter should return 0
		resp = doRequest(t, "GET", "/v1/alerts?unread=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result2 map[string]interface{}
		decodeJSON(t, resp, &result2)
		items2 := result2["data"].([]interface{})
		if len(items2) != 0 {
			t.Errorf("expected 0 unread alerts after marking read, got %d", len(items2))
		}
	})

	t.Run("driver with non-expiring CNH does not trigger alert", func(t *testing.T) {
		// Clean any alerts first by using a fresh tenant
		token2, _ := setupTenant(t, uniqueName("NoAlertTenant"))
		futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
			"name":            "Motorista Seguro",
			"cnh_number":      "00112233",
			"cnh_expiry_date": futureDate,
		}, token2)
		mustStatus(t, resp, http.StatusCreated)

		resp = doRequest(t, "GET", "/v1/alerts", nil, token2)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if len(result["data"].([]interface{})) != 0 {
			t.Error("expected no alerts for far-future CNH expiry")
		}
	})

	t.Run("receiver with expiring license triggers alert", func(t *testing.T) {
		token3, _ := setupTenant(t, uniqueName("ReceiverAlertTenant"))
		expiryDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
		resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
			"name":           "Receptor Vencendo",
			"license_number": "LIC-999",
			"license_expiry": expiryDate,
		}, token3)
		mustStatus(t, resp, http.StatusCreated)

		resp = doRequest(t, "GET", "/v1/alerts", nil, token3)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) == 0 {
			t.Fatal("expected license_expiry alert, got none")
		}
		alertType := items[0].(map[string]interface{})["type"].(string)
		if alertType != "license_expiry" {
			t.Errorf("expected license_expiry alert, got %s", alertType)
		}
	})

	t.Run("update driver CNH to future date removes alert", func(t *testing.T) {
		// Re-read alerts for original tenant — should have one alert for the driver
		resp := doRequest(t, "GET", "/v1/alerts", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var before map[string]interface{}
		decodeJSON(t, resp, &before)
		countBefore := len(before["data"].([]interface{}))

		// Update CNH to far future
		futureDate := time.Now().AddDate(2, 0, 0).Format("2006-01-02")
		resp = doRequest(t, "PATCH", fmt.Sprintf("/v1/drivers/%s", driverID), map[string]interface{}{
			"cnh_expiry_date": futureDate,
		}, token)
		mustStatus(t, resp, http.StatusOK)

		// Alert should be gone
		resp = doRequest(t, "GET", "/v1/alerts", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var after map[string]interface{}
		decodeJSON(t, resp, &after)
		countAfter := len(after["data"].([]interface{}))
		if countAfter >= countBefore {
			t.Errorf("expected fewer alerts after CNH update (before=%d, after=%d)", countBefore, countAfter)
		}
	})

	t.Run("invalid alert id returns 400", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/alerts/not-a-uuid/read", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

func TestCollects_BulkCancel(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("BulkCancelTenant"))

	// Create generator and receiver
	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "Gerador Bulk", "cnpj": "11.111.111/0001-11",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)
	genID := gen["id"].(string)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
		"name": "Receptor Bulk",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)
	recID := rec["id"].(string)

	// Create 3 collects
	var ids []string
	for i := 0; i < 3; i++ {
		resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"planned_date": time.Now().AddDate(0, 0, i+1).Format("2006-01-02"),
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var col map[string]interface{}
		decodeJSON(t, resp, &col)
		ids = append(ids, col["id"].(string))
	}

	t.Run("bulk cancel all three collects", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects/bulk-cancel", map[string]interface{}{
			"ids": ids,
		}, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		cancelled := result["cancelled"].(float64)
		if int(cancelled) != 3 {
			t.Errorf("expected 3 cancelled, got %v", cancelled)
		}
	})

	t.Run("verify all collects are cancelled", func(t *testing.T) {
		for _, id := range ids {
			resp := doRequest(t, "GET", fmt.Sprintf("/v1/collects/%s", id), nil, token)
			mustStatus(t, resp, http.StatusOK)
			var col map[string]interface{}
			decodeJSON(t, resp, &col)
			status := col["status"].(float64)
			if int(status) != 3 { // CollectStatusCancelled = 3
				t.Errorf("collect %s expected status 3 (cancelled), got %v", id, status)
			}
		}
	})

	t.Run("empty ids rejected", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects/bulk-cancel", map[string]interface{}{
			"ids": []string{},
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

func TestRoutes_Search(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("RouteSearchTenant"))

	names := []string{"Rota Norte", "Rota Sul", "Coleta Especial"}
	for _, name := range names {
		resp := doRequest(t, "POST", "/v1/routes", map[string]interface{}{
			"name":        name,
			"week_day":    1,
			"week_number": 1,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	}

	t.Run("search by partial name returns matching routes", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes?search=Rota", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 2 {
			t.Errorf("expected 2 routes matching 'Rota', got %d", len(items))
		}
	})

	t.Run("search with no match returns empty", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes?search=Inexistente", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 0 {
			t.Errorf("expected 0 results, got %d", len(items))
		}
	})

	t.Run("no search returns all routes", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		total := result["total"].(float64)
		if int(total) != 3 {
			t.Errorf("expected 3 total routes, got %v", total)
		}
	})
}

func TestInvalidInput_Returns400(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ValidationTenant"))

	t.Run("invalid UUID in driver GET returns 400", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/drivers/not-a-uuid", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid UUID in receiver GET returns 400", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/receivers/not-a-uuid", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid UUID in route GET returns 400", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes/not-a-uuid", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid date format in driver create returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
			"name":            "Test Driver",
			"cnh_expiry_date": "31-12-2025", // wrong format
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid date format in receiver create returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
			"name":           "Test Receiver",
			"license_expiry": "31/12/2025", // wrong format
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("create collect with invalid planned_date returns 400", func(t *testing.T) {
		// Need valid generator and receiver UUIDs first
		resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name": "Gen", "cnpj": "22.222.222/0001-22",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var gen map[string]interface{}
		decodeJSON(t, resp, &gen)

		resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "Rec"}, token)
		mustStatus(t, resp, http.StatusCreated)
		var rec map[string]interface{}
		decodeJSON(t, resp, &rec)

		resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen["id"],
			"receiver_id":  rec["id"],
			"planned_date": "not-a-date",
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("bulk-cancel with missing ids returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects/bulk-cancel", map[string]interface{}{}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}
