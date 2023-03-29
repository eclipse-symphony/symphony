package utils

import "strings"

func GetSubscriptionFromResourceId(id string) string {
	parts := strings.Split(id, "/")
	return parts[2]
}

func GetResourceGroupFromResourceId(id string) string {
	parts := strings.Split(id, "/")
	return parts[4]
}

func GetResourceNameFromResourceId(id string) string {
	parts := strings.Split(id, "/")
	return parts[len(parts)-1]
}
