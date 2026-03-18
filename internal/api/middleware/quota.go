package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/quota"
)

const defaultPrecheckTokens int64 = 1

func QuotaBalanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		store := quota.DefaultStore()
		if store == nil {
			c.Next()
			return
		}
		value, exists := c.Get("apiKey")
		if !exists {
			c.Next()
			return
		}
		apiKey := strings.TrimSpace(toString(value))
		if apiKey == "" {
			c.Next()
			return
		}
		required := estimateRequiredTokens(c)
		ok, account, err := store.HasEnoughBalance(c.Request.Context(), apiKey, required)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "quota_check_failed"})
			return
		}
		if !ok {
			balance := int64(0)
			if account != nil {
				balance = account.TokenBalance
			}
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error": "insufficient_token_balance",
				"required_tokens_estimate": required,
				"balance_tokens": balance,
			})
			return
		}
		c.Next()
	}
}

func estimateRequiredTokens(c *gin.Context) int64 {
	if c == nil {
		return defaultPrecheckTokens
	}
	if v := strings.TrimSpace(c.GetHeader("X-Required-Tokens")); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultPrecheckTokens
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
