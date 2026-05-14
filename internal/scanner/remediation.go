package scanner

import "strings"

func ApplyRemediations(findings []Finding) []Finding {
	for i, finding := range findings {
		if strings.TrimSpace(finding.Remediation) != "" {
			continue
		}
		if remediation := remediationFor(finding.RuleID); remediation != "" {
			findings[i].Remediation = remediation
		}
	}
	return findings
}

func remediationFor(ruleID string) string {
	normalized := strings.ToLower(strings.TrimSpace(ruleID))
	if base := baseRuleID(normalized); base != "" {
		normalized = base
	}
	switch normalized {
	case "env_localhost":
		return "Use service DNS names or production hostnames instead of localhost in deployable env files."
	case "env_required":
		return "Add the missing key to committed examples and configure the real value in your deployment secrets."
	case "env_empty":
		return "Set a real value or remove the empty variable if the application no longer reads it."
	case "env_duplicate":
		return "Keep one definition for the variable so runtime configuration is predictable."
	case "env_unused":
		return "Remove the unused variable or wire it into the application if it is still required."
	case "secret_high_entropy":
		return "Move the suspected secret to a secret manager and rotate it if it may have been committed."
	case "docker_latest_tag":
		return "Pin the image to an explicit version or digest."
	case "docker_no_user":
		return "Add a non-root USER directive after installing runtime dependencies."
	case "docker_env_secret":
		return "Pass secrets at runtime through environment injection or a secret manager."
	case "docker_restart_policy":
		return "Add restart: unless-stopped or another intentional restart policy."
	case "cicd_permissions_write_all":
		return "Set the workflow or job permissions block to the minimum scopes required."
	case "cicd_pull_request_target":
		return "Use pull_request for untrusted code, or isolate pull_request_target jobs from checkout and secret access."
	case "cicd_checkout_v1":
		return "Upgrade actions/checkout to a current major version."
	case "k8s_latest_image":
		return "Pin container images to immutable tags or digests."
	case "k8s_missing_resources":
		return "Set CPU and memory requests and limits for each container."
	case "k8s_privileged_container":
		return "Remove privileged mode and grant only the specific capabilities the workload needs."
	case "k8s_run_as_non_root":
		return "Set securityContext.runAsNonRoot: true and use an image that supports a non-root user."
	case "k8s_host_namespace":
		return "Disable hostNetwork, hostPID, and hostIPC unless node-level access is required."
	case "k8s_load_balancer_service":
		return "Confirm the service must be public, or use ClusterIP/Ingress with controlled exposure."
	case "redis_protected_mode_disabled":
		return "Enable protected-mode unless Redis is isolated and authenticated."
	case "redis_missing_auth", "redis_compose_missing_auth":
		return "Configure Redis AUTH or ACL users before exposing Redis beyond localhost."
	case "redis_bind_all_interfaces":
		return "Bind Redis to localhost or a private interface."
	case "redis_port_exposed":
		return "Avoid publishing port 6379 publicly; keep Redis on a private network."
	case "redis_appendonly_disabled":
		return "Enable appendonly when the workload needs stronger durability."
	case "redis_url_no_auth":
		return "Use an authenticated Redis URL for shared or deployed environments."
	case "postgres_trust_auth":
		return "Replace trust authentication with SCRAM, password, certificate, or peer auth."
	case "postgres_hba_open_cidr":
		return "Restrict pg_hba.conf CIDR ranges to trusted application networks."
	case "postgres_listen_all_interfaces":
		return "Bind PostgreSQL to private interfaces unless public access is explicitly required."
	case "postgres_ssl_disabled", "postgres_url_ssl_disabled":
		return "Enable SSL for networked PostgreSQL connections."
	case "postgres_weak_password":
		return "Set a strong POSTGRES_PASSWORD or source it from a secret manager."
	case "postgres_port_exposed":
		return "Avoid publishing port 5432 publicly; keep PostgreSQL on a private network."
	case "module_timeout":
		return "Increase the module timeout or reduce the scan scope for that module."
	default:
		return ""
	}
}
