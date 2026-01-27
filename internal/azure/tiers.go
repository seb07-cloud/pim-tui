package azure

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

// Embed tier data files at compile time
//
//go:embed data/tiered-entra-roles.json
var entraTierData []byte

//go:embed data/tiered-azure-roles.json
var azureTierData []byte

var (
	entraTiers map[string]TierMetadata
	azureTiers map[string]TierMetadata
	once       sync.Once
	parseErr   error
)

// initTiers parses embedded JSON files once on first access
func initTiers() {
	entraTiers = make(map[string]TierMetadata)
	azureTiers = make(map[string]TierMetadata)

	// Parse Entra roles
	var entraRoles []TierMetadata
	if err := json.Unmarshal(entraTierData, &entraRoles); err != nil {
		parseErr = err
		return
	}
	for _, role := range entraRoles {
		// Normalize to lowercase for case-insensitive lookups
		entraTiers[strings.ToLower(role.ID)] = role
	}

	// Parse Azure roles
	var azureRoles []TierMetadata
	if err := json.Unmarshal(azureTierData, &azureRoles); err != nil {
		parseErr = err
		return
	}
	for _, role := range azureRoles {
		// Normalize to lowercase for case-insensitive lookups
		azureTiers[strings.ToLower(role.ID)] = role
	}
}

// GetEntraTier returns tier metadata for an Entra role ID.
// Returns (metadata, true) if found, (zero value, false) if not found or parse error.
func GetEntraTier(roleID string) (TierMetadata, bool) {
	once.Do(initTiers)
	if parseErr != nil {
		return TierMetadata{}, false
	}
	tier, ok := entraTiers[strings.ToLower(roleID)]
	return tier, ok
}

// GetAzureTier returns tier metadata for an Azure RBAC role ID.
// Returns (metadata, true) if found, (zero value, false) if not found or parse error.
func GetAzureTier(roleID string) (TierMetadata, bool) {
	once.Do(initTiers)
	if parseErr != nil {
		return TierMetadata{}, false
	}
	tier, ok := azureTiers[strings.ToLower(roleID)]
	return tier, ok
}
