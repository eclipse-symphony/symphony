/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package actions

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	federationv1 "gopls-workspace/apis/federation/v1"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/go-logr/logr"
	giterror "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CatalogReconciler reconciles a Site object
type CatalogEvalReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient
}

//+kubebuilder:rbac:groups=infrastructure.hybridaks.microsoft.com,resources=hybridaksclusterlistusercredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.hybridaks.microsoft.com,resources=hybridaksclusterlistusercredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.hybridaks.microsoft.com,resources=hybridaksclusters,verbs=get;list

func (r *CatalogEvalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ret ctrl.Result, reterr error) {
	log := ctrllog.FromContext(ctx)

	result := &federationv1.CatalogEvalExpression{}

	err := r.Get(ctx, req.NamespacedName, result)
	if err != nil {
		log.Error(err, "failed to parse CatalogEvalExpression")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// update logger and status with correlation id once we successfully get the CR
	logger = clusterutils.AttachCorrelationIDToLogger(result, logger)
	status = hybridAksStatus.NewStatusReporter(r.Client, logger, r.Recorder, result)

	patchHelper, err := patch.NewHelper(result, r.Client)
	if err != nil {
		return reconcile.Result{}, giterror.Wrap(err, "failed to init patch helper")
	}

	// Always issue a patch when exiting this function so changes to the
	// resource are patched back to the API server.
	defer func() {
		if reterr == nil {
			result.Status.ActionStatus.Status = infrastructurev1beta1.StatusSucceeded
		} else {
			result.Status.ActionStatus.Status = infrastructurev1beta1.StatusFailed
		}
		if err := patchHelper.Patch(ctx, result); err != nil && !apierrors.IsNotFound(err) {
			if reterr == nil {
				reterr = err
			}
			logger.Error(err, "patch failed", "HybridAKSClusterListUserCredential", result.Name)
		}
	}()

	// Clean up the cr if it is older than 1 day
	if clusterutils.CheckNeedtoDelete(result.ObjectMeta) {
		// Delete the CR
		logger.Info("Deleting the HybridAKSClusterListUserCredential as it is older than 1 day", "CR", req.NamespacedName)
		if err := r.Delete(ctx, result); err != nil {
			logger.Error(err, "Error deleting the HybridAKSClusterListUserCredential", "CR", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if result.DeletionTimestamp.IsZero() {

		// Initialize status
		kubeconfigs := infrastructurev1beta1.KubeConfigs{}
		if result.Status.ActionStatus.OperationID != result.GetOperationID() {
			result.Status.ActionStatus.OperationID = result.GetOperationID()
			status.Reset(ctx)
		}

		// Check cluster existing and get tenantid
		hybridAksCluster := &infrastructurev1beta1.HybridAKSCluster{}
		cmKey := client.ObjectKey{
			Namespace: result.Spec.ResourceRef.Namespace,
			Name:      result.Spec.ResourceRef.Name,
		}
		err = r.Get(ctx, cmKey, hybridAksCluster)
		if apierrors.IsNotFound(err) {
			logger.Error(err, "hybridakscluster not found", "Name", result.Spec.ResourceRef.Name)
			status.Error(ctx, "GetCluster", *infrastructurev1beta1.ErrorInvalidResourceReferenceClusterName(result.Spec.ResourceRef.Name, err))
			return ctrl.Result{}, err
		} else if err != nil {
			logger.Error(err, "Error while fetching hybridakscluster", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "GetCluster", err)
			return ctrl.Result{}, err
		}
		tenantId, ok := hybridAksCluster.ObjectMeta.Annotations["management.azure.com/tenantId"]
		if !ok || tenantId == "" {
			err = errors.New("TenantId not set")
			logger.Error(err, "Error while fetching tenantid from hybridakscluster", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "GetTenantId", err)
			return ctrl.Result{}, err
		}
		parentResourceId, ok := hybridAksCluster.ObjectMeta.Annotations["management.azure.com/parentResourceId"]
		if !ok || parentResourceId == "" {
			err = errors.New("ParentResourceId not set")
			logger.Error(err, "Error while fetching parentResourceId from hybridakscluster", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "GetParentResourceId", err)
			return ctrl.Result{}, err
		}

		// Fetch cluster secret
		clusterKey := client.ObjectKey{Namespace: result.Spec.ResourceRef.Namespace, Name: result.Spec.ResourceRef.Name}
		adminKubeConfigBytes, err := kcfg.FromSecret(context.Background(), r.Client, clusterKey)
		if apierrors.IsNotFound(err) {
			logger.Error(err, "cluster kubeconfig secret not found", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "GetClusterSecret", err)
			return ctrl.Result{}, err
		} else if err != nil {
			logger.Error(err, "Error while processing kcfg.FromSecret", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "GetClusterSecret", err)
			return ctrl.Result{}, err
		}

		// Convert to AAD kubeconfig
		aadKubeConfigBytes, err := ConvertToAADKubeconfig(adminKubeConfigBytes, tenantId, result.Spec.ResourceRef.Name, parentResourceId)
		if err != nil {
			logger.Error(err, "Error while processing ConvertToAADKubeconfig", "Name", result.Spec.ResourceRef.Name)
			status.ErrorReconcile(ctx, "ConvertToAADKubeconfig", err)
			return ctrl.Result{}, err
		}

		kubeValue := base64.StdEncoding.EncodeToString(aadKubeConfigBytes)

		// Update status with results
		credentialOutput := &infrastructurev1beta1.CredentialOutput{}
		credentialOutput.Value = kubeValue
		credentialOutput.Name = UserKubeConfigName
		kubeconfigs.KubeConfigs = append(kubeconfigs.KubeConfigs, *credentialOutput)
		result.Status.ActionStatus.Output = kubeconfigs

	}

	return ctrl.Result{}, nil
}

func ConvertToAADKubeconfig(data []byte, tenantId string, clustername string, parentResourceId string) ([]byte, error) {
	kubeconfig := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &kubeconfig)
	if err != nil {
		return nil, err
	}

	//Set user
	users, ok := kubeconfig["users"].([]interface{})
	if !ok {
		return nil, errors.New("kubeconfig is not in the expected format")
	}
	user, ok := users[0].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("kubeconfig is not in the expected format")
	}

	user["name"] = AADUser
	/*
	   exec:
	     apiVersion: client.authentication.k8s.io/v1beta1
	     args:
	     - get-token
	     - --login
	     - interactive
	     - --server-id
	     - 6256c85f-0aad-4d50-b960-e6e9b21efe35
	     - --client-id
	     - 3f4439ff-e698-4d6d-84fe-09c9d574f06b
	     - --tenant-id
	     - 5d110407-81fa-4824-9a46-5f2eb31ebd87
	     - --environment
	     - AzurePublicCloud
	     - --pop-enabled
	     - --pop-claims
	     - u=/subscriptions/f60be53e-e5ed-45da-a66e-34ef8fd8a3e5/resourceGroups/mztestccrbacrg2/providers/Microsoft.Kubernetes/connectedClusters/mztestccrbac2
	     command: kubelogin
	     env: null
	     installHint: |2

	       kubelogin is not installed which is required to connect to AAD enabled cluster.

	       To learn more, please go to https://azure.github.io/kubelogin/
	     provideClusterInfo: false
	*/
	user["user"] = map[string]interface{}{
		"exec": map[string]interface{}{
			"apiVersion": "client.authentication.k8s.io/v1beta1",
			"args": []string{
				"get-token",
				"--login",
				"interactive",
				"--server-id",
				ServerAppID,
				"--client-id",
				ClientAppID,
				"--tenant-id",
				tenantId,
				"--environment",
				"AzurePublicCloud",
				"--pop-enabled",
				"--pop-claims",
				fmt.Sprintf("u=%s", parentResourceId),
			},
			"command":            "kubelogin",
			"env":                nil,
			"installHint":        InstallHint,
			"provideClusterInfo": false,
		},
	}

	//Set context
	contexts, ok := kubeconfig["contexts"].([]interface{})
	if !ok {
		return nil, errors.New("kubeconfig is not in the expected format")
	}
	context, ok := contexts[0].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("kubeconfig is not in the expected format")
	}
	context["context"] = map[string]interface{}{
		"cluster": clustername,
		"user":    AADUser,
	}
	context["name"] = AADUser + "@" + clustername

	//Set current context
	kubeconfig["current-context"] = AADUser + "@" + clustername

	aadKubeConfigBytes, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		return nil, err
	}
	return aadKubeConfigBytes, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HybridAKSClusterListUserCredentialReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		WithLogConstructor(r.ConstructLogger).
		For(&infrastructurev1beta1.HybridAKSClusterListUserCredential{}).
		Complete(r)
}

func (r *HybridAKSClusterListUserCredentialReconciler) ConstructLogger(req *reconcile.Request) logr.Logger {
	log := r.Log.WithName("")
	if req == nil {
		return log
	}
	log = log.WithValues("HybridAKSClusterListUserCredential", req.NamespacedName)
	cxt := context.Background()
	result := &infrastructurev1beta1.HybridAKSClusterListUserCredential{}
	err := r.Get(cxt, req.NamespacedName, result)
	if err != nil {
		log.Error(err, "failed to get HybridAKSClusterListUserCredential")
		return log
	}
	log = clusterutils.AttachCorrelationIDToLogger(result, log)
	return log
}
