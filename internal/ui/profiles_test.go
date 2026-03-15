package ui

import (
	"testing"

	"github.com/seb07-cloud/pim-tui/internal/azure"
	"github.com/seb07-cloud/pim-tui/internal/config"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{"empty string", "", nil},
		{"no variables", "plain text", nil},
		{"single variable", "ticket: {{ticket_id}}", []string{"ticket_id"}},
		{"multiple variables", "{{ticket}} - {{desc}}", []string{"ticket", "desc"}},
		{"duplicate deduplicated", "{{a}} and {{a}} again", []string{"a"}},
		{"mixed text and vars", "P1: {{ticket}} on {{env}} for {{ticket}}", []string{"ticket", "env"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVariables(tt.template)
			if len(got) != len(tt.want) {
				t.Errorf("extractVariables(%q) = %v, want %v", tt.template, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractVariables(%q)[%d] = %q, want %q", tt.template, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		values   map[string]string
		want     string
	}{
		{
			"single replacement",
			"ticket: {{id}}",
			map[string]string{"id": "INC-123"},
			"ticket: INC-123",
		},
		{
			"multiple replacements",
			"{{ticket}} on {{env}}",
			map[string]string{"ticket": "INC-456", "env": "prod"},
			"INC-456 on prod",
		},
		{
			"missing key left as-is",
			"{{known}} and {{unknown}}",
			map[string]string{"known": "yes"},
			"yes and {{unknown}}",
		},
		{
			"no variables",
			"plain text",
			map[string]string{"key": "val"},
			"plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteVariables(tt.template, tt.values)
			if got != tt.want {
				t.Errorf("substituteVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveProfile_AllMatch(t *testing.T) {
	roles := []azure.Role{
		{DisplayName: "Global Reader", RoleDefinitionID: "def-1"},
		{DisplayName: "Security Reader", RoleDefinitionID: "def-2"},
	}
	groups := []azure.Group{
		{ID: "grp-1", DisplayName: "SRE Access"},
	}
	subs := []azure.LighthouseSubscription{
		{
			ID:          "sub-1",
			DisplayName: "Production",
			EligibleRoles: []azure.EligibleAzureRole{
				{RoleDefinitionName: "Contributor", RoleDefinitionID: "contrib-id"},
			},
		},
	}

	p := config.Profile{
		Name:          "Test Profile",
		Justification: "Reason: {{ticket}}",
		Entries: []config.ProfileEntry{
			{Type: "role", Name: "Global Reader"},
			{Type: "group", Name: "SRE Access"},
			{Type: "subscription-role", Name: "Contributor", Subscription: "Production"},
		},
	}

	rp := resolveProfile(p, roles, groups, subs)

	if !rp.AllValid {
		t.Error("resolveProfile() AllValid = false, want true")
	}
	if len(rp.Entries) != 3 {
		t.Fatalf("resolveProfile() got %d entries, want 3", len(rp.Entries))
	}
	for i, e := range rp.Entries {
		if !e.Matched {
			t.Errorf("Entry[%d] Matched = false, want true (warning: %s)", i, e.Warning)
		}
	}
	if len(rp.Variables) != 1 || rp.Variables[0] != "ticket" {
		t.Errorf("Variables = %v, want [ticket]", rp.Variables)
	}
}

func TestResolveProfile_MissingEntries(t *testing.T) {
	roles := []azure.Role{
		{DisplayName: "Global Reader", RoleDefinitionID: "def-1"},
	}

	p := config.Profile{
		Name: "Missing",
		Entries: []config.ProfileEntry{
			{Type: "role", Name: "Global Reader"},
			{Type: "role", Name: "Nonexistent Role"},
			{Type: "group", Name: "No Such Group"},
		},
	}

	rp := resolveProfile(p, roles, nil, nil)

	if rp.AllValid {
		t.Error("resolveProfile() AllValid = true, want false")
	}
	if rp.Entries[0].Matched != true {
		t.Error("Entry[0] should match")
	}
	if rp.Entries[1].Matched != false {
		t.Error("Entry[1] should not match")
	}
	if rp.Entries[2].Matched != false {
		t.Error("Entry[2] should not match")
	}
}

func TestResolveProfile_CaseInsensitive(t *testing.T) {
	roles := []azure.Role{
		{DisplayName: "Global Reader", RoleDefinitionID: "def-1"},
	}

	p := config.Profile{
		Entries: []config.ProfileEntry{
			{Type: "role", Name: "global reader"},
		},
	}

	rp := resolveProfile(p, roles, nil, nil)

	if !rp.AllValid || !rp.Entries[0].Matched {
		t.Error("Case-insensitive match should succeed")
	}
}

func TestResolveProfile_IDFallback(t *testing.T) {
	roles := []azure.Role{
		{DisplayName: "Some Role", RoleDefinitionID: "abc-123"},
	}

	p := config.Profile{
		Entries: []config.ProfileEntry{
			{Type: "role", Name: "Wrong Name", RoleDefinitionID: "abc-123"},
		},
	}

	rp := resolveProfile(p, roles, nil, nil)

	if !rp.AllValid || !rp.Entries[0].Matched {
		t.Error("ID fallback match should succeed")
	}
}

func TestResolveProfile_SubscriptionNotFound(t *testing.T) {
	p := config.Profile{
		Entries: []config.ProfileEntry{
			{Type: "subscription-role", Name: "Contributor", Subscription: "Missing Sub"},
		},
	}

	rp := resolveProfile(p, nil, nil, nil)

	if rp.AllValid {
		t.Error("Should fail when subscription not found")
	}
	if rp.Entries[0].Warning == "" {
		t.Error("Should have warning for missing subscription")
	}
}

func TestResolveProfile_UnknownType(t *testing.T) {
	p := config.Profile{
		Entries: []config.ProfileEntry{
			{Type: "invalid-type", Name: "Something"},
		},
	}

	rp := resolveProfile(p, nil, nil, nil)

	if rp.AllValid {
		t.Error("Should fail for unknown entry type")
	}
	if rp.Entries[0].Warning == "" {
		t.Error("Should have warning for unknown type")
	}
}

func TestProfileTierBadgeStr(t *testing.T) {
	tests := []struct {
		name    string
		maxTier int
		want    string
	}{
		{"tier 0", 0, "0"},
		{"tier 1", 1, "1"},
		{"tier 3", 3, "3"},
		{"no tier", -1, ""},
		{"out of range", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &ResolvedProfile{MaxTier: tt.maxTier}
			got := profileTierBadgeStr(rp)
			if got != tt.want {
				t.Errorf("profileTierBadgeStr() = %q, want %q", got, tt.want)
			}
		})
	}
}
