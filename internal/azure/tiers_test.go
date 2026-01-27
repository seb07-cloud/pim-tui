package azure

import (
	"strings"
	"sync"
	"testing"
)

// TestGetEntraTier_KnownRoles tests tier lookup for well-known Entra roles
func TestGetEntraTier_KnownRoles(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		expectedTier   string
		expectedName   string
		expectedPath   string // PathType should be present for tier 0
	}{
		{
			name:           "Global Administrator",
			roleID:         "62e90394-69f5-4237-9190-012177145e10",
			expectedTier:   "0",
			expectedName:   "Global Administrator",
			expectedPath:   "Direct",
		},
		{
			name:           "Privileged Role Administrator",
			roleID:         "e8611ab8-c189-46e8-94e1-60213ab1f814",
			expectedTier:   "0",
			expectedName:   "Privileged Role Administrator",
			expectedPath:   "Direct",
		},
		{
			name:           "Attack Simulation Administrator",
			roleID:         "c430b396-e693-46cc-96f3-db01bf8bb62a",
			expectedTier:   "1",
			expectedName:   "Attack Simulation Administrator",
			expectedPath:   "", // Tier 1 may not have PathType
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier, ok := GetEntraTier(tt.roleID)
			if !ok {
				t.Errorf("GetEntraTier(%s) returned ok=false, expected role to be found", tt.roleID)
				return
			}
			if tier.Tier != tt.expectedTier {
				t.Errorf("GetEntraTier(%s) tier = %s, want %s", tt.roleID, tier.Tier, tt.expectedTier)
			}
			if tier.AssetName != tt.expectedName {
				t.Errorf("GetEntraTier(%s) assetName = %s, want %s", tt.roleID, tier.AssetName, tt.expectedName)
			}
			if tt.expectedPath != "" && tier.PathType != tt.expectedPath {
				t.Errorf("GetEntraTier(%s) pathType = %s, want %s", tt.roleID, tier.PathType, tt.expectedPath)
			}
			// Verify documentation URI is present
			if tier.DocumentationURI == "" {
				t.Errorf("GetEntraTier(%s) documentationUri is empty", tt.roleID)
			}
		})
	}
}

// TestGetAzureTier_KnownRoles tests tier lookup for well-known Azure RBAC roles
func TestGetAzureTier_KnownRoles(t *testing.T) {
	tests := []struct {
		name         string
		roleID       string
		expectedTier string
		expectedName string
	}{
		{
			name:         "Owner",
			roleID:       "8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
			expectedTier: "0",
			expectedName: "Owner",
		},
		{
			name:         "Contributor",
			roleID:       "b24988ac-6180-42a0-ab88-20f7382dd24c",
			expectedTier: "0", // Contributor is tier 0 in azure-tiering
			expectedName: "Contributor",
		},
		{
			name:         "Reader",
			roleID:       "acdd72a7-3385-48ef-bd42-f606fba81ae7",
			expectedTier: "1",
			expectedName: "Reader",
		},
		{
			name:         "Backup Contributor",
			roleID:       "5e467623-bb1f-42f4-a55d-6e525e11384b",
			expectedTier: "2",
			expectedName: "Backup Contributor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier, ok := GetAzureTier(tt.roleID)
			if !ok {
				t.Errorf("GetAzureTier(%s) returned ok=false, expected role to be found", tt.roleID)
				return
			}
			if tier.Tier != tt.expectedTier {
				t.Errorf("GetAzureTier(%s) tier = %s, want %s", tt.roleID, tier.Tier, tt.expectedTier)
			}
			if tier.AssetName != tt.expectedName {
				t.Errorf("GetAzureTier(%s) assetName = %s, want %s", tt.roleID, tier.AssetName, tt.expectedName)
			}
		})
	}
}

// TestGetEntraTier_UnknownRole tests that unknown role IDs return ok=false gracefully
func TestGetEntraTier_UnknownRole(t *testing.T) {
	tier, ok := GetEntraTier("00000000-0000-0000-0000-000000000000")
	if ok {
		t.Errorf("GetEntraTier(unknown) returned ok=true, want false")
	}
	// Verify it returns zero value TierMetadata
	if tier.Tier != "" || tier.AssetName != "" {
		t.Errorf("GetEntraTier(unknown) returned non-zero TierMetadata: %+v", tier)
	}
}

// TestGetAzureTier_UnknownRole tests that unknown Azure role IDs return ok=false gracefully
func TestGetAzureTier_UnknownRole(t *testing.T) {
	tier, ok := GetAzureTier("00000000-0000-0000-0000-000000000000")
	if ok {
		t.Errorf("GetAzureTier(unknown) returned ok=true, want false")
	}
	// Verify it returns zero value TierMetadata
	if tier.Tier != "" || tier.AssetName != "" {
		t.Errorf("GetAzureTier(unknown) returned non-zero TierMetadata: %+v", tier)
	}
}

// TestGetEntraTier_CaseInsensitive tests that lookups work regardless of UUID case
func TestGetEntraTier_CaseInsensitive(t *testing.T) {
	// Global Administrator UUID in different cases
	lowerUUID := "62e90394-69f5-4237-9190-012177145e10"
	upperUUID := strings.ToUpper(lowerUUID)
	mixedUUID := "62E90394-69f5-4237-9190-012177145E10"

	tests := []struct {
		name   string
		roleID string
	}{
		{"lowercase", lowerUUID},
		{"uppercase", upperUUID},
		{"mixed", mixedUUID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier, ok := GetEntraTier(tt.roleID)
			if !ok {
				t.Errorf("GetEntraTier(%s) returned ok=false, want true", tt.roleID)
				return
			}
			if tier.Tier != "0" {
				t.Errorf("GetEntraTier(%s) tier = %s, want 0", tt.roleID, tier.Tier)
			}
			if tier.AssetName != "Global Administrator" {
				t.Errorf("GetEntraTier(%s) assetName = %s, want Global Administrator", tt.roleID, tier.AssetName)
			}
		})
	}
}

// TestInitTiers_CalledOnce tests that sync.Once prevents concurrent initialization issues
func TestInitTiers_CalledOnce(t *testing.T) {
	// Reset initialization for this test (careful: affects other tests if run in parallel)
	// This is a concurrency validation test
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]TierMetadata, numGoroutines)

	// Launch multiple goroutines calling GetEntraTier concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			tier, ok := GetEntraTier("62e90394-69f5-4237-9190-012177145e10")
			if ok {
				results[index] = tier
			}
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got consistent results
	expectedName := "Global Administrator"
	for i, tier := range results {
		if tier.AssetName != expectedName {
			t.Errorf("Goroutine %d got assetName=%s, want %s", i, tier.AssetName, expectedName)
		}
	}
}

// TestGetAzureTier_CaseInsensitive tests Azure role lookup case insensitivity
func TestGetAzureTier_CaseInsensitive(t *testing.T) {
	// Owner UUID in different cases
	lowerUUID := "8e3af657-a8ff-443c-a75c-2fe8c4bcb635"
	upperUUID := strings.ToUpper(lowerUUID)

	lowerTier, lowerOk := GetAzureTier(lowerUUID)
	upperTier, upperOk := GetAzureTier(upperUUID)

	if !lowerOk || !upperOk {
		t.Errorf("GetAzureTier case-insensitive lookup failed: lower=%v, upper=%v", lowerOk, upperOk)
		return
	}

	if lowerTier.AssetName != upperTier.AssetName {
		t.Errorf("GetAzureTier case mismatch: lower=%s, upper=%s", lowerTier.AssetName, upperTier.AssetName)
	}
}
