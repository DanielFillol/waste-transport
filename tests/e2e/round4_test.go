package e2e

import (
	"net/http"
	"testing"
	"time"
)

// ---- include_deleted across all list endpoints ----

func TestIncludeDeleted_Generator(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("IncDelGen"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "ToDeleteGen", "cnpj": "11.222.333/0001-44",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)
	genID := gen["id"].(string)

	// Delete it
	resp = doRequest(t, "DELETE", "/v1/generators/"+genID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("default list excludes deleted", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == genID {
				t.Fatal("deleted generator should not appear in default list")
			}
		}
	})

	t.Run("include_deleted=true shows deleted with deleted_at set", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators?include_deleted=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			m := item.(map[string]interface{})
			if m["id"].(string) == genID {
				found = true
				if m["deleted_at"] == nil {
					t.Fatal("expected deleted_at to be set")
				}
			}
		}
		if !found {
			t.Fatal("deleted generator not found with include_deleted=true")
		}
	})
}

func TestIncludeDeleted_Receiver(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("IncDelRec"))

	resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
		"name": "ToDeleteRec",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)
	recID := rec["id"].(string)

	resp = doRequest(t, "DELETE", "/v1/receivers/"+recID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("include_deleted=true shows deleted receiver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/receivers?include_deleted=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == recID {
				found = true
			}
		}
		if !found {
			t.Fatal("deleted receiver not found with include_deleted=true")
		}
	})
}

func TestIncludeDeleted_Driver(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("IncDelDrv"))

	resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "ToDeleteDriver",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var drv map[string]interface{}
	decodeJSON(t, resp, &drv)
	drvID := drv["id"].(string)

	resp = doRequest(t, "DELETE", "/v1/drivers/"+drvID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("include_deleted=true shows deleted driver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/drivers?include_deleted=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == drvID {
				found = true
			}
		}
		if !found {
			t.Fatal("deleted driver not found with include_deleted=true")
		}
	})
}

func TestIncludeDeleted_Truck(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("IncDelTrk"))

	resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
		"plate": "ZZZ-9999", "model": "DelTruck", "year": 2020,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var trk map[string]interface{}
	decodeJSON(t, resp, &trk)
	trkID := trk["id"].(string)

	resp = doRequest(t, "DELETE", "/v1/trucks/"+trkID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("include_deleted=true shows deleted truck", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/trucks?include_deleted=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == trkID {
				found = true
			}
		}
		if !found {
			t.Fatal("deleted truck not found with include_deleted=true")
		}
	})
}

func TestIncludeDeleted_Route(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("IncDelRte"))

	resp := doRequest(t, "POST", "/v1/routes", map[string]interface{}{
		"name": "ToDeleteRoute", "week_day": 1, "week_number": 1,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rte map[string]interface{}
	decodeJSON(t, resp, &rte)
	rteID := rte["id"].(string)

	resp = doRequest(t, "DELETE", "/v1/routes/"+rteID, nil, token)
	mustStatus(t, resp, http.StatusNoContent)

	t.Run("include_deleted=true shows deleted route", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes?include_deleted=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == rteID {
				found = true
			}
		}
		if !found {
			t.Fatal("deleted route not found with include_deleted=true")
		}
	})
}

// ---- Driver: Active flag + ?active= filter ----

func TestDriver_ActiveFlag(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("DrvActive"))

	resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "ActiveDriver",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var drv map[string]interface{}
	decodeJSON(t, resp, &drv)
	drvID := drv["id"].(string)

	if drv["active"] != true {
		t.Fatal("new driver should be active by default")
	}

	t.Run("deactivate driver", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/drivers/"+drvID, map[string]interface{}{
			"active": false,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var updated map[string]interface{}
		decodeJSON(t, resp, &updated)
		if updated["active"] != false {
			t.Fatal("driver should be inactive after update")
		}
	})

	t.Run("?active=false returns inactive driver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/drivers?active=false", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == drvID {
				found = true
			}
		}
		if !found {
			t.Fatal("inactive driver not found with ?active=false")
		}
	})

	t.Run("?active=true does not return inactive driver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/drivers?active=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == drvID {
				t.Fatal("inactive driver should not appear with ?active=true")
			}
		}
	})
}

// ---- Truck: ?search= filter ----

