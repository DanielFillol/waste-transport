package e2e

import (
	"net/http"
	"testing"
)

// ---- Pricing Rules ----

func TestPricingRules_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "PricingCorp")
	var ruleID string

	t.Run("create rule - all three factors", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"collect_type":  "normal",
			"material_id":   1,
			"packaging_id":  1,
			"price_per_unit": 12.50,
			"unit":          "KG",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		ruleID = body["id"].(string)
		if body["price_per_unit"].(float64) != 12.50 {
			t.Fatalf("price_per_unit mismatch: %v", body["price_per_unit"])
		}
		if body["unit"].(string) != "KG" {
			t.Fatalf("unit mismatch: %s", body["unit"])
		}
	})

	t.Run("create rule - only material (less specific)", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"material_id":    1,
			"price_per_unit": 8.00,
			"unit":           "KG",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	})

	t.Run("create rule - no factors (catch-all)", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"price_per_unit": 5.00,
			"unit":           "KG",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	})

	t.Run("create rule with LITER unit", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"collect_type":  "special",
			"price_per_unit": 3.50,
			"unit":          "LITER",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	})

	t.Run("create rule with M3 unit", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"price_per_unit": 150.00,
			"unit":           "M3",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
	})

	t.Run("list rules - all active by default", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/pricing-rules", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected pricing rules")
		}
	})

	t.Run("get rule", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/pricing-rules/"+ruleID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["id"].(string) != ruleID {
			t.Fatal("id mismatch")
		}
	})

	t.Run("update rule price", func(t *testing.T) {
		resp := doRequest(t, "PUT", "/v1/financial/pricing-rules/"+ruleID, map[string]interface{}{
			"price_per_unit": 15.00,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["price_per_unit"].(float64) != 15.00 {
			t.Fatalf("price not updated: %v", body["price_per_unit"])
		}
	})

	t.Run("deactivate rule", func(t *testing.T) {
		falseVal := false
		resp := doRequest(t, "PUT", "/v1/financial/pricing-rules/"+ruleID, map[string]interface{}{
			"active": &falseVal,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["active"].(bool) != false {
			t.Fatal("expected rule to be inactive")
		}
	})

	t.Run("invalid unit rejected", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
			"price_per_unit": 10.00,
			"unit":           "GALLON",
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("delete rule", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/financial/pricing-rules/"+ruleID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})

	t.Run("tenant isolation on pricing rules", func(t *testing.T) {
		otherToken, _ := setupTenant(t, "OtherPricingCorp")
		resp := doRequest(t, "GET", "/v1/financial/pricing-rules", nil, otherToken)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) != 0 {
			t.Fatalf("expected 0 rules for new tenant, got %d", len(data))
		}
	})
}

// ---- Invoice Generation (full flow) ----

func TestInvoice_FullFlow(t *testing.T) {
	token, _ := setupTenant(t, "InvoiceCorp")

	// Create pricing rules (three levels of specificity)
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"collect_type":  "normal",
		"material_id":   2,
		"packaging_id":  1,
		"price_per_unit": 20.00,
		"unit":          "KG",
	}, token)
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"material_id":    2,
		"price_per_unit": 15.00,
		"unit":          "KG",
	}, token)
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 10.00,
		"unit":          "KG",
	}, token)

	// Create generator and receiver
	genResp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{"name": "Gen Invoice"}, token)
	var gen map[string]interface{}
	decodeJSON(t, genResp, &gen)
	genID := gen["id"].(string)

	recResp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "Rec Invoice"}, token)
	var rec map[string]interface{}
	decodeJSON(t, recResp, &rec)
	recID := rec["id"].(string)

	// Create and collect 3 collects in the period
	collectIDs := []string{}
	for i := 1; i <= 3; i++ {
		cResp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"material_id":  2,
			"packaging_id": 1,
			"collect_type": "normal",
			"planned_date": "2026-05-01",
		}, token)
		var c map[string]interface{}
		decodeJSON(t, cResp, &c)
		cID := c["id"].(string)
		collectIDs = append(collectIDs, cID)

		// Mark as collected with quantity
		doRequest(t, "PATCH", "/v1/collects/"+cID, map[string]interface{}{
			"status":             2,
			"collected_quantity": 100.0,
			"collected_unit":     "KG",
			"collected_weight":   100.0,
		}, token)
	}

	var invoiceID string

	t.Run("generate invoice", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": genID,
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
			"notes":        "Fatura maio 2026",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		invoiceID = body["id"].(string)

		if body["status"].(string) != "draft" {
			t.Fatalf("expected draft, got %s", body["status"])
		}
		// 3 collects × 100 KG × R$20 (most specific rule) = R$6000
		if body["total_amount"].(float64) != 6000.00 {
			t.Fatalf("expected total_amount 6000.00, got %v", body["total_amount"])
		}
	})

	t.Run("get invoice with items", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices/"+invoiceID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		items := body["items"].([]interface{})
		if len(items) != 3 {
			t.Fatalf("expected 3 invoice items, got %d", len(items))
		}
		// Each item: 100 KG × R$20 = R$2000
		for _, item := range items {
			it := item.(map[string]interface{})
			if it["total_price"].(float64) != 2000.00 {
				t.Fatalf("item total_price mismatch: %v", it["total_price"])
			}
			if it["unit"].(string) != "KG" {
				t.Fatalf("item unit mismatch: %s", it["unit"])
			}
		}
	})

	t.Run("list invoices filtered by generator", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices?generator_id="+genID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["total"].(float64) < 1 {
			t.Fatal("expected at least one invoice")
		}
	})

	t.Run("issue invoice", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/financial/invoices/"+invoiceID+"/issue", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["status"].(string) != "issued" {
			t.Fatalf("expected issued, got %s", body["status"])
		}
		if body["issued_at"] == nil {
			t.Fatal("expected issued_at to be set")
		}
	})

	t.Run("cannot issue an already issued invoice", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/financial/invoices/"+invoiceID+"/issue", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("mark invoice as paid", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/financial/invoices/"+invoiceID+"/paid", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["status"].(string) != "paid" {
			t.Fatalf("expected paid, got %s", body["status"])
		}
	})

	t.Run("cannot mark draft invoice as paid", func(t *testing.T) {
		// Create another generator and invoice
		g2Resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{"name": "Gen2 Invoice"}, token)
		var g2 map[string]interface{}
		decodeJSON(t, g2Resp, &g2)
		g2ID := g2["id"].(string)

		c := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": g2ID,
			"receiver_id":  recID,
			"material_id":  2,
			"packaging_id": 1,
			"collect_type": "normal",
			"planned_date": "2026-05-02",
		}, token)
		var cBody map[string]interface{}
		decodeJSON(t, c, &cBody)
		doRequest(t, "PATCH", "/v1/collects/"+cBody["id"].(string), map[string]interface{}{
			"status":             2,
			"collected_quantity": 50.0,
			"collected_unit":     "KG",
		}, token)

		invResp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": g2ID,
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
		}, token)
		var invBody map[string]interface{}
		decodeJSON(t, invResp, &invBody)
		inv2ID := invBody["id"].(string)

		// Try to mark draft as paid (should fail)
		resp := doRequest(t, "PATCH", "/v1/financial/invoices/"+inv2ID+"/paid", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("no collects returns error", func(t *testing.T) {
		g3Resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{"name": "Gen3 Empty"}, token)
		var g3 map[string]interface{}
		decodeJSON(t, g3Resp, &g3)

		resp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": g3["id"].(string),
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

// ---- Pricing Rule Specificity ----

func TestPricingRules_Specificity(t *testing.T) {
	token, _ := setupTenant(t, "SpecCorp")

	// Rule 1: catch-all (score 0) → R$5
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 5.00,
		"unit":          "KG",
	}, token)

	// Rule 2: only material (score 1) → R$10
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"material_id":    3,
		"price_per_unit": 10.00,
		"unit":          "KG",
	}, token)

	// Rule 3: material + collect_type (score 2) → R$18
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"collect_type":   "special",
		"material_id":    3,
		"price_per_unit": 18.00,
		"unit":          "KG",
	}, token)

	// Rule 4: all three factors (score 3) → R$25
	doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"collect_type":   "special",
		"material_id":    3,
		"packaging_id":   2,
		"price_per_unit": 25.00,
		"unit":          "KG",
	}, token)

	genResp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{"name": "Gen Spec"}, token)
	var gen map[string]interface{}
	decodeJSON(t, genResp, &gen)
	genID := gen["id"].(string)

	recResp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "Rec Spec"}, token)
	var rec map[string]interface{}
	decodeJSON(t, recResp, &rec)
	recID := rec["id"].(string)

	createCollect := func(t *testing.T, collectType string, materialID, packagingID interface{}, date string) string {
		t.Helper()
		body := map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"collect_type": collectType,
			"planned_date": date,
		}
		if materialID != nil {
			body["material_id"] = materialID
		}
		if packagingID != nil {
			body["packaging_id"] = packagingID
		}
		resp := doRequest(t, "POST", "/v1/collects", body, token)
		var b map[string]interface{}
		decodeJSON(t, resp, &b)
		cID := b["id"].(string)
		doRequest(t, "PATCH", "/v1/collects/"+cID, map[string]interface{}{
			"status":             2,
			"collected_quantity": 10.0,
			"collected_unit":     "KG",
		}, token)
		return cID
	}

	t.Run("most specific rule wins - all three factors → R$25", func(t *testing.T) {
		createCollect(t, "special", 3, 2, "2026-06-01")

		invResp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": genID,
			"period_start": "2026-06-01",
			"period_end":   "2026-06-01",
		}, token)
		mustStatus(t, invResp, http.StatusCreated)
		var inv map[string]interface{}
		decodeJSON(t, invResp, &inv)
		// 10 KG × R$25 = R$250
		if inv["total_amount"].(float64) != 250.00 {
			t.Fatalf("expected 250.00 (rule with 3 factors), got %v", inv["total_amount"])
		}
	})

	t.Run("two-factor rule wins when third missing - R$18", func(t *testing.T) {
		// no packaging_id → score 2 rule (collect_type + material) should win
		createCollect(t, "special", 3, nil, "2026-06-02")

		invResp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": genID,
			"period_start": "2026-06-02",
			"period_end":   "2026-06-02",
		}, token)
		mustStatus(t, invResp, http.StatusCreated)
		var inv map[string]interface{}
		decodeJSON(t, invResp, &inv)
		// 10 KG × R$18 = R$180
		if inv["total_amount"].(float64) != 180.00 {
			t.Fatalf("expected 180.00 (rule with 2 factors), got %v", inv["total_amount"])
		}
	})

	t.Run("single-factor rule wins - R$10", func(t *testing.T) {
		// normal collect with material 3 → score 1 (material only)
		createCollect(t, "normal", 3, nil, "2026-06-03")

		invResp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": genID,
			"period_start": "2026-06-03",
			"period_end":   "2026-06-03",
		}, token)
		mustStatus(t, invResp, http.StatusCreated)
		var inv map[string]interface{}
		decodeJSON(t, invResp, &inv)
		// 10 KG × R$10 = R$100
		if inv["total_amount"].(float64) != 100.00 {
			t.Fatalf("expected 100.00 (rule with material only), got %v", inv["total_amount"])
		}
	})

	t.Run("catch-all rule - R$5", func(t *testing.T) {
		// normal collect with material 5 (no specific rule) → catch-all
		createCollect(t, "normal", 5, nil, "2026-06-04")

		invResp := doRequest(t, "POST", "/v1/financial/invoices/generate", map[string]interface{}{
			"generator_id": genID,
			"period_start": "2026-06-04",
			"period_end":   "2026-06-04",
		}, token)
		mustStatus(t, invResp, http.StatusCreated)
		var inv map[string]interface{}
		decodeJSON(t, invResp, &inv)
		// 10 KG × R$5 = R$50
		if inv["total_amount"].(float64) != 50.00 {
			t.Fatalf("expected 50.00 (catch-all rule), got %v", inv["total_amount"])
		}
	})
}

