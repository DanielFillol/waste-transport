package e2e

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"
)

func doMultipartRequest(t *testing.T, path string, csvContent string, token string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "import.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, err = fw.Write([]byte(csvContent))
	if err != nil {
		t.Fatalf("write csv: %v", err)
	}
	w.Close()

	req, err := http.NewRequest("POST", url(path), &buf)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func TestImport_Generators(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ImportGenTenant"))

	t.Run("happy path - all rows valid", func(t *testing.T) {
		csv := "name,cnpj,address,external_id\nGerador CSV A,11.222.333/0001-44,Rua A 100,EXT-001\nGerador CSV B,22.333.444/0001-55,Rua B 200,EXT-002\n"
		resp := doMultipartRequest(t, "/v1/generators/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["total"].(float64) != 2 {
			t.Errorf("expected total=2, got %v", result["total"])
		}
		if result["created"].(float64) != 2 {
			t.Errorf("expected created=2, got %v", result["created"])
		}
		errs := result["errors"].([]interface{})
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %v", errs)
		}
	})

	t.Run("row missing required name is reported as error", func(t *testing.T) {
		csv := "name,cnpj\nGerador Valido,11.111.111/0001-11\n,22.222.222/0001-22\n"
		resp := doMultipartRequest(t, "/v1/generators/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 1 {
			t.Errorf("expected created=1, got %v", result["created"])
		}
		errs := result["errors"].([]interface{})
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
		errRow := errs[0].(map[string]interface{})
		if errRow["row"].(float64) != 3 {
			t.Errorf("expected error on row 3, got %v", errRow["row"])
		}
	})

	t.Run("missing file returns 400", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/generators/import", nil, token)
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

func TestImport_Drivers(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ImportDrvTenant"))

	t.Run("import drivers with valid CNH expiry", func(t *testing.T) {
		csv := "name,cnh_number,cnh_category,cnh_expiry_date\nMotorista CSV,12345678,B,2027-12-31\n"
		resp := doMultipartRequest(t, "/v1/drivers/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 1 {
			t.Errorf("expected created=1, got %v", result["created"])
		}
	})

	t.Run("invalid date in row produces error for that row", func(t *testing.T) {
		csv := "name,cnh_expiry_date\nBom Motorista,2027-06-01\nMau Motorista,31/12/2027\n"
		resp := doMultipartRequest(t, "/v1/drivers/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 1 {
			t.Errorf("expected 1 created, got %v", result["created"])
		}
		errs := result["errors"].([]interface{})
		if len(errs) != 1 {
			t.Errorf("expected 1 error for invalid date row, got %d", len(errs))
		}
	})
}

func TestImport_Receivers(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ImportRecTenant"))

	t.Run("import receivers successfully", func(t *testing.T) {
		csv := "name,license_number\nReceptor CSV 1,LIC-001\nReceptor CSV 2,LIC-002\n"
		resp := doMultipartRequest(t, "/v1/receivers/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 2 {
			t.Errorf("expected created=2, got %v", result["created"])
		}
	})
}

func TestImport_Collects(t *testing.T) {
	token, _ := setupTenant(t, uniqueName("ImportColTenant"))

	// Seed generator and receiver
	resp := doRequest(t, "POST", "/v1/generators", map[string]interface{}{
		"name": "Gen Import",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var gen map[string]interface{}
	decodeJSON(t, resp, &gen)
	genID := gen["id"].(string)

	resp = doRequest(t, "POST", "/v1/receivers", map[string]interface{}{
		"name": "Rec Import",
	}, token)
	mustStatus(t, resp, http.StatusCreated)
	var rec map[string]interface{}
	decodeJSON(t, resp, &rec)
	recID := rec["id"].(string)

	t.Run("import collects - happy path", func(t *testing.T) {
		csv := fmt.Sprintf("generator_id,receiver_id,planned_date\n%s,%s,2025-08-01\n%s,%s,2025-08-15\n",
			genID, recID, genID, recID)
		resp := doMultipartRequest(t, "/v1/collects/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 2 {
			t.Errorf("expected created=2, got %v", result["created"])
		}
	})

	t.Run("row missing required fields is reported as error", func(t *testing.T) {
		csv := fmt.Sprintf("generator_id,receiver_id,planned_date\n%s,%s,2025-09-01\n,,\n",
			genID, recID)
		resp := doMultipartRequest(t, "/v1/collects/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		if result["created"].(float64) != 1 {
			t.Errorf("expected created=1, got %v", result["created"])
		}
		errs := result["errors"].([]interface{})
		if len(errs) != 1 {
			t.Errorf("expected 1 error, got %d", len(errs))
		}
	})

	t.Run("invalid planned_date format produces error", func(t *testing.T) {
		csv := fmt.Sprintf("generator_id,receiver_id,planned_date\n%s,%s,01/08/2025\n",
			genID, recID)
		resp := doMultipartRequest(t, "/v1/collects/import", csv, token)
		mustStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		decodeJSON(t, resp, &result)
		errs := result["errors"].([]interface{})
		if len(errs) != 1 {
			t.Errorf("expected 1 error for bad date, got %d", len(errs))
		}
	})
}
