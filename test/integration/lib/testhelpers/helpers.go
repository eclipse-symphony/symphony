package testhelpers

import (
	"context"

	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
)

func DumpClusterState(ctx context.Context) {
	shell.Exec(ctx, "kubectl get all -A -o wide")
	shell.Exec(ctx, "kubectl get events -A --sort-by=.metadata.creationTimestamp")
	shell.Exec(ctx, "kubectl get targets.fabric.symphony -A -o yaml")
	shell.Exec(ctx, "kubectl get solutions.solution.symphony -A -o yaml")
	shell.Exec(ctx, "kubectl get instances.solution.symphony -A -o yaml")
	shell.Exec(ctx, "helm list -A -o yaml")
}

func CleanupManifests(ctx context.Context) error {
	return shell.ExecAll(
		ctx,
		"kubectl delete instances.solution.symphony --all -A",
		"kubectl delete targets.fabric.symphony --all -A",
		"kubectl delete solutions.solution.symphony --all -A",
	)
}
