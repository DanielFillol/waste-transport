package e2e

import (
	"net/http"
	"testing"
	"time"
)

func TestAuditLog_CreateGenerator(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditGen"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "AuditedGen", "cnpj": "11.111.111/0001-11",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)
	genID := gen["id"].(string)

	t.Run("create is logged", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=generator&entity_id="+genID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		data := result["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one audit log entry for the create")
		}
		entry := data[0].(map[string]interface{})
		if entry["action"] != "create" {
			t.Fatalf("expected action=create, got %v", entry["action"])
		}
		if entry["entity_type"] != "generator" {
			t.Fatalf("expected entity_type=generator, got %v", entry["entity_type"])
		}
		if entry["entity_id"] != genID {
			t.Fatalf("expected entity_id=%s, got %v", genID, entry["entity_id"])
		}
		if entry["actor_id"] == nil {
			t.Fatal("expected actor_id to be set")
		}
	})
}

func TestAuditLog_UpdateAndDelete(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditUpd"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "UpdGen", "cnpj": "22.222.222/0001-22",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)
	genID := gen["id"].(string)

	resp = doRequest(t, "PATCH", "/v1/generators/"+genID, map[string]interface{}{
		"name": "UpdGenRenamed",
	}, token)
	mustStatus(t, resp, http.StatusOK)

	resp = doRequest(t, "DELETE", "/v1/generators/"+genID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("all actions are logged in order", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=generator&entity_id="+genID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		data := result["data"].([]interface{})
		if len(data) < 3 {
			t.Fatalf("expected 3 audit log entries (create, update, delete), got %d", len(data))
		}

		// Entries are ordered by created_at DESC, so newest first.
		actions := make([]string, len(data))
		for i, item := range data {
			actions[i] = item.(map[string]interface{})["action"].(string)
		}
		if actions[0] != "delete" {
			t.Fatalf("expected first entry to be delete, got %s", actions[0])
		}
		if actions[1] != "update" {
			t.Fatalf("expected second entry to be update, got %s", actions[1])
		}
		if actions[2] != "create" {
			t.Fatalf("expected third entry to be create, got %s", actions[2])
		}
	})
}

func TestAuditLog_FilterByAction(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditFilter"))

	resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
		"plate": "AUD-0001", "model": "AuditTruck", "year": 2023,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var trk map[string]interface{}
	decodeJSON(t, resp, &trk)
	trkID := trk["id"].(string)

	resp = doRequest(t, "PATCH", "/v1/trucks/"+trkID, map[string]interface{}{
		"model": "AuditTruckV2",
	}, token)
	mustStatus(t, resp, http.StatusOK)

	t.Run("filter by action=create returns only creates", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=truck&entity_id="+trkID+"&action=create", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		data := result["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("expected exactly 1 create entry, got %d", len(data))
		}
		if data[0].(map[string]interface{})["action"] != "create" {
			t.Fatal("expected action=create")
		}
	})

	t.Run("filter by action=update returns only updates", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=truck&entity_id="+trkID+"&action=update", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		data := result["data"].([]interface{})
		if len(data) != 1 {
			t.Fatalf("expected exactly 1 update entry, got %d", len(data))
		}
	})
}

func TestAuditLog_PayloadContainsRequestBody(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditPayload"))

	resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
		"name": "PayloadRec", "cnpj": "33.333.333/0001-33",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)
	recID := rec["id"].(string)

	resp = doRequest(t, "GET", "/v1/audit-logs?entity_type=receiver&entity_id="+recID+"&action=create", nil, token)
	mustStatus(t, resp, http.StatusOK)
	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	data := result["data"].([]interface{})
	if len(data) == 0 {
		t.Fatal("expected audit log entry")
	}
	payload := data[0].(map[string]interface{})["payload"].(string)
	if payload == "" {
		t.Fatal("expected non-empty payload")
	}
	// The payload should contain the request body fields
	if len(payload) < 5 {
		t.Fatalf("payload looks too short: %s", payload)
	}
}

func TestAuditLog_TenantIsolation(t *testing.T) {
	token1, _ := setupTenant(t, uniqueName("AuditTenant1"))
	token2, _ := setupTenant(t, uniqueName("AuditTenant2"))

	// Create a generator as tenant1
	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "T1Gen", "cnpj": "44.444.444/0001-44",
	}, token1)
	mustStatus(t, resp, http.StatusCreated)

	// Tenant2 should see zero audit logs (their own tenant only)
	resp = doRequest(t, "GET", "/v1/audit-logs?entity_type=generator", nil, token2)
	mustStatus(t, resp, http.StatusOK)
	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	data := result["data"].([]interface{})
	if len(data) != 0 {
		t.Fatalf("tenant2 should see 0 audit logs, got %d", len(data))
	}
}

func TestAuditLog_DateRangeFilter(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditDate"))

	resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "DateDriver",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var drv map[string]interface{}
	decodeJSON(t, resp, &drv)
	drvID := drv["id"].(string)

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	t.Run("date range covering today returns entry", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=driver&entity_id="+drvID+"&date_from="+yesterday+"&date_to="+tomorrow, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if len(result["data"].([]interface{})) == 0 {
			t.Fatal("expected entry in date range")
		}
	})

	t.Run("date range before today returns nothing", func(t *testing.T) {
		pastStart := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
		pastEnd := yesterday
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=driver&entity_id="+drvID+"&date_from="+pastStart+"&date_to="+pastEnd, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if len(result["data"].([]interface{})) != 0 {
			t.Fatal("expected no entries before today")
		}
	})

	_ = today
}

func TestAuditLog_InvoiceActions(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("AuditInv"))

	resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 8.0, "unit": "KG",
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "AuditInvGen", "cnpj": "55.555.555/0001-55",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "AuditInvRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": time.Now().Format("2006-01-02"),
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var col map[string]interface{}
	decodeJSON(t, resp, &col)

	resp = doRequest(t, "PATCH", "/v1/collects/"+col["id"].(string), map[string]interface{}{
		"status": 2, "collected_quantity": 20.0, "collected_unit": "KG",
	}, token)
	mustStatus(t, resp, http.StatusOK)

	year := time.Now().Year()
	resp = doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
		"generator_id": gen["id"],
		"period_start": time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		"period_end":   time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var inv map[string]interface{}
	decodeJSON(t, resp, &inv)
	invID := inv["id"].(string)

	resp = doRequest(t, "PATCH", "/v1/financial/invoices/"+invID+"/issue", nil, token)
	mustStatus(t, resp, http.StatusOK)

	resp = doRequest(t, "PATCH", "/v1/financial/invoices/"+invID+"/paid", nil, token)
	mustStatus(t, resp, http.StatusOK)

	t.Run("invoice lifecycle is fully logged", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/audit-logs?entity_type=invoice&entity_id="+invID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		data := result["data"].([]interface{})
		if len(data) < 3 {
			t.Fatalf("expected at least 3 entries (generate, issue, mark_paid), got %d", len(data))
		}

		actions := make(map[string]bool)
		for _, item := range data {
			actions[item.(map[string]interface{})["action"].(string)] = true
		}
		for _, expected := range []string{"generate", "issue", "mark_paid"} {
			if !actions[expected] {
				t.Fatalf("expected action %q in audit log, got actions: %v", expected, actions)
			}
		}
	})
}
