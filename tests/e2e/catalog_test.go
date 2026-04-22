package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

// ---- Generators ----

func TestGenerators_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "GenCorp")
	var genID string

	t.Run("create generator", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name":    "Fábrica ABC",
			"cnpj":    "12.345.678/0001-99",
			"address": "Rua das Flores, 100",
			"zipcode": "01310-100",
			"city_id": 1,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		genID = body["id"].(string)
		if body["name"].(string) != "Fábrica ABC" {
			t.Fatalf("unexpected name: %s", body["name"])
		}
	})

	t.Run("get generator", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators/"+genID, nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["id"].(string) != genID {
			t.Fatal("id mismatch")
		}
	})

	t.Run("list generators with pagination", func(t *testing.T) {
		// Create a few more
		for i := 0; i < 3; i++ {
			doRequest(t, "POST", "/v1/generators", map[string]interface{}{
				"name": fmt.Sprintf("Fábrica %d", i),
				"cnpj": fmt.Sprintf("00.000.000/0001-%02d", i),
			}, token)
		}
		resp := doRequest(t, "GET", "/v1/generators?page=1&limit=2", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["limit"].(float64) != 2 {
			t.Fatalf("expected limit 2, got %v", body["limit"])
		}
		if body["total"].(float64) < 2 {
			t.Fatalf("expected at least 2 total, got %v", body["total"])
		}
	})

	t.Run("list generators with search", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators?search=Fábrica+ABC", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one result")
		}
	})

	t.Run("update generator", func(t *testing.T) {
		newAddr := "Av. Paulista, 1000"
		resp := doRequest(t, "PATCH", "/v1/generators/"+genID, map[string]interface{}{
			"address": newAddr,
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["address"].(string) != newAddr {
			t.Fatalf("address not updated: %s", body["address"])
		}
	})

	t.Run("delete generator", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/generators/"+genID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
		// Confirm deleted
		resp2 := doRequest(t, "GET", "/v1/generators/"+genID, nil, token)
		mustStatus(t, resp2, http.StatusNotFound)
	})

	t.Run("create generator with coordinates", func(t *testing.T) {
		lat := -23.5505
		lon := -46.6333
		resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
			"name":      "Fábrica Localizada",
			"latitude":  lat,
			"longitude": lon,
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["latitude"].(float64) != lat {
			t.Fatalf("latitude mismatch: %v", body["latitude"])
		}
	})

	t.Run("tenant isolation - cannot see other tenant generators", func(t *testing.T) {
		otherToken, _ := setupTenant(t, "OtherGenCorp")
		resp := doRequest(t, "GET", "/v1/generators", nil, otherToken)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) != 0 {
			t.Fatalf("expected 0 generators for new tenant, got %d", len(data))
		}
	})
}

// ---- Receivers ----

func TestReceivers_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "RecCorp")
	var recID string

	t.Run("create receiver", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
			"name":           "Aterro XYZ",
			"cnpj":           "98.765.432/0001-10",
			"address":        "Rodovia SP-100, Km 50",
			"license_number": "LIC-2024-001",
			"license_expiry": "2026-12-31",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		recID = body["id"].(string)
		if body["license_number"].(string) != "LIC-2024-001" {
			t.Fatalf("license_number mismatch")
		}
	})

	t.Run("get receiver", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/receivers/"+recID, nil, token)
		mustStatus(t, resp, http.StatusOK)
	})

	t.Run("list receivers", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/receivers?page=1&limit=10", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected at least one receiver")
		}
	})

	t.Run("update receiver license", func(t *testing.T) {
		resp := doRequest(t, "PATCH", "/v1/receivers/"+recID, map[string]interface{}{
			"license_expiry": "2027-06-30",
		}, token)
		mustStatus(t, resp, http.StatusOK)
	})

	t.Run("delete receiver", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/receivers/"+recID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
		resp2 := doRequest(t, "GET", "/v1/receivers/"+recID, nil, token)
		mustStatus(t, resp2, http.StatusNotFound)
	})
}

// ---- Domain ----

func TestDomain_Lists(t *testing.T) {
	token, _ := setupTenant(t, "DomainCorp")

	t.Run("list materials", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/domain/materials", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected seeded materials")
		}
	})

	t.Run("list packagings", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/domain/packagings", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected seeded packagings")
		}
	})

	t.Run("list treatments", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/domain/treatments", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected seeded treatments")
		}
	})

	t.Run("list UFs", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/domain/ufs", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected seeded UFs")
		}
	})

	t.Run("list cities filtered by UF", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/domain/cities?uf_id=1", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		data := body["data"].([]interface{})
		if len(data) == 0 {
			t.Fatal("expected cities for UF 1")
		}
	})
}