func TestTruck_SearchFilter(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("TrkSearch"))

	resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
		"plate": "ABC-1234", "model": "UniqueSearchModel", "year": 2022,
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var trk map[string]interface{}
	decodeJSON(t, resp, &trk)
	trkID := trk["id"].(string)

	t.Run("search by plate finds truck", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/trucks?search=ABC-1234", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == trkID {
				found = true
			}
		}
		if !found {
			t.Fatal("truck not found by plate search")
		}
	})

	t.Run("search by model finds truck", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/trucks?search=UniqueSearchModel", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == trkID {
				found = true
			}
		}
		if !found {
			t.Fatal("truck not found by model search")
		}
	})

	t.Run("search with no match returns empty", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/trucks?search=NOMATCH99999", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if len(result["data"].([]interface{})) != 0 {
			t.Fatal("expected empty results for non-matching search")
		}
	})
}

// ---- Invoice: ?status= filter + paid_at ----

func TestInvoice_StatusFilterAndPaidAt(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("InvStatus"))

	resp := doRequest(t, "POST", "/v1/financial/pricing-rules", map[string]interface{}{
		"price_per_unit": 5.0, "unit": "KG",
	}, token)
	mustStatus(t, resp, http.StatusCreated)

	resp = doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "InvStatusGen", "cnpj": "55.666.777/0001-88",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "InvStatusRec"}, token)
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
		"status": 2, "collected_quantity": 100.0, "collected_unit": "KG",
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

	t.Run("?status=draft returns draft invoice", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices?status=draft", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == invID {
				found = true
				if item.(map[string]interface{})["status"] != "draft" {
					t.Fatal("expected draft status")
				}
			}
		}
		if !found {
			t.Fatal("draft invoice not found with ?status=draft")
		}
	})

	t.Run("?status=issued returns empty before issuing", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices?status=issued", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == invID {
				t.Fatal("invoice should not appear as issued before being issued")
			}
		}
	})

	// Issue the invoice
	resp = doRequest(t, "PATCH", "/v1/financial/invoices/"+invID+"/issue", nil, token)
	mustStatus(t, resp, http.StatusOK)

	t.Run("?status=issued returns issued invoice", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices?status=issued", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == invID {
				found = true
			}
		}
		if !found {
			t.Fatal("issued invoice not found with ?status=issued")
		}
	})

	// Mark as paid
	resp = doRequest(t, "PATCH", "/v1/financial/invoices/"+invID+"/paid", nil, token)
	mustStatus(t, resp, http.StatusOK)
	var paidInv map[string]interface{}
	decodeJSON(t, resp, &paidInv)

	t.Run("paid_at is set when marking invoice as paid", func(t *testing.T) {
		if paidInv["paid_at"] == nil {
			t.Fatal("paid_at should be set when invoice is marked as paid")
		}
		if paidInv["status"] != "paid" {
			t.Fatal("expected status to be paid")
		}
	})

	t.Run("?status=paid returns paid invoice", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/financial/invoices?status=paid", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		found := false
		for _, item := range result["data"].([]interface{}) {
			if item.(map[string]interface{})["id"].(string) == invID {
				found = true
			}
		}
		if !found {
			t.Fatal("paid invoice not found with ?status=paid")
		}
	})
}

// ---- Collect: state machine enforcement ----

func TestCollect_StateMachine(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ColSM"))

	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "SMGen", "cnpj": "99.888.777/0001-66",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "SMRec"}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)

	createCollect := func(t *testing.T) string {
		t.Helper()
		resp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": gen["id"],
			"receiver_id":  rec["id"],
			"planned_date": time.Now().Format("2006-01-02"),
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var col map[string]interface{}
		decodeJSON(t, resp, &col)
		return col["id"].(string)
	}

	t.Run("cannot change status of collected collect", func(t *testing.T) {
		colID := createCollect(t)
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 2, "collected_quantity": 10.0, "collected_unit": "KG",
		}, token)
		mustStatus(t, resp, http.StatusOK)

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 3,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("cannot change status of cancelled collect", func(t *testing.T) {
		colID := createCollect(t)
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 3,
		}, token)
		mustStatus(t, resp, http.StatusOK)

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 1,
		}, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("can update non-status fields on collected collect", func(t *testing.T) {
		colID := createCollect(t)
		resp := doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"status": 2, "collected_quantity": 10.0, "collected_unit": "KG",
		}, token)
		mustStatus(t, resp, http.StatusOK)

		resp = doRequest(t, "PATCH", "/v1/collects/"+colID, map[string]interface{}{
			"notes": "updated note after collection",
		}, token)
		mustStatus(t, resp, http.StatusOK)
	})
}
