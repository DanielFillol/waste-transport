package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// ---- Grupo A: Invoice transaction + duplicate prevention ----

func TestInvoice_DuplicatePrevention(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("DupInvTenant"))

	resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 10.0,
		"unit":           "KG",
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "DupInvGen", "cnpj": "77.777.777/0001-77",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "DupInvRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	// Create and mark a collect
	resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
		"generator_id": gen["id"],
		"receiver_id":  rec["id"],
		"planned_date": time.Now().Format("2006-01-02"),
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var col map[string]interface{}
	decodeJSON(t, resp, &col)

	resp = doRequest(t, "PATCH", "/v1/collects/"+col["id"].(string), map[string]interface{}{
		"status": 2, "collected_quantity": 50.0, "collected_unit": "KG",
	}, token)
	mustStatus(t, resp, http.StatusOK)

	year := time.Now().Year()
	body := map[string]interface{}{
		"generator_id": gen["id"],
		"period_start": fmt.Sprintf("%d-01-01", year),
		"period_end":   fmt.Sprintf("%d-12-31", year),
	}

	t.Run("first generate succeeds", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/invoices/generate", body, token)
		mustStatus(t, resp, http.StatusCreated)
	})

	t.Run("second generate for same period returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/invoices/generate", body, token)
		mustStatus(t, resp, http.StatusBadRequest)

		var errBody map[string]interface{}
		decodeJSON(t, resp, &errBody)
		if errBody["error"] == nil {
			t.Fatal("expected error message in response")
		}
	})
}

// ---- Grupo B: collected_at, notes, due_date ----

func TestCollects_CollectedAtAndNotes(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("CollectedAtTenant"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "CAGen", "cnpj": "88.888.888/0001-88",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "CARec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	t.Run("create collect with notes", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen["id"],
			"receiver_id":  rec["id"],
			"planned_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			"notes":        "Observação de campo",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var col map[string]interface{}
		decodeJSON(t, resp, &col)
		if col["notes"].(string) != "Observação de campo" {
			t.Errorf("expected notes='Observação de campo', got %v", col["notes"])
		}
	})

	t.Run("collected_at is set when status changes to collected", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen["id"],
			"receiver_id":  rec["id"],
			"planned_date": time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var col map[string]interface{}
		decodeJSON(t, resp, &col)
		colID := col["id"].(string)

		if col["collected_at"] != nil {
			t.Error("collected_at should be nil for planned collect")
		}

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 2, "collected_quantity": 30.0, "collected_unit": "KG",
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)

		if updated["collected_at"] == nil {
			t.Error("expected collected_at to be set after marking as collected")
		}
	})

	t.Run("update notes via PATCH", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen["id"],
			"receiver_id":  rec["id"],
			"planned_date": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var col map[string]interface{}
		decodeJSON(t, resp, &col)
		colID := col["id"].(string)

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"notes": "Nota atualizada",
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["notes"].(string) != "Nota atualizada" {
			t.Errorf("expected notes='Nota atualizada', got %v", updated["notes"])
		}
	})
}

