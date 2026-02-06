package specs

import (
	"encoding/json"
	"fmt"
	"os"
)

// WatchRule represents a single file monitoring rule with path, events, and action
type WatchRule struct {
	Path   string   `json:"path"`
	Events []string `json:"events"`
	Action string   `json:"action"`
}

// MonitorConfig represents the overall configuration structure
type MonitorConfig struct {
	WatchRules []WatchRule `json:"watch_rules"`
}

// LoadConfig reads and parses the JSON configuration file
func LoadConfigFanotify(filename string) (*MonitorConfig, error) {
	// Read the configuration file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON into MonitorConfig struct
	var config MonitorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate that we have at least one rule
	if len(config.WatchRules) == 0 {
		return nil, fmt.Errorf("config must contain at least one watch rule")
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *MonitorConfig) Validate() error {
	for i, rule := range c.WatchRules {
		if rule.Path == "" {
			return fmt.Errorf("watch_rule[%d]: path cannot be empty", i)
		}
		if len(rule.Events) == 0 {
			return fmt.Errorf("watch_rule[%d]: must specify at least one event", i)
		}
		if rule.Action == "" {
			return fmt.Errorf("watch_rule[%d]: action cannot be empty", i)
		}

		// Validate event types
		validEvents := map[string]bool{
			"open":   true,
			"read":   true,
			"write":  true,
			"close":  true,
			"access": true,
			"exec": true,
		}
		for _, event := range rule.Events {
			if !validEvents[event] {
				return fmt.Errorf("watch_rule[%d]: invalid event type '%s'", i, event)
			}
		}

		// Validate action types
		validActions := map[string]bool{
			"audit": true,
			"block": true,
			"alert": true,
		}
		if !validActions[rule.Action] {
			return fmt.Errorf("watch_rule[%d]: invalid action '%s'", i, rule.Action)
		}
	}
	return nil
}