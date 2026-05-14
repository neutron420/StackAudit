package env

import (
	"bufio"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func parseEnvFile(path string) (FileEnv, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return FileEnv{}, err
	}

	vars, err := godotenv.Parse(strings.NewReader(string(content)))
	if err != nil {
		return FileEnv{}, err
	}

	seen := map[string]bool{}
	duplicates := []string{}
	empty := []string{}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 0 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		if seen[key] {
			duplicates = append(duplicates, key)
		}
		seen[key] = true
		if len(parts) == 1 || strings.TrimSpace(parts[1]) == "" {
			empty = append(empty, key)
		}
	}

	return FileEnv{
		Path:       path,
		Vars:       vars,
		Duplicates: duplicates,
		Empty:      empty,
	}, scanner.Err()
}
