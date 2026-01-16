package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ThemeConfig struct {
	ColorActive    string `yaml:"color_active"`
	ColorExpiring  string `yaml:"color_expiring"`
	ColorInactive  string `yaml:"color_inactive"`
	ColorPending   string `yaml:"color_pending"`
	ColorError     string `yaml:"color_error"`
	ColorHighlight string `yaml:"color_highlight"`
	ColorBorder    string `yaml:"color_border"`
}

type Config struct {
	DefaultDuration     int         `yaml:"default_duration"`
	DurationPresets     []int       `yaml:"duration_presets"`
	LogLevel            string      `yaml:"log_level"`
	AutoRefreshInterval int         `yaml:"auto_refresh_interval"`
	AutoRefreshEnabled  bool        `yaml:"auto_refresh_enabled"`
	Theme               ThemeConfig `yaml:"theme"`
}

func DefaultTheme() ThemeConfig {
	return ThemeConfig{
		ColorActive:    "#00ff00", // Green
		ColorExpiring:  "#ffff00", // Yellow
		ColorInactive:  "#808080", // Gray
		ColorPending:   "#00bfff", // Blue
		ColorError:     "#ff0000", // Red
		ColorHighlight: "#7d56f4", // Purple
		ColorBorder:    "#444444",
	}
}

func Default() Config {
	return Config{
		DefaultDuration:     4,
		DurationPresets:     []int{1, 2, 4, 8},
		LogLevel:            "info",
		AutoRefreshInterval: 60,
		AutoRefreshEnabled:  true,
		Theme:               DefaultTheme(),
	}
}

func Load() (Config, error) {
	cfg := Default()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return cfg, nil // Return defaults if we can't find config dir
	}

	configPath := filepath.Join(configDir, "pim-tui", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if config doesn't exist
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
