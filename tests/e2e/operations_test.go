package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

// ---- Drivers ----

func TestDrivers_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "DriverCorp")
	var driverID string

	t.Run("create driver", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
			"name":         "Carlos Motorista",
			"email":        "carlos@example.com",
			"phone":        "11999991111",
			"cpf":          "123.456.789-00",
			"cnh_number":   "00123456789",
			"cnh_category": "D",
			"cnh_expiry_date": "2027-03-15",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		driverID = body["id"].(string)
		if body["cnh_category"].(string) != "D" {
			t.Fatalf("unexpected cnh_category: %s", body["cnh_category"])
		}
	})

	t.Run("get driver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/drivers/"+driverID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["email"].(string) != "carlos@example.com" {
			t.Fatalf("email mismatch")
		}
	})

	t.Run("list drivers with search", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
				"name": fmt.Sprintf("Motorista %d", i),
				"cpf":  fmt.Sprintf("000.000.000-%02d", i),
			}, token)
		}
		resp := doRequest(t, "GET", "/v1/drivers?search=Carlos", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least Carlos")
		}
	})

	t.Run("update driver phone", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/drivers/"+driverID, map[string]interface{}{
			"phone": "11988887777",
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["phone"].(string) != "11988887777" {
			t.Fatalf("phone not updated: %s", body["phone"])
		}
	})

	t.Run("delete driver", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/drivers/"+driverID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})
}

// ---- Trucks ----

func TestTrucks_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "TruckCorp")
	var truckID string

	t.Run("create truck", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
			"plate":       "ABC-1234",
			"model":       "Mercedes Actros",
			"year":        2022,
			"capacity_kg": 15000.0,
			"capacity_m3": 30.0,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		truckID = body["id"].(string)
		if body["plate"].(string) != "ABC-1234" {
			t.Fatalf("plate mismatch")
		}
		if body["active"].(bool) != true {
			t.Fatal("expected truck to be active by default")
		}
	})

	t.Run("get truck", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/trucks/"+truckID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["capacity_kg"].(float64) != 15000.0 {
			t.Fatalf("capacity_kg mismatch")
		}
	})

	t.Run("list trucks - only active", func(t *testing.T) {
		// Create an inactive truck
		inactiveResp := doRequest(t, "POST", "/v1/trucks", map[string]interface{}{
			"plate": "ZZZ-9999", "model": "Old Truck", "year": 2005,
		}, token)
		var inactive map[string]interface{}
		decodeJSON(t, inactiveResp, &inactive)
		inactiveID := inactive["id"].(string)
		falseVal := false
		doRequest(t, "PATCH", "/v1/trucks/"+inactiveID, map[string]interface{}{
			"active": &falseVal,
		}, token)

		resp := doRequest(t, "GET", "/v1/trucks?active=true", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		for _, item := range data {
			truck := item.(map[string]interface{})
			if truck["active"].(bool) == false {
				t.Fatal("inactive truck in active-only list")
			}
		}
	})

	t.Run("update truck", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/trucks/"+truckID, map[string]interface{}{
			"capacity_kg": 18000.0,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["capacity_kg"].(float64) != 18000.0 {
			t.Fatalf("capacity_kg not updated")
		}
	})

	t.Run("deactivate truck", func(t *testing.T) {
		falseVal := false
		resp := doRequest(t, "PATCH", "/v1/trucks/"+truckID, map[string]interface{}{
			"active": &falseVal,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["active"].(bool) != false {
			t.Fatal("expected truck to be inactive")
		}
	})

	t.Run("delete truck", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/trucks/"+truckID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})
}

// ---- Routes ----

func TestRoutes_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "RouteCorp")

	// Create two drivers to associate
	d1Resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "Motorista Rota 1",
	}, token)
	var d1 map[string]interface{}
	decodeJSON(t, d1Resp, &d1)
	driver1ID := d1["id"].(string)

	d2Resp := doRequest(t, "POST", "/v1/drivers", map[string]interface{}{
		"name": "Motorista Rota 2",
	}, token)
	var d2 map[string]interface{}
	decodeJSON(t, d2Resp, &d2)
	driver2ID := d2["id"].(string)

	var routeID string

	t.Run("create route with drivers", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/routes", map[string]interface{}{
			"name":         "Rota Norte",
			"material_id":  1,
			"packaging_id": 1,
			"treatment_id": 1,
			"week_day":     2,
			"week_number":  1,
			"driver_ids":   []string{driver1ID, driver2ID},
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		routeID = body["id"].(string)
		drivers := body["drivers"].([]interface{})
		if len(drivers) != 2 {
			t.Fatalf("expected 2 drivers, got %d", len(drivers))
		}
	})

	t.Run("get route with associations", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes/"+routeID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["name"].(string) != "Rota Norte" {
			t.Fatalf("name mismatch")
		}
		if body["material"] == nil {
			t.Fatal("expected material to be preloaded")
		}
	})

	t.Run("update route - change driver", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/routes/"+routeID, map[string]interface{}{
			"name":       "Rota Norte Atualizada",
			"driver_ids": []string{driver1ID},
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["name"].(string) != "Rota Norte Atualizada" {
			t.Fatalf("name not updated")
		}
		drivers := body["drivers"].([]interface{})
		if len(drivers) != 1 {
			t.Fatalf("expected 1 driver after update, got %d", len(drivers))
		}
	})

	t.Run("list routes paginated", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/routes?page=1&limit=10", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["total"].(float64) < 1 {
			t.Fatal("expected at least 1 route")
		}
	})

	t.Run("delete route", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/routes/"+routeID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})
}

