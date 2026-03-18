package quota

import (
	"context"
	"strings"

	coreusage "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
	log "github.com/sirupsen/logrus"
)

type UsagePlugin struct{}

func NewUsagePlugin() *UsagePlugin { return &UsagePlugin{} }

func (p *UsagePlugin) HandleUsage(ctx context.Context, record coreusage.Record) {
	store := DefaultStore()
	if store == nil {
		return
	}
	apiKey := strings.TrimSpace(record.APIKey)
	if apiKey == "" {
		return
	}
	tokens := record.Detail.TotalTokens
	if tokens <= 0 {
		return
	}
	if _, err := store.ConsumeTokens(ctx, apiKey, tokens, record.Model, ""); err != nil {
		log.Warnf("quota consume failed for api key %s: %v", apiKey, err)
	}
}
