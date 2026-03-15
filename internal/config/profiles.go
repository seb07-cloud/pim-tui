package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProfileEntry represents a single role, group, or subscription-role in a profile.
type ProfileEntry struct {
	Type             string `yaml:"type"`               // "role", "group", or "subscription-role"
	Name             string `yaml:"name"`               // Display name (primary match key)
	RoleDefinitionID string `yaml:"role_definition_id"` // Optional fallback ID for precise matching
	Subscription     string `yaml:"subscription"`       // Required for type "subscription-role"
}

// Profile represents a named activation scenario.
type Profile struct {
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Duration      int            `yaml:"duration"`      // Hours (0 = use global default)
	Justification string         `yaml:"justification"` // Template with {{variable}} placeholders
	Entries       []ProfileEntry `yaml:"entries"`
}

// ProfilesConfig is the top-level struct for profiles.yaml.
type ProfilesConfig struct {
	Profiles []Profile `yaml:"profiles"`
}

// LoadProfiles loads activation profiles from ~/.config/pim-tui/profiles.yaml.
// Returns an empty ProfilesConfig if the file doesn't exist.
func LoadProfiles() (ProfilesConfig, error) {
	var pc ProfilesConfig

	configDir, err := os.UserConfigDir()
	if err != nil {
		return pc, nil
	}

	profilesPath := filepath.Join(configDir, "pim-tui", "profiles.yaml")
	data, err := os.ReadFile(profilesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return pc, nil
		}
		return pc, err
	}

	if err := yaml.Unmarshal(data, &pc); err != nil {
		return pc, err
	}

	return pc, nil
}
