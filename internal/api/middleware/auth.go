package middleware

import (
	"net/http"
	"strings"

	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	CtxUserID   = "user_id"
	CtxTenantID = "tenant_id"
	CtxRole     = "role"
	CtxUsername = "username"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &jwtutil.Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxTenantID, claims.TenantID)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxUsername, claims.Username)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

func GetTenantID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxTenantID)
	id, _ := v.(uuid.UUID)
	return id
}

func GetUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CtxUserID)
	id, _ := v.(uuid.UUID)
	return id
}
