//go:build !azure

package v1

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetGlobalDiagnosticResourceInCluster(sourceResourceAnnotations map[string]string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (*Diagnostic, error) {
	annotationFilterFunc := func(diagResourceAnnotations map[string]string) bool {
		return true
	}
	return getGlobalDiagnosticResourceInCluster(annotationFilterFunc, k8sClient, ctx, logger)
}

func GetDiagnosticCloudResourceInfo(sourceResourceAnnotations map[string]string, k8sClient client.Reader, ctx context.Context, logger logr.Logger) (string, string) {
	return "", ""
}

func ValidateDiagnosticResourceAnnoations(annotations map[string]string) *field.Error {
	return nil
}

func GenerateDiagnosticResourceUniquenessFieldError(newDiagnosticResName string, newDiagnosticResNamespace string, globalDiagnostic *Diagnostic) *field.Error {
	if globalDiagnostic == nil || newDiagnosticResName == "" || newDiagnosticResNamespace == "" {
		return nil
	}

	return field.Invalid(field.NewPath("metadata.name"), newDiagnosticResName, fmt.Sprintf("Cannot create Diagnostic resource name: %s in namespace: %s, Diagnostic resource already exists in this cluster, name: %s, namespace: %s", newDiagnosticResName, newDiagnosticResNamespace, globalDiagnostic.Name, globalDiagnostic.Namespace))
}
