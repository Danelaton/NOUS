package config

// OpenCodeProfile represents an SDD profile for OpenCode
type OpenCodeProfile struct {
	Name    string `json:"name"`
	Phases  map[string]PhaseConfig `json:"phases"`
	Default string `json:"default"`
}

// PhaseConfig defines model routing per SDD phase
type PhaseConfig struct {
	Model   string `json:"model"`
	Provider string `json:"provider"`
}

var DefaultProfiles = map[string]OpenCodeProfile{
	"balanced": {
		Name:    "Balanced",
		Default: "sdd-orchestrator",
		Phases: map[string]PhaseConfig{
			"sdd-design":     {Model: "claude-sonnet-4", Provider: "anthropic"},
			"sdd-implement":  {Model: "claude-sonnet-4", Provider: "anthropic"},
			"sdd-verify":     {Model: "gpt-4o", Provider: "openai"},
			"sdd-document":   {Model: "gpt-4o-mini", Provider: "openai"},
		},
	},
	"fast": {
		Name:    "Fast",
		Default: "sdd-orchestrator",
		Phases: map[string]PhaseConfig{
			"sdd-design":     {Model: "claude-haiku-4", Provider: "anthropic"},
			"sdd-implement":  {Model: "claude-haiku-4", Provider: "anthropic"},
			"sdd-verify":     {Model: "gpt-4o-mini", Provider: "openai"},
			"sdd-document":   {Model: "gpt-4o-mini", Provider: "openai"},
		},
	},
	"quality": {
		Name:    "Quality",
		Default: "sdd-orchestrator",
		Phases: map[string]PhaseConfig{
			"sdd-design":     {Model: "claude-opus-4", Provider: "anthropic"},
			"sdd-implement":   {Model: "claude-sonnet-4", Provider: "anthropic"},
			"sdd-verify":     {Model: "gpt-4o", Provider: "openai"},
			"sdd-document":    {Model: "claude-sonnet-4", Provider: "anthropic"},
		},
	},
}

// GetProfile returns a profile by name
func GetProfile(name string) *OpenCodeProfile {
	if profile, ok := DefaultProfiles[name]; ok {
		return &profile
	}
	return nil
}

// ListProfiles returns all available profile names
func ListProfiles() []string {
	names := make([]string, 0, len(DefaultProfiles))
	for name := range DefaultProfiles {
		names = append(names, name)
	}
	return names
}