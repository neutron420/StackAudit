package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"stack/internal/rules"
	"stack/internal/scanner"
	"stack/internal/utils"

	"gopkg.in/yaml.v3"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "kubernetes"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}
	files, err := utils.WalkFiles(root, utils.WalkOptions{
		MaxSize: 1_000_000,
		Extensions: map[string]bool{
			".yaml": true,
			".yml":  true,
		},
		SkipDirs: defaultSkipDirs(),
		Ignore:   ignore,
	})
	if err != nil {
		return nil, err
	}

	findings := []scanner.Finding{}
	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		findings = append(findings, scanManifest(file, data, ruleSet)...)
	}
	if len(findings) == 0 {
		findings = append(findings, scanner.Finding{
			Category:    "kubernetes",
			Title:       "No Kubernetes manifests found",
			Description: "We couldn't find any .yaml or .yml files that look like Kubernetes manifests in your project.",
			Severity:    scanner.SeverityInfo,
		})
	}

	return findings, nil
}

func scanManifest(path string, data []byte, ruleSet rules.RuleSet) []scanner.Finding {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	findings := []scanner.Finding{}
	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return []scanner.Finding{{
				Severity:    scanner.SeverityCritical,
				Title:       "Kubernetes manifest is invalid",
				Description: err.Error(),
				File:        path,
				Category:    "kubernetes",
				RuleID:      "k8s_invalid_yaml",
			}}
		}
		if len(doc) == 0 || !isKubernetesManifest(doc) {
			continue
		}
		findings = append(findings, scanObject(path, doc, ruleSet)...)
	}
	return findings
}

func scanObject(path string, obj map[string]interface{}, ruleSet rules.RuleSet) []scanner.Finding {
	kind := stringValue(obj["kind"])
	name := objectName(obj)
	findings := []scanner.Finding{}

	podSpec := podSpecFor(obj)
	if podSpec != nil {
		if boolValue(podSpec["hostNetwork"]) || boolValue(podSpec["hostPID"]) || boolValue(podSpec["hostIPC"]) {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityCritical,
				Title:       fmt.Sprintf("%s %s uses host namespaces", kind, name),
				Description: "Avoid hostNetwork, hostPID, and hostIPC unless the workload truly needs node-level access.",
				File:        path,
				Category:    "kubernetes",
				RuleID:      "k8s_host_namespace",
			})
		}
		for _, container := range containersFor(podSpec) {
			findings = append(findings, scanContainer(path, kind, name, container, ruleSet)...)
		}
	}

	if strings.EqualFold(kind, "Service") {
		spec := mapValue(obj["spec"])
		if stringValue(spec["type"]) == "LoadBalancer" {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityInfo,
				Title:       fmt.Sprintf("Service %s is externally exposed", name),
				Description: "Confirm this LoadBalancer service is intended to be reachable outside the cluster.",
				File:        path,
				Category:    "kubernetes",
				RuleID:      "k8s_load_balancer_service",
			})
		}
	}

	return findings
}

func scanContainer(path, kind, workload string, container map[string]interface{}, ruleSet rules.RuleSet) []scanner.Finding {
	name := stringValue(container["name"])
	if name == "" {
		name = "container"
	}
	titlePrefix := fmt.Sprintf("%s %s container %s", kind, workload, name)
	findings := []scanner.Finding{}

	image := stringValue(container["image"])
	if image != "" && (!strings.Contains(image, ":") || strings.HasSuffix(image, ":latest")) && !ruleSet.LatestTagOK {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       titlePrefix + " uses latest image tag",
			Description: "Pin Kubernetes container image tags for reproducible deploys.",
			File:        path,
			Category:    "kubernetes",
			RuleID:      "k8s_latest_image",
		})
	}

	if resources := mapValue(container["resources"]); len(resources) == 0 {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       titlePrefix + " has no resource limits",
			Description: "Set CPU and memory requests/limits to protect cluster capacity.",
			File:        path,
			Category:    "kubernetes",
			RuleID:      "k8s_missing_resources",
		})
	}

	securityContext := mapValue(container["securityContext"])
	if boolValue(securityContext["privileged"]) {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityCritical,
			Title:       titlePrefix + " is privileged",
			Description: "Privileged containers can access host-level capabilities and should be avoided.",
			File:        path,
			Category:    "kubernetes",
			RuleID:      "k8s_privileged_container",
		})
	}
	if !boolValue(securityContext["runAsNonRoot"]) {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       titlePrefix + " can run as root",
			Description: "Set securityContext.runAsNonRoot: true for application containers.",
			File:        path,
			Category:    "kubernetes",
			RuleID:      "k8s_run_as_non_root",
		})
	}

	return findings
}

func isKubernetesManifest(obj map[string]interface{}) bool {
	return stringValue(obj["apiVersion"]) != "" && stringValue(obj["kind"]) != ""
}

func podSpecFor(obj map[string]interface{}) map[string]interface{} {
	kind := strings.ToLower(stringValue(obj["kind"]))
	spec := mapValue(obj["spec"])
	switch kind {
	case "pod":
		return spec
	case "cronjob":
		jobSpec := mapValue(mapValue(spec["jobTemplate"])["spec"])
		return mapValue(mapValue(jobSpec["template"])["spec"])
	default:
		template := mapValue(spec["template"])
		if len(template) == 0 {
			return nil
		}
		return mapValue(template["spec"])
	}
}

func containersFor(podSpec map[string]interface{}) []map[string]interface{} {
	var containers []map[string]interface{}
	for _, key := range []string{"containers", "initContainers"} {
		for _, entry := range sliceValue(podSpec[key]) {
			if container := mapValue(entry); len(container) > 0 {
				containers = append(containers, container)
			}
		}
	}
	return containers
}

func objectName(obj map[string]interface{}) string {
	name := stringValue(mapValue(obj["metadata"])["name"])
	if name == "" {
		return "unnamed"
	}
	return name
}

func mapValue(value interface{}) map[string]interface{} {
	result, _ := value.(map[string]interface{})
	return result
}

func sliceValue(value interface{}) []interface{} {
	result, _ := value.([]interface{})
	return result
}

func stringValue(value interface{}) string {
	result, _ := value.(string)
	return strings.TrimSpace(result)
}

func boolValue(value interface{}) bool {
	result, _ := value.(bool)
	return result
}

func defaultSkipDirs() map[string]bool {
	return map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
	}
}
