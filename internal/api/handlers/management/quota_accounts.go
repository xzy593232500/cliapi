package management

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/quota"
)

func (h *Handler) quota() *quota.Store {
	if h == nil {
		return nil
	}
	if h.quotaStore != nil {
		return h.quotaStore
	}
	return quota.DefaultStore()
}

func (h *Handler) GetQuotaAccounts(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	items, err := store.ListAccounts(c.Request.Context(), 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetQuotaTransactions(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	items, err := store.ListTransactions(c.Request.Context(), c.Query("api_key"), 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetRedeemCodes(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	items, err := store.ListRedeemCodes(c.Request.Context(), 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) PostRedeemCodes(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	var body struct {
		USDAmount float64 `json:"usd_amount"`
		Count     int     `json:"count"`
		ExpiresAt string  `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.USDAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 500 {
		body.Count = 500
	}
	var expiresAt *time.Time
	if strings.TrimSpace(body.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, body.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_expires_at"})
			return
		}
		expiresAt = &parsed
	}
	codes := make([]quota.RedeemCode, 0, body.Count)
	plainCodes := make([]string, 0, body.Count)
	for i := 0; i < body.Count; i++ {
		code := randomCode()
		plainCodes = append(plainCodes, code)
		codes = append(codes, quota.RedeemCode{Code: code, USDAmount: body.USDAmount, TokenAmount: store.USDToTokens(body.USDAmount), ExpiresAt: expiresAt})
	}
	if err := store.CreateRedeemCodes(c.Request.Context(), codes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codes": plainCodes, "usd_amount": body.USDAmount, "token_amount": store.USDToTokens(body.USDAmount)})
}

func (h *Handler) PostQuotaAdjust(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	var body struct {
		APIKey      string  `json:"api_key"`
		TokenAmount *int64  `json:"token_amount"`
		USDAmount   *float64 `json:"usd_amount"`
		Note        string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.APIKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	var tokenDelta int64
	var usdDelta float64
	if body.TokenAmount != nil {
		tokenDelta = *body.TokenAmount
		usdDelta = store.TokensToUSD(tokenDelta)
	} else if body.USDAmount != nil {
		usdDelta = *body.USDAmount
		tokenDelta = store.USDToTokens(usdDelta)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token_amount_or_usd_amount_required"})
		return
	}
	account, err := store.AdjustBalance(c.Request.Context(), body.APIKey, tokenDelta, usdDelta, body.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

func (h *Handler) PostQuotaRedeem(c *gin.Context) {
	store := h.quota()
	if store == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "quota_store_not_configured"})
		return
	}
	var body struct {
		APIKey string `json:"api_key"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.APIKey) == "" || strings.TrimSpace(body.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	account, err := store.RedeemCode(c.Request.Context(), body.APIKey, body.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

func randomCode() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	text := strings.ToUpper(hex.EncodeToString(buf))
	if len(text) < 12 {
		return "TOPUP-" + text
	}
	return "TOPUP-" + text[:6] + "-" + text[6:12]
}
