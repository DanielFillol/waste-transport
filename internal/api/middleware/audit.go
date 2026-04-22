package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// routeEntityTypes maps Gin full-path patterns to domain entity types.
var routeEntityTypes = map[string]string{
	"/v1/users":                                "user",
	"/v1/users/:id":                            "user",
	"/v1/generators":                           "generator",
	"/v1/generators/:id":                       "generator",
	"/v1/generators/import":                    "generator",
	"/v1/receivers":                            "receiver",
	"/v1/receivers/:id":                        "receiver",
	"/v1/receivers/import":                     "receiver",
	"/v1/drivers":                              "driver",
	"/v1/drivers/:id":                          "driver",
	"/v1/drivers/import":                       "driver",
	"/v1/trucks":                               "truck",
	"/v1/trucks/:id":                           "truck",
	"/v1/routes":                               "route",
	"/v1/routes/:id":                           "route",
	"/v1/routes/:id/generate-collects":         "collect",
	"/v1/collects":                             "collect",
	"/v1/collects/:id":                         "collect",
	"/v1/collects/import":                      "collect",
	"/v1/collects/bulk-status":                 "collect",
	"/v1/collects/bulk-cancel":                 "collect",
	"/v1/collects/bulk-assign-route":           "collect",
	"/v1/alerts/:id/read":                      "alert",
	"/v1/alerts/read-all":                      "alert",
	"/v1/financial/pricing-rules":              "pricing_rule",
	"/v1/financial/pricing-rules/:id":          "pricing_rule",
	"/v1/financial/invoices/generate":          "invoice",
	"/v1/financial/invoices/:id/issue":         "invoice",
	"/v1/financial/invoices/:id/paid":          "invoice",
	"/v1/financial/truck-costs":                "truck_cost",
	"/v1/financial/truck-costs/:id":            "truck_cost",
	"/v1/financial/personnel-costs":            "personnel_cost",
	"/v1/financial/personnel-costs/:id":        "personnel_cost",
}

type captureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *captureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Audit(repo *repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Read and restore the request body so handlers can still use it.
		rawBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		// Wrap the writer to capture the response body.
		cw := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = cw

		c.Next()

		status := cw.Status()
		if status < 200 || status >= 300 {
			return
		}

		fullPath := c.FullPath()
		entityType, ok := routeEntityTypes[fullPath]
		if !ok {
			return
		}

		entityID := c.Param("id")
		if entityID == "" {
			// POST to a collection endpoint — try to extract the created record's ID from the response.
			entityID = extractIDFromBody(cw.body.Bytes())
		}

		action := deriveAction(method, fullPath)

		tenantID := GetTenantID(c)
		actorID := GetUserID(c)

		var actorPtr *uuid.UUID
		if actorID != uuid.Nil {
			actorPtr = &actorID
		}

		log := &entity.AuditLog{
			CreatedAt:  time.Now(),
			TenantID:   tenantID,
			ActorID:    actorPtr,
			EntityType: entityType,
			EntityID:   entityID,
			Action:     action,
			Payload:    string(rawBody),
		}
		_ = repo.Create(log)
	}
}

func deriveAction(method, fullPath string) string {
	parts := strings.Split(strings.TrimSuffix(fullPath, "/"), "/")
	last := parts[len(parts)-1]

	switch {
	case method == "DELETE":
		return "delete"
	case last == "import":
		return "import"
	case last == "generate":
		return "generate"
	case last == "generate-collects":
		return "generate_collects"
	case last == "bulk-status":
		return "bulk_status"
	case last == "bulk-cancel":
		return "bulk_cancel"
	case last == "bulk-assign-route":
		return "bulk_assign_route"
	case last == "issue":
		return "issue"
	case last == "paid":
		return "mark_paid"
	case last == "read":
		return "mark_read"
	case last == "read-all":
		return "mark_all_read"
	case method == "POST":
		return "create"
	default:
		return "update"
	}
}

func extractIDFromBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}
	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}
