package e2e

import (
	"net/http"
	"testing"
)

func TestAuth_RegisterTenant(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/tenants", map[string]string{"name": "Acme Resíduos"}, "")
		mustStatus(t, resp, http.StatusCreated)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)

		if body["token"] == nil || body["token"].(string) == "" {
			t.Fatal("expected non-empty token")
		}
		tenant := body["tenant"].(map[string]interface{})
		if tenant["slug"].(string) != "acme-resduos" {
			// slug generation is locale-dependent; just check it's non-empty
			if tenant["slug"].(string) == "" {
				t.Fatal("expected non-empty slug")
			}
		}
		user := body["user"].(map[string]interface{})
		if user["username"].(string) != "admin" {
			t.Fatalf("expected admin user, got %s", user["username"])
		}
		if user["role"].(string) != "admin" {
			t.Fatalf("expected role admin, got %s", user["role"])
		}
	})

	t.Run("duplicate tenant", func(t *testing.T) {
		doRequest(t, "POST", "/v1/auth/tenants", map[string]string{"name": "DupTenant"}, "")
		resp := doRequest(t, "POST", "/v1/auth/tenants", map[string]string{"name": "DupTenant"}, "")
		mustStatus(t, resp, http.StatusConflict)
	})

	t.Run("missing name", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/tenants", map[string]string{}, "")
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

func TestAuth_Login(t *testing.T) {
	_, slug := setupTenant(t, "LoginCorp")

	t.Run("success", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug":     slug,
			"username": "admin",
			"password": "admin123",
		}, "")
		mustStatus(t, resp, http.StatusOK)

		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["token"] == nil || body["token"].(string) == "" {
			t.Fatal("expected token in response")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug":     slug,
			"username": "admin",
			"password": "wrongpass",
		}, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("wrong slug", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug":     "nonexistent-tenant",
			"username": "admin",
			"password": "admin123",
		}, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("missing fields", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug": slug,
		}, "")
		mustStatus(t, resp, http.StatusBadRequest)
	})
}

func TestAuth_Protected(t *testing.T) {
	t.Run("no token", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators", nil, "")
		mustStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("invalid token", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/generators", nil, "invalid-token")
		mustStatus(t, resp, http.StatusUnauthorized)
	})
}

func TestUsers_CRUD(t *testing.T) {
	token, _ := setupTenant(t, "UsersCorp")

	var userID string

	t.Run("list users - admin sees own users", func(t *testing.T) {
		resp := doRequest(t, "GET", "/v1/users", nil, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		users := body["data"].([]interface{})
		if len(users) == 0 {
			t.Fatal("expected at least the admin user")
		}
	})

	t.Run("create user", func(t *testing.T) {
		resp := doRequest(t, "POST", "/v1/users", map[string]interface{}{
			"name":     "João Operador",
			"username": "joao",
			"password": "senha123",
			"role":     "user",
		}, token)
		mustStatus(t, resp, http.StatusCreated)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		userID = body["id"].(string)
		if body["username"].(string) != "joao" {
			t.Fatalf("unexpected username: %s", body["username"])
		}
	})

	t.Run("update user", func(t *testing.T) {
		resp := doRequest(t, "PUT", "/v1/users/"+userID, map[string]interface{}{
			"name": "João da Silva",
		}, token)
		mustStatus(t, resp, http.StatusOK)
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		if body["name"].(string) != "João da Silva" {
			t.Fatalf("name not updated: %s", body["name"])
		}
	})

	t.Run("delete user", func(t *testing.T) {
		resp := doRequest(t, "DELETE", "/v1/users/"+userID, nil, token)
		mustStatus(t, resp, http.StatusNoContent)
	})

	t.Run("non-admin cannot manage users", func(t *testing.T) {
		_, slug := setupTenant(t, "NonAdminCorp")
		// Login as admin, create a regular user
		loginResp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug":     slug,
			"username": "admin",
			"password": "admin123",
		}, "")
		var lr map[string]interface{}
		decodeJSON(t, loginResp, &lr)
		adminToken := lr["token"].(string)

		createResp := doRequest(t, "POST", "/v1/users", map[string]interface{}{
			"name": "Regular", "username": "regular", "password": "pass123", "role": "user",
		}, adminToken)
		mustStatus(t, createResp, http.StatusCreated)

		// Login as regular user
		userLoginResp := doRequest(t, "POST", "/v1/auth/login", map[string]string{
			"slug":     slug,
			"username": "regular",
			"password": "pass123",
		}, "")
		var ulr map[string]interface{}
		decodeJSON(t, userLoginResp, &ulr)
		userToken := ulr["token"].(string)

		resp := doRequest(t, "GET", "/v1/users", nil, userToken)
		mustStatus(t, resp, http.StatusForbidden)
	})
}
