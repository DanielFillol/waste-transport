package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// ---- GET /v1/me ----

func TestAuth_Me(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("MeTenant"))

	t.Run("returns authenticated user info", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/me", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		if body["username"].(string) != "admin" {
			t.Fatalf("expected username=admin, got %s", body["username"])
		}
		if body["role"].(string) != "admin" {
			t.Fatalf("expected role=admin, got %s", body["role"])
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/me", nil, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})
}

// ---- POST /v1/auth/refresh ----

func TestAuth_Refresh(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("RefreshTenant"))

	t.Run("returns new token", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/refresh", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		newToken, ok := body["token"].(string)
		if !ok || newToken == "" {
			t.Fatal("expected non-empty token in refresh response")
		}
		user, ok := body["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in refresh response")
		}
		if user["username"].(string) != "admin" {
			t.Fatalf("expected username=admin, got %s", user["username"])
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/refresh", nil, "invalid-token")
		mustStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("no token returns 401", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/refresh", nil, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})
}

// ---- GET /v1/dashboard ----

func TestDashboard_Get(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("DashTenant"))

	t.Run("returns dashboard data", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/dashboard", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		if _, ok := body["collects"]; !ok {
			t.Fatal("expected collects key in dashboard response")
		}
		if _, ok := body["alerts_unread"]; !ok {
			t.Fatal("expected alerts_unread key in dashboard response")
		}
		if _, ok := body["invoices"]; !ok {
			t.Fatal("expected invoices key in dashboard response")
		}
		if _, ok := body["this_month"]; !ok {
			t.Fatal("expected this_month key in dashboard response")
		}
	})

	t.Run("collects counts are accurate", func(t *testing.T) {
		token2, _ := setupTenant(t, uniqueName("DashCountTenant"))

		// Create generator and receiver
		resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name": "DashGen", "cnpj": "33.333.333/0001-33",
		}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var gen map[string]interface{}
		decodeJSON(t, resp, &gen)

		resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "DashRec"}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var rec map[string]interface{}
		decodeJSON(t, resp, &rec)

		// Create 2 planned collects
		for i := 0; i < 2; i++ {
			resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
				"generator_id": gen["id"],
				"receiver_id":  rec["id"],
				"planned_date": time.Now().AddDate(0, 0, i+1).Format("2006-01-02"),
			}, token2)
			mustStatus(t, resp, http.StatusCreated)
		}

		resp = doRequest(t, "GET", "/v1/dashboard", nil, token2)
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		collects := body["collects"].(map[string]interface{})
		planned := collects["planned"].(float64)
		if int(planned) != 2 {
			t.Errorf("expected 2 planned collects, got %v", planned)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/dashboard", nil, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})
}

// ---- GET /v1/financial/summary ----

func TestFinancial_Summary(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("SummaryTenant"))

	t.Run("missing period returns 400", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/summary", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("invalid date format returns 400", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/summary?period_start=01/01/2025&period_end=31/01/2025", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("returns summary with zero values when no data", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/summary?period_start=2025-01-01&period_end=2025-01-31", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		if _, ok := body["revenue"]; !ok {
			t.Fatal("expected revenue key in summary response")
		}
		if _, ok := body["truck_costs"]; !ok {
			t.Fatal("expected truck_costs key in summary response")
		}
		if _, ok := body["personnel_costs"]; !ok {
			t.Fatal("expected personnel_costs key in summary response")
		}
		if _, ok := body["gross_margin"]; !ok {
			t.Fatal("expected gross_margin key in summary response")
		}
	})
}

// ---- PATCH /v1/alerts/read-all ----