func TestInvoice_DueDate(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("DueDateTenant"))

	resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 5.0,
		"unit":           "KG",
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "DDGen", "cnpj": "99.999.999/0001-99",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "DDRec"}, token)
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
		"period_start": fmt.Sprintf("%d-01-01", year),
		"period_end":   fmt.Sprintf("%d-12-31", year),
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var inv map[string]interface{}
	decodeJSON(t, resp, &inv)
	invID := inv["id"].(string)

	t.Run("issue invoice sets due_date with default 30 days", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/financial/invoices/"+invID+"/issue", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var issued map[string]interface{}
		decodeJSON(t, resp, &issued)

		if issued["due_date"] == nil {
			t.Fatal("expected due_date to be set after issuing invoice")
		}
		if issued["issued_at"] == nil {
			t.Error("expected issued_at to be set")
		}
	})

	t.Run("issue invoice with custom due_days", func(t *testing.T) {
		// Create another invoice in a different tenant to test custom due_days
		token2, _ := setupTenant(t, uniqueName("CustomDueTenant"))

		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"price_per_unit": 5.0,
			"unit":           "KG",
		}, token2)
		mustStatus(t, resp, http.StatusCreated)

		resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name": "CDGen", "cnpj": "11.222.333/0001-00",
		}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var gen2 map[string]interface{}
		decodeJSON(t, resp, &gen2)

		resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "CDRec"}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var rec2 map[string]interface{}
		decodeJSON(t, resp, &rec2)

		resp = doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen2["id"],
			"receiver_id":  rec2["id"],
			"planned_date": time.Now().Format("2006-01-02"),
		}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var col2 map[string]interface{}
		decodeJSON(t, resp, &col2)

		resp = doRequest(t, "PATCH", "/v1/collects/"+col2["id"].(string), map[string]interface{}{
			"status": 2, "collected_quantity": 10.0, "collected_unit": "KG",
		}, token2)
		mustStatus(t, resp, http.StatusOK)

		resp = doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": gen2["id"],
			"period_start": fmt.Sprintf("%d-01-01", year),
			"period_end":   fmt.Sprintf("%d-12-31", year),
		}, token2)
		mustStatus(t, resp, http.StatusCreated)
		var inv2 map[string]interface{}
		decodeJSON(t, resp, &inv2)

		resp = doRequest(t, "PATCH", "/v1/financial/invoices/"+inv2["id"].(string)+"/issue",
			map[string]interface{}{"due_days": 60}, token2)
		mustStatus(t, resp, http.StatusOK)

		var issued2 map[string]interface{}
		decodeJSON(t, resp, &issued2)
		if issued2["due_date"] == nil {
			t.Fatal("expected due_date to be set with custom due_days=60")
		}
	})
}

// ---- Grupo C: Active flag on Generator and Receiver ----

func TestGenerators_ActiveFlag(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("GenActiveTenant"))

	// Create two generators
	for _, name := range []string{"Gen Ativo A", "Gen Ativo B"} {
		resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name": name,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var g map[string]interface{}
		decodeJSON(t, resp, &g)
		// verify active=true by default
		if g["active"] != true {
			t.Errorf("expected new generator to be active=true, got %v", g["active"])
		}
	}

	// Get second generator to deactivate
	resp := doRequest(t, "GET", "/v1/generators", nil, token)
	mustStatus(t, resp, http.StatusOK)
	var list map[string]interface{}
	decodeJSON(t, resp, &list)
	items := list["data"].([]interface{})
	if len(items) < 2 {
		t.Fatal("expected at least 2 generators")
	}
	genID := items[0].(map[string]interface{})["id"].(string)

	t.Run("deactivate generator", func(t *testing.T) {
		falseVal := false
		resp := doRequest(t, "PATCH", "/v1/generators/"+genID, map[string]interface{}{
			"active": falseVal,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["active"] != false {
			t.Errorf("expected active=false, got %v", updated["active"])
		}
	})

	t.Run("filter active=true hides deactivated", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators?active=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		for _, item := range items {
			g := item.(map[string]interface{})
			if g["id"].(string) == genID {
				t.Error("deactivated generator should not appear in active=true filter")
			}
		}
	})

	t.Run("filter active=false shows only inactive", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators?active=false", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		items := result["data"].([]interface{})
		if len(items) != 1 {
			t.Errorf("expected 1 inactive generator, got %d", len(items))
		}
		if items[0].(map[string]interface{})["id"].(string) != genID {
			t.Error("wrong generator returned for active=false filter")
		}
	})

	t.Run("no filter returns all", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		total := result["total"].(float64)
		if int(total) != 2 {
			t.Errorf("expected 2 generators without filter, got %v", total)
		}
	})

	t.Run("reactivate generator", func(t *testing.T) {
		trueVal := true
		resp := doRequest(t, "PATCH", "/v1/generators/"+genID, map[string]interface{}{
			"active": trueVal,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["active"] != true {
			t.Errorf("expected active=true after reactivation, got %v", updated["active"])
		}
	})
}

func TestReceivers_ActiveFlag(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("RecActiveTenant"))

	resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
		"name": "Receptor Inativo",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)
	recID := rec["id"].(string)

	if rec["active"] != true {
		t.Errorf("expected new receiver to be active=true, got %v", rec["active"])
	}

	t.Run("deactivate receiver", func(t *testing.T) {
		falseVal := false
		resp := doRequest(t, "PATCH", "/v1/receivers/"+recID, map[string]interface{}{
			"active": falseVal,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["active"] != false {
			t.Errorf("expected active=false, got %v", updated["active"])
		}
	})

	t.Run("filter active=true hides inactive receiver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/receivers?active=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		for _, item := range result["data"].([]interface{}) {
			r := item.(map[string]interface{})
			if r["id"].(string) == recID {
				t.Error("inactive receiver should not appear in active=true filter")
			}
		}
	})
}

