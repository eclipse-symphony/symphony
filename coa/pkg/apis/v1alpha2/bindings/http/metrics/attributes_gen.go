//go:build !azure

package metrics

import "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"

func ConstructApiOperationStatusAttributes(
	context contexts.ActivityLogContext,
	operation string,
	operationType string,
	statusCode int,
	formatStatusCode string,
) map[string]any {
	return mergeAttrs(
		Deployment(
			operation,
			operationType,
		),
		Status(
			statusCode,
			formatStatusCode,
		),
	)
}
