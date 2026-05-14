package main

import (
	"fmt"
	"os"
	"os/exec"

	"stack/internal/output"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the local environment for scanning readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		results := []output.CheckResult{}

		// Check Docker
		dockerRes := output.CheckResult{Name: "Docker Engine"}
		if _, err := exec.LookPath("docker"); err != nil {
			dockerRes.Status = "error"
			dockerRes.Message = "Docker binary not found in PATH"
			dockerRes.Hint = "Install Docker Desktop or Docker Engine"
		} else if err := exec.Command("docker", "info").Run(); err != nil {
			dockerRes.Status = "warning"
			dockerRes.Message = "Docker daemon is not running or accessible"
			dockerRes.Hint = "Start Docker Desktop and ensure your user has permissions"
		} else {
			dockerRes.Status = "success"
			dockerRes.Message = "Docker is ready"
		}
		results = append(results, dockerRes)

		// Check Kubernetes
		k8sRes := output.CheckResult{Name: "Kubernetes (kubectl)"}
		if _, err := exec.LookPath("kubectl"); err != nil {
			k8sRes.Status = "warning"
			k8sRes.Message = "kubectl not found in PATH"
			k8sRes.Hint = "Required only if scanning K8s manifests or clusters"
		} else {
			k8sRes.Status = "success"
			k8sRes.Message = "kubectl is ready"
		}
		results = append(results, k8sRes)

		// Check Configuration
		configRes := output.CheckResult{Name: "stack Config"}
		if _, err := os.Stat(".stack.yaml"); os.IsNotExist(err) {
			configRes.Status = "warning"
			configRes.Message = "No .stack.yaml found in current directory"
			configRes.Hint = "Run 'stack init' to create a default configuration"
		} else {
			configRes.Status = "success"
			configRes.Message = "Config file found and accessible"
		}
		results = append(results, configRes)

		fmt.Fprint(cmd.OutOrStdout(), output.RenderDoctor(results))
		return nil
	},
}
