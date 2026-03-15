package ui

import (
	"regexp"
	"strings"

	"github.com/seb07-cloud/pim-tui/internal/azure"
	"github.com/seb07-cloud/pim-tui/internal/config"
)

// ResolvedEntry is a ProfileEntry matched to actual Azure data.
type ResolvedEntry struct {
	Entry     config.ProfileEntry
	Matched   bool
	MatchedAs interface{} // azure.Role | azure.Group | SubscriptionRoleActivation
	Warning   string
}

// ResolvedProfile is a fully validated profile ready for activation.
type ResolvedProfile struct {
	Profile  config.Profile
	Entries  []ResolvedEntry
	AllValid bool
	MaxTier  int // 0-3, -1 if no tier data
	Variables []string
}

var templateVarPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// extractVariables finds all {{variable}} placeholders in a template string.
// Returns deduplicated variable names in order of first appearance.
func extractVariables(template string) []string {
	matches := templateVarPattern.FindAllStringSubmatch(template, -1)
	seen := make(map[string]bool)
	var vars []string
	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}

// substituteVariables replaces each {{key}} with the corresponding value from the map.
func substituteVariables(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

// resolveProfile matches profile entries against loaded Azure data.
func resolveProfile(p config.Profile, roles []azure.Role, groups []azure.Group, subs []azure.LighthouseSubscription) ResolvedProfile {
	rp := ResolvedProfile{
		Profile:  p,
		AllValid: true,
		MaxTier:  -1,
		Variables: extractVariables(p.Justification),
	}

	for _, entry := range p.Entries {
		re := ResolvedEntry{Entry: entry}

		switch entry.Type {
		case "role":
			re = matchRole(entry, roles)
		case "group":
			re = matchGroup(entry, groups)
		case "subscription-role":
			re = matchSubscriptionRole(entry, subs)
		default:
			re.Warning = "unknown entry type: " + entry.Type
		}

		if !re.Matched {
			rp.AllValid = false
		}

		// Update max tier from matched entry
		updateMaxTier(&rp, re)

		rp.Entries = append(rp.Entries, re)
	}

	return rp
}

func matchRole(entry config.ProfileEntry, roles []azure.Role) ResolvedEntry {
	re := ResolvedEntry{Entry: entry}
	lowerName := strings.ToLower(entry.Name)

	// Try display name match first (case-insensitive)
	for _, role := range roles {
		if strings.ToLower(role.DisplayName) == lowerName {
			re.Matched = true
			re.MatchedAs = role
			return re
		}
	}

	// Fall back to RoleDefinitionID if provided
	if entry.RoleDefinitionID != "" {
		lowerID := strings.ToLower(entry.RoleDefinitionID)
		for _, role := range roles {
			if strings.ToLower(role.RoleDefinitionID) == lowerID {
				re.Matched = true
				re.MatchedAs = role
				return re
			}
		}
	}

	re.Warning = "no matching eligible role found"
	return re
}

func matchGroup(entry config.ProfileEntry, groups []azure.Group) ResolvedEntry {
	re := ResolvedEntry{Entry: entry}
	lowerName := strings.ToLower(entry.Name)

	for _, group := range groups {
		if strings.ToLower(group.DisplayName) == lowerName {
			re.Matched = true
			re.MatchedAs = group
			return re
		}
	}

	if entry.RoleDefinitionID != "" {
		lowerID := strings.ToLower(entry.RoleDefinitionID)
		for _, group := range groups {
			if strings.ToLower(group.ID) == lowerID {
				re.Matched = true
				re.MatchedAs = group
				return re
			}
		}
	}

	re.Warning = "no matching eligible group found"
	return re
}

func matchSubscriptionRole(entry config.ProfileEntry, subs []azure.LighthouseSubscription) ResolvedEntry {
	re := ResolvedEntry{Entry: entry}
	lowerSubName := strings.ToLower(entry.Subscription)
	lowerRoleName := strings.ToLower(entry.Name)

	for _, sub := range subs {
		if strings.ToLower(sub.DisplayName) != lowerSubName {
			continue
		}
		// Found the subscription, now find the role
		for _, role := range sub.EligibleRoles {
			if strings.ToLower(role.RoleDefinitionName) == lowerRoleName {
				re.Matched = true
				re.MatchedAs = SubscriptionRoleActivation{
					SubscriptionID:   sub.ID,
					SubscriptionName: sub.DisplayName,
					Role:             role,
				}
				return re
			}
		}
		// Subscription found but role not found — try ID fallback
		if entry.RoleDefinitionID != "" {
			lowerID := strings.ToLower(entry.RoleDefinitionID)
			for _, role := range sub.EligibleRoles {
				if strings.ToLower(role.RoleDefinitionID) == lowerID || strings.HasSuffix(strings.ToLower(role.RoleDefinitionID), "/"+lowerID) {
					re.Matched = true
					re.MatchedAs = SubscriptionRoleActivation{
						SubscriptionID:   sub.ID,
						SubscriptionName: sub.DisplayName,
						Role:             role,
					}
					return re
				}
			}
		}
		re.Warning = "role not found in subscription " + sub.DisplayName
		return re
	}

	re.Warning = "subscription not found: " + entry.Subscription
	return re
}

func updateMaxTier(rp *ResolvedProfile, re ResolvedEntry) {
	if !re.Matched {
		return
	}

	tierStr := ""
	switch v := re.MatchedAs.(type) {
	case azure.Role:
		if tier, found := azure.GetEntraTier(v.RoleDefinitionID); found {
			tierStr = tier.Tier
		}
	case azure.Group:
		// Check linked Entra roles for tier
		for _, lr := range v.LinkedRoles {
			if tier, found := azure.GetEntraTier(lr.RoleDefinitionID); found {
				if tierStr == "" || tier.Tier < tierStr {
					tierStr = tier.Tier
				}
			}
		}
		// Check linked Azure RBAC roles
		for _, lar := range v.LinkedAzureRBac {
			if tier, found := azure.GetAzureTier(lar.RoleDefinitionID); found {
				if tierStr == "" || tier.Tier < tierStr {
					tierStr = tier.Tier
				}
			}
		}
	case SubscriptionRoleActivation:
		if tier, found := azure.GetAzureTier(v.Role.RoleDefinitionID); found {
			tierStr = tier.Tier
		}
	}

	if tierStr == "" {
		return
	}

	tierNum := int(tierStr[0] - '0')
	if tierNum >= 0 && tierNum <= 3 {
		if rp.MaxTier == -1 || tierNum < rp.MaxTier {
			rp.MaxTier = tierNum
		}
	}
}

// profileTierBadgeStr returns the tier string for a resolved profile, or "" if no tier.
func profileTierBadgeStr(rp *ResolvedProfile) string {
	if rp.MaxTier < 0 || rp.MaxTier > 3 {
		return ""
	}
	return string(rune('0' + rp.MaxTier))
}
