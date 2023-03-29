package models

import (
	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/utils"
)

type azResource struct {
	Location string            `json:"location,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Id       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Type     string            `json:"type,omitempty"`
}

// TODO: Find a better way to do this
func (r *azResource) GetSubscription() string {
	return utils.GetSubscriptionFromResourceId(r.Id)
}

func (r *azResource) GetResourceGroup() string {
	return utils.GetResourceGroupFromResourceId(r.Id)
}

func (r *azResource) GetResourceName() string {
	return r.Name
}
