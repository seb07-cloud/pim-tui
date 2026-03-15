package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfiles_MissingFile(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalHome)

	pc, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v, want nil", err)
	}
	if len(pc.Profiles) != 0 {
		t.Errorf("LoadProfiles() got %d profiles, want 0", len(pc.Profiles))
	}
}

func TestLoadProfiles_ValidFile(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalHome)

	configDir := filepath.Join(tempDir, "pim-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	content := `
profiles:
  - name: "Incident Response"
    description: "Emergency access for P1 incidents"
    duration: 2
    justification: "P1 Incident: {{ticket_id}} - {{description}}"
    entries:
      - type: role
        name: "Global Reader"
      - type: group
        name: "SRE Emergency Access"
      - type: subscription-role
        name: "Contributor"
        subscription: "Production Sub"
        role_definition_id: "b24988ac-6180-42a0-ab88-20f7382dd24c"
  - name: "Weekly Deploy"
    description: "Standard release process"
    duration: 4
    justification: "Scheduled deploy"
    entries:
      - type: role
        name: "Application Administrator"
`
	profilesPath := filepath.Join(configDir, "profiles.yaml")
	if err := os.WriteFile(profilesPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write profiles file: %v", err)
	}

	pc, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v, want nil", err)
	}
	if len(pc.Profiles) != 2 {
		t.Fatalf("LoadProfiles() got %d profiles, want 2", len(pc.Profiles))
	}

	// Verify first profile
	p := pc.Profiles[0]
	if p.Name != "Incident Response" {
		t.Errorf("Profile[0].Name = %q, want %q", p.Name, "Incident Response")
	}
	if p.Duration != 2 {
		t.Errorf("Profile[0].Duration = %d, want 2", p.Duration)
	}
	if p.Justification != "P1 Incident: {{ticket_id}} - {{description}}" {
		t.Errorf("Profile[0].Justification = %q, want template", p.Justification)
	}
	if len(p.Entries) != 3 {
		t.Fatalf("Profile[0].Entries = %d, want 3", len(p.Entries))
	}

	// Check entry types
	if p.Entries[0].Type != "role" || p.Entries[0].Name != "Global Reader" {
		t.Errorf("Entry[0] = %+v, want role/Global Reader", p.Entries[0])
	}
	if p.Entries[1].Type != "group" || p.Entries[1].Name != "SRE Emergency Access" {
		t.Errorf("Entry[1] = %+v, want group/SRE Emergency Access", p.Entries[1])
	}
	if p.Entries[2].Type != "subscription-role" || p.Entries[2].Subscription != "Production Sub" {
		t.Errorf("Entry[2] = %+v, want subscription-role on Production Sub", p.Entries[2])
	}
	if p.Entries[2].RoleDefinitionID != "b24988ac-6180-42a0-ab88-20f7382dd24c" {
		t.Errorf("Entry[2].RoleDefinitionID = %q, want ID", p.Entries[2].RoleDefinitionID)
	}

	// Verify second profile
	p2 := pc.Profiles[1]
	if p2.Name != "Weekly Deploy" {
		t.Errorf("Profile[1].Name = %q, want %q", p2.Name, "Weekly Deploy")
	}
	if len(p2.Entries) != 1 {
		t.Errorf("Profile[1].Entries = %d, want 1", len(p2.Entries))
	}
}

func TestLoadProfiles_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalHome)

	configDir := filepath.Join(tempDir, "pim-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	invalidYAML := `
profiles:
  - name: [invalid
    bad_indent:
`
	profilesPath := filepath.Join(configDir, "profiles.yaml")
	if err := os.WriteFile(profilesPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write profiles file: %v", err)
	}

	_, err := LoadProfiles()
	if err == nil {
		t.Errorf("LoadProfiles() error = nil, want YAML parse error")
	}
}

func TestLoadProfiles_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalHome)

	configDir := filepath.Join(tempDir, "pim-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	profilesPath := filepath.Join(configDir, "profiles.yaml")
	if err := os.WriteFile(profilesPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write profiles file: %v", err)
	}

	pc, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v, want nil", err)
	}
	if len(pc.Profiles) != 0 {
		t.Errorf("LoadProfiles() got %d profiles, want 0", len(pc.Profiles))
	}
}