// ---- Truck Costs ----

func TestTruckCosts_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "TruckCostCorp")
	var costID string

	// Create a truck first
	trkResp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
		"plate": "DEF-5678", "model": "Volvo FH", "year": 2021,
	}, token)
	var trk map[string]interface{}
	decodeJSON(t, trkResp, &trk)
	truckID := trk["id"].(string)

	t.Run("create fuel cost", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/truck-costs", map[string]interface{}{
			"truck_id":     truckID,
			"type":         "fuel",
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
			"total_amount": 4500.00,
			"total_km":     3000.0,
			"notes":        "Abastecimento maio",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		costID = body["id"].(string)

		// cost_per_km = 4500 / 3000 = 1.5
		if body["cost_per_km"].(float64) != 1.5 {
			t.Fatalf("expected cost_per_km 1.5, got %v", body["cost_per_km"])
		}
	})

	t.Run("create maintenance cost without KM", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/truck-costs", map[string]interface{}{
			"truck_id":     truckID,
			"type":         "maintenance",
			"period_start": "2026-05-15",
			"period_end":   "2026-05-15",
			"total_amount": 800.00,
			"notes":        "Troca de óleo",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		// cost_per_km = 0 when total_km = 0
		if body["cost_per_km"].(float64) != 0 {
			t.Fatalf("expected cost_per_km 0 when no km, got %v", body["cost_per_km"])
		}
	})

	t.Run("get truck cost with truck preloaded", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/truck-costs/"+costID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["truck"] == nil {
			t.Fatal("expected truck to be preloaded")
		}
	})

	t.Run("list truck costs filtered by truck", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/truck-costs?truck_id="+truckID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["total"].(float64) < 2 {
			t.Fatalf("expected at least 2 costs, got %v", body["total"])
		}
	})

	t.Run("update truck cost recalculates cost_per_km", func(t *testing.T) {
		newKM := 4000.0
		resp := doRequest(t, "PUT", "/v1/financial/truck-costs/"+costID, map[string]interface{}{
			"total_km": newKM,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		// cost_per_km = 4500 / 4000 = 1.125
		if body["cost_per_km"].(float64) != 1.125 {
			t.Fatalf("expected 1.125 after update, got %v", body["cost_per_km"])
		}
	})

	t.Run("delete truck cost", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/financial/truck-costs/"+costID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})

	t.Run("invalid type rejected", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/truck-costs", map[string]interface{}{
			"truck_id":     truckID,
			"type":         "invalid",
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
			"total_amount": 100.00,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

// ---- Personnel Costs ----

func TestPersonnelCosts_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "PersonnelCorp")
	var costID string

	// Create a driver
	drvResp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "Motorista Salário",
	}, token)
	var drv map[string]interface{}
	decodeJSON(t, drvResp, &drv)
	driverID := drv["id"].(string)

	t.Run("create personnel cost - driver", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/personnel-costs", map[string]interface{}{
			"driver_id":    driverID,
			"role":         "driver",
			"period_month": "2026-05",
			"base_salary":  3800.00,
			"benefits":     950.00,
			"notes":        "Maio 2026",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		costID = body["id"].(string)

		// total_cost = base_salary + benefits = 4750
		if body["total_cost"].(float64) != 4750.00 {
			t.Fatalf("expected total_cost 4750.00, got %v", body["total_cost"])
		}
	})

	t.Run("create personnel cost - collector", func(t *testing.T) {
		drvResp2 := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
			"name": "Coletor João",
		}, token)
		var drv2 map[string]interface{}
		decodeJSON(t, drvResp2, &drv2)

		resp := doRequest(t, "POST", "/v1/financial/personnel-costs", map[string]interface{}{
			"driver_id":    drv2["id"].(string),
			"role":         "collector",
			"period_month": "2026-05",
			"base_salary":  2200.00,
			"benefits":     550.00,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["role"].(string) != "collector" {
			t.Fatalf("expected role collector, got %s", body["role"])
		}
		if body["total_cost"].(float64) != 2750.00 {
			t.Fatalf("expected total_cost 2750.00, got %v", body["total_cost"])
		}
	})

	t.Run("get personnel cost with driver preloaded", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/personnel-costs/"+costID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["driver"] == nil {
			t.Fatal("expected driver to be preloaded")
		}
	})

	t.Run("list personnel costs filtered by driver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/personnel-costs?driver_id="+driverID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one cost for driver")
		}
	})

	t.Run("list personnel costs filtered by month", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/personnel-costs?month=2026-05", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["total"].(float64) < 2 {
			t.Fatalf("expected at least 2 costs for month, got %v", body["total"])
		}
	})

	t.Run("update recalculates total", func(t *testing.T) {
		newBenefits := 1200.00
		resp := doRequest(t, "PUT", "/v1/financial/personnel-costs/"+costID, map[string]interface{}{
			"benefits": newBenefits,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		// total = 3800 + 1200 = 5000
		if body["total_cost"].(float64) != 5000.00 {
			t.Fatalf("expected total_cost 5000.00, got %v", body["total_cost"])
		}
	})

	t.Run("delete personnel cost", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/financial/personnel-costs/"+costID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})

	t.Run("invalid role rejected", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/financial/personnel-costs", map[string]interface{}{
			"driver_id":    driverID,
			"role":         "manager",
			"period_month": "2026-05",
			"base_salary":  5000.00,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}
