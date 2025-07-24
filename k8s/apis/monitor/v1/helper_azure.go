//go:build azure

package v1

import (
	"context"
	"fmt"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetGlobalDiagnosticResourceInCluster(sourceResourceAnnotations map[string]string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (*Diagnostic, error) {
	// skip check the sourceResourceAnnotations - no need to check since all objects are mapped to global unique diagnostic resource
	// no need to check whether the edge location is the same as the source resource
	annotationFilterFunc := func(diagResourceAnnotations map[string]string) bool {
		edgeLocation := diagResourceAnnotations[constants.AzureEdgeLocationKey]
		return edgeLocation != "" // validate azure diagnostic resource has to have edge location annotation
	}
	return getGlobalDiagnosticResourceInCluster(annotationFilterFunc, k8sClient, ctx, logger)
}

func GetDiagnosticCloudResourceInfo(sourceResourceAnnotations map[string]string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (string, string) {
	d, err := GetDiagnosticCustomResourceFromCache(sourceResourceAnnotations, k8sClient, ctx, logger)
	if err != nil {
		diagnostic.InfoWithCtx(logger, ctx, "Failed to get diagnostic resource", "error", err)
		return "", ""
	}
	if d != nil {
		return d.Annotations[constants.AzureResourceIdKey], d.Annotations[constants.AzureLocationKey]
	}
	return "", ""
}

func ValidateDiagnosticResourceAnnoations(annotations map[string]string) *field.Error {
	edgeLocation := annotations[constants.AzureEdgeLocationKey]
	if edgeLocation == "" {
		return field.Required(field.NewPath("metadata.annotations").Child(constants.AzureEdgeLocationKey), "Azure Edge Location is required")
	} else {
		return nil
	}
}

func GenerateDiagnosticResourceUniquenessFieldError(newDiagnosticResName string, newDiagnosticResNamespace string, globalDiagnostic *Diagnostic) *field.Error {
	if globalDiagnostic == nil || newDiagnosticResName == "" || newDiagnosticResNamespace == "" {
		return nil
	}

	edgeLocation := globalDiagnostic.Annotations[constants.AzureEdgeLocationKey]
	return field.Invalid(field.NewPath("metadata.name"), newDiagnosticResName, fmt.Sprintf("Cannot create Diagnostic resource name: %s in namespace: %s, Diagnostic resource already exists in this cluster, name: %s, namespace: %s, edgeLocation: %s", newDiagnosticResName, newDiagnosticResNamespace, globalDiagnostic.Name, globalDiagnostic.Namespace, edgeLocation))
}