// ---- Grupo D: Route → Generate Collects ----

func TestRoutes_GenerateCollects(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("GenColsTenant"))

	// Create route with material and packaging
	resp := doRequest(t, "POST", "/v1/routes", map[string]interface{}{
		"name":        "Rota Geração",
		"week_day":    1,
		"week_number": 1,
		"material_id": 1,
		"packaging_id": 1,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var route map[string]interface{}
	decodeJSON(t, resp, &route)
	routeID := route["id"].(string)

	// Create 3 generators
	var genIDs []string
	for i := 0; i < 3; i++ {
		resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name": fmt.Sprintf("Gerador %d", i),
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var g map[string]interface{}
		decodeJSON(t, resp, &g)
		genIDs = append(genIDs, g["id"].(string))
	}

	// Create receiver
	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "Receptor Geração"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	targetDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	t.Run("generates one collect per generator", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/routes/"+routeID+"/generate-collects", map[string]interface{}{
			"target_date":   targetDate,
			"generator_ids": genIDs,
			"receiver_id":   rec["id"],
		}, token)
		mustStatus(t, resp, http.StatusCreated)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)

		created := result["created"].(float64)
		if int(created) != 3 {
			t.Errorf("expected 3 collects created, got %v", created)
		}
		collects := result["collects"].([]interface{})
		if len(collects) != 3 {
			t.Fatalf("expected 3 collects in response, got %d", len(collects))
		}

		// Verify each inherits route fields
		for _, item := range collects {
			c := item.(map[string]interface{})
			if c["route_id"].(string) != routeID {
				t.Errorf("expected route_id=%s, got %v", routeID, c["route_id"])
			}
			if c["status"].(float64) != 1 {
				t.Errorf("expected status=1 (planned), got %v", c["status"])
			}
			if c["material_id"] == nil {
				t.Error("expected material_id inherited from route")
			}
			if c["packaging_id"] == nil {
				t.Error("expected packaging_id inherited from route")
			}
		}
	})

	t.Run("invalid route id returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/routes/not-a-uuid/generate-collects", map[string]interface{}{
			"target_date":   targetDate,
			"generator_ids": genIDs,
			"receiver_id":   rec["id"],
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("non-existent route returns 404", func(t *testing.T) {
		fakeID := "00000000-0000-0000-0000-000000000001"
		resp := doRequest(t, "POST", "/v1/routes/"+fakeID+"/generate-collects", map[string]interface{}{
			"target_date":   targetDate,
			"generator_ids": genIDs,
			"receiver_id":   rec["id"],
		}, token)
		mustStatus(t, resp, http.StatusNotFound)
	})

	t.Run("empty generator_ids returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/routes/"+routeID+"/generate-collects", map[string]interface{}{
			"target_date":   targetDate,
			"generator_ids": []string{},
			"receiver_id":   rec["id"],
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("generated collects appear in collect listing with route filter", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?route_id="+routeID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		total := result["total"].(float64)
		if int(total) != 3 {
			t.Errorf("expected 3 collects for route filter, got %v", total)
		}
	})
}
