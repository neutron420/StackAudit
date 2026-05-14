package docker

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func AddRestartPolicies(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no services defined in docker-compose.yml")
	}

	changed := false
	for _, entry := range services {
		service, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if _, ok := service["restart"]; !ok {
			service["restart"] = "unless-stopped"
			changed = true
		}
	}

	if !changed {
		return nil
	}

	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(strings.TrimSpace(string(out))+"\n"), 0o644)
}
