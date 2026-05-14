package health

type Scores struct {
	Overall        int
	Security       int
	Infrastructure int
	Configuration  int
}

type Finding struct {
	Severity string
	Category string
}

func CalculateScores(findings []Finding) Scores {
	overall := 100
	security := 100
	infra := 100
	config := 100

	for _, finding := range findings {
		penalty := severityPenalty(finding.Severity)
		overall -= penalty

		switch finding.Category {
		case "secrets":
			security -= penalty
		case "docker", "cicd", "kubernetes", "redis", "postgres":
			infra -= penalty
		case "env", "custom":
			config -= penalty
		}
	}

	return Scores{
		Overall:        clampScore(overall),
		Security:       clampScore(security),
		Infrastructure: clampScore(infra),
		Configuration:  clampScore(config),
	}
}

func severityPenalty(sev string) int {
	switch sev {
	case "critical":
		return 15
	case "warning":
		return 5
	case "info":
		return 1
	default:
		return 0
	}
}

func clampScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