func TestAlerts_MarkAllRead(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ReadAllTenant"))

	// Generate two alerts by creating drivers with expiring CNHs
	for i := 0; i < 2; i++ {
		expiryDate := time.Now().AddDate(0, 0, i+5).Format("2006-01-02")
		resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
			"name":            fmt.Sprintf("Motorista %d", i),
			"cnh_number":      fmt.Sprintf("CNH-%d", i),
			"cnh_expiry_date": expiryDate,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	}

	t.Run("mark all alerts as read", func(t *testing.T) {
		// Verify there are unread alerts
		resp := doRequest(t, "GET", "/v1/alerts?unread=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var before map[string]interface{}
		decodeJSON(t, resp, &before)
		if len(before["data"].([]interface{})) == 0 {
			t.Fatal("expected at least 1 unread alert before mark-all-read")
		}

		// Mark all read
		resp = doRequest(t, "PATCH", "/v1/alerts/read-all", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["read"] != true {
			t.Error("expected read=true in response")
		}

		// Verify no unread alerts remain
		resp = doRequest(t, "GET", "/v1/alerts?unread=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var after map[string]interface{}
		decodeJSON(t, resp, &after)
		if len(after["data"].([]interface{})) != 0 {
			t.Error("expected 0 unread alerts after mark-all-read")
		}
	})
}

// ---- COLLECTED status validation ----

func TestCollects_CollectedValidation(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("CollectedValTenant"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "ColValGen", "cnpj": "44.444.444/0001-44",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "ColValRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var col map[string]interface{}
	decodeJSON(t, resp, &col)
	colID := col["id"].(string)

	t.Run("marking collected without quantity returns 400", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 2,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("marking collected without unit returns 400", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status":             2,
			"collected_quantity": 100.0,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("marking collected with both quantity and unit succeeds", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status":             2,
			"collected_quantity": 100.0,
			"collected_unit":     "KG",
		}, token)
		mustStatus(t, resp, http.StatusOK)
	})
}

// ---- Invoice number ----

func TestInvoice_Number(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("InvoiceNumTenant"))

	// Create pricing rule, generator, receiver, and collected collects
	resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 10.0,
		"unit":           "KG",
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "InvNumGen", "cnpj": "55.555.555/0001-55",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "InvNumRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	// Create and mark a collect as collected
	collectDate := time.Now().Format("2006-01-02")
	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": collectDate,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var col map[string]interface{}
	decodeJSON(t, resp, &col)

	resp = doRequest(t, "PATCH", "/v1/collects/"+col["id"].(string), map[string]interface{}{
		"status":             2,
		"collected_quantity": 50.0,
		"collected_unit":     "KG",
	}, token)
	mustStatus(t, resp, http.StatusOK)

	t.Run("generated invoice has invoice_number", func(t *testing.T) {
		year := time.Now().Year()
		resp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": gen["id"],
			"period_start": fmt.Sprintf("%d-01-01", year),
			"period_end":   fmt.Sprintf("%d-12-31", year),
		}, token)
		mustStatus(t, resp, http.StatusCreated)

		var inv map[string]interface{}
		decodeJSON(t, resp, &inv)

		num, ok := inv["invoice_number"].(string)
		if !ok || num == "" {
			t.Fatalf("expected non-empty invoice_number, got %v", inv["invoice_number"])
		}
		expectedPrefix := fmt.Sprintf("%d/", year)
		if len(num) < len(expectedPrefix) || num[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected invoice_number to start with %s, got %s", expectedPrefix, num)
		}
	})
}

// ---- Collect TruckID + material/packaging filters ----

func TestCollects_TruckAndFilters(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("CollectFilterTenant"))

	// Create a truck
	resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
		"plate": "ABC-1234",
		"model": "Volvo FH",
		"year":  2020,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var truck map[string]interface{}
	decodeJSON(t, resp, &truck)
	truckID := truck["id"].(string)

	// Create generator and receiver
	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "FilterGen", "cnpj": "66.666.666/0001-66",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "FilterRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	// Create collect with truck_id and material/packaging
	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		"truck_id":     truckID,
		"material_id":  1,
		"packaging_id": 1,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var col map[string]interface{}
	decodeJSON(t, resp, &col)

	// Create another collect without truck/material/packaging
	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	t.Run("collect stores truck_id", func(t *testing.T) {
		storedTruckID, ok := col["truck_id"].(string)
		if !ok || storedTruckID == "" {
			t.Fatalf("expected truck_id in created collect, got %v", col["truck_id"])
		}
		if storedTruckID != truckID {
			t.Errorf("expected truck_id=%s, got %s", truckID, storedTruckID)
		}
	})

	t.Run("filter by truck_id returns only matching collects", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?truck_id="+truckID, nil, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 1 {
			t.Errorf("expected 1 collect with truck_id filter, got %d", len(items))
		}
	})

	t.Run("filter by material_id returns only matching collects", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?material_id=1", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 1 {
			t.Errorf("expected 1 collect with material_id=1 filter, got %d", len(items))
		}
	})

	t.Run("filter by packaging_id returns only matching collects", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?packaging_id=1", nil, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 1 {
			t.Errorf("expected 1 collect with packaging_id=1 filter, got %d", len(items))
		}
	})

	t.Run("update collect truck_id", func(t *testing.T) {
		colID := col["id"].(string)

		// Create a second truck
		resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
			"plate": "XYZ-9999",
			"model": "Mercedes Atego",
			"year":  2021,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var truck2 map[string]interface{}
		decodeJSON(t, resp, &truck2)
		truck2ID := truck2["id"].(string)

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"truck_id": truck2ID,
		}, token)
		mustStatus(t, resp, http.StatusOK)

		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["truck_id"].(string) != truck2ID {
			t.Errorf("expected truck_id=%s, got %s", truck2ID, updated["truck_id"])
		}
	})
}