// ---- Collects ----

func TestCollects_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "CollectCorp")

	// Setup dependencies
	genResp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{"name": "Gen Coleta"}, token)
	var gen map[string]interface{}
	decodeJSON(t, genResp, &gen)
	genID := gen["id"].(string)

	recResp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{"name": "Rec Coleta"}, token)
	var rec map[string]interface{}
	decodeJSON(t, recResp, &rec)
	recID := rec["id"].(string)

	var collectID string

	t.Run("create collect - normal", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"material_id":  1,
			"packaging_id": 2,
			"treatment_id": 1,
			"collect_type": "normal",
			"planned_date": "2026-05-01",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		collectID = body["id"].(string)
		if body["status"].(float64) != 1 {
			t.Fatal("expected status PLANNED (1)")
		}
		if body["collect_type"].(string) != "normal" {
			t.Fatalf("expected collect_type normal, got %s", body["collect_type"])
		}
	})

	t.Run("get collect with all associations", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects/"+collectID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["generator"] == nil {
			t.Fatal("expected generator to be preloaded")
		}
		if body["receiver"] == nil {
			t.Fatal("expected receiver to be preloaded")
		}
	})

	t.Run("list collects with filters", func(t *testing.T) {
		// Create a special collect
		doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"collect_type": "special",
			"planned_date": "2026-05-15",
		}, token)

		// Filter by collect_type
		resp := doRequest(t, "GET", "/v1/collects?collect_type=special", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		for _, item := range data {
			c := item.(map[string]interface{})
			if c["collect_type"].(string) != "special" {
				t.Fatal("filter by collect_type not working")
			}
		}
	})

	t.Run("list collects filtered by date range", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?date_from=2026-05-01&date_to=2026-05-10", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected collect in date range")
		}
	})

	t.Run("list collects filtered by generator", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?generator_id="+genID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected collects for generator")
		}
	})

	t.Run("update collect - mark as collected with quantity", func(t *testing.T) {
		qty := 250.5
		resp := doRequest(t, "PATCH", "/v1/collects/"+collectID, map[string]interface{}{
			"status":             2,
			"collected_quantity": qty,
			"collected_unit":     "KG",
			"collected_weight":   250.5,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["status"].(float64) != 2 {
			t.Fatal("expected status COLLECTED (2)")
		}
		if body["collected_quantity"].(float64) != qty {
			t.Fatalf("quantity mismatch: %v", body["collected_quantity"])
		}
		if body["collected_unit"].(string) != "KG" {
			t.Fatalf("unit mismatch: %s", body["collected_unit"])
		}
	})

	t.Run("cancel collect via DELETE", func(t *testing.T) {
		// Create a fresh collect to cancel
		cResp := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"planned_date": "2026-06-01",
		}, token)
		var cBody map[string]interface{}
		decodeJSON(t, cResp, &cBody)
		cID := cBody["id"].(string)

		resp := doRequest(t, "DELETE", "/v1/collects/"+cID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["status"].(float64) != 3 {
			t.Fatal("expected status CANCELLED (3)")
		}
	})

	t.Run("bulk status update", func(t *testing.T) {
		// Create 3 planned collects
		var ids []string
		for i := 0; i < 3; i++ {
			r := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
				"generator_id": genID,
				"receiver_id":  recID,
				"planned_date": fmt.Sprintf("2026-07-%02d", i+1),
			}, token)
			var b map[string]interface{}
			decodeJSON(t, r, &b)
			ids = append(ids, b["id"].(string))
		}

		resp := doRequest(t, "POST", "/v1/collects/bulk-status", map[string]interface{}{
			"ids":    ids,
			"status": 3,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if int(body["updated"].(float64)) != 3 {
			t.Fatalf("expected 3 updated, got %v", body["updated"])
		}
	})

	t.Run("bulk assign route", func(t *testing.T) {
		rteResp := doRequest(t, "POST", "/v1/routes", map[string]interface{}{
			"name":        "Rota Bulk",
			"week_day":    3,
			"week_number": 2,
		}, token)
		var rte map[string]interface{}
		decodeJSON(t, rteResp, &rte)
		routeID := rte["id"].(string)

		c1 := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"planned_date": "2026-08-01",
		}, token)
		var c1b map[string]interface{}
		decodeJSON(t, c1, &c1b)

		c2 := doRequest(t, "POST", "/v1/collects", map[string]interface{}{
			"generator_id": genID,
			"receiver_id":  recID,
			"planned_date": "2026-08-02",
		}, token)
		var c2b map[string]interface{}
		decodeJSON(t, c2, &c2b)

		resp := doRequest(t, "POST", "/v1/collects/bulk-assign-route", map[string]interface{}{
			"ids":      []string{c1b["id"].(string), c2b["id"].(string)},
			"route_id": routeID,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if int(body["updated"].(float64)) != 2 {
			t.Fatalf("expected 2 updated, got %v", body["updated"])
		}

		// Verify route assignment
		detail := doRequest(t, "GET", "/v1/collects/"+c1b["id"].(string), nil, token)
		var detailBody map[string]interface{}
		decodeJSON(t, detail, &detailBody)
		if detailBody["route_id"].(string) != routeID {
			t.Fatalf("route_id not assigned: %v", detailBody["route_id"])
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/collects?status=1", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		for _, item := range data {
			c := item.(map[string]interface{})
			if c["status"].(float64) != 1 {
				t.Fatal("filter by status not working")
			}
		}
	})
}
