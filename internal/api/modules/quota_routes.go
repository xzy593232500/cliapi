package modules

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/quota"
)

func RegisterQuotaRoutes(engine *gin.Engine, authMiddleware gin.HandlerFunc) {
	if engine == nil || authMiddleware == nil {
		return
	}
	group := engine.Group("/v1/quota")
	group.Use(authMiddleware)
	{
		group.GET("/balance", func(c *gin.Context) {
			store := quota.DefaultStore()
			if store == nil {
				c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
				return
			}
			apiKey := strings.TrimSpace(readAPIKey(c))
			if apiKey == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_api_key"})
				return
			}
			account, err := store.GetAccount(c.Request.Context(), apiKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, account)
		})
		group.POST("/redeem", func(c *gin.Context) {
			store := quota.DefaultStore()
			if store == nil {
				c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
				return
			}
			apiKey := strings.TrimSpace(readAPIKey(c))
			if apiKey == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_api_key"})
				return
			}
			var body struct { Code string `json:"code"` }
			if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Code) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
				return
			}
			account, err := store.RedeemCode(c.Request.Context(), apiKey, body.Code)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, account)
		})
	}
}

func readAPIKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get("apiKey"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
