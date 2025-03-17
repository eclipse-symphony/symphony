//go:build azure

package metrics

import "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"

func SLI(
	customerResourceId string,
	locationId string,
) map[string]any {
	return map[string]any{
		"CustomerResourceId": customerResourceId,
		"LocationId":         locationId,
	}
}

func ConstructApiOperationStatusAttributes(
	context contexts.ActivityLogContext,
	operation string,
	operationType string,
	statusCode int,
	formatStatusCode string,
) map[string]any {

	customerResourceId := context.GetResourceCloudId()
	locationId := context.GetResourceCloudLocation()

	return mergeAttrs(
		SLI(
			customerResourceId,
			locationId,
		),
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
