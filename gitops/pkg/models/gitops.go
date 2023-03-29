package models

import (
	"time"
)

type BaseGitOpsProperties struct {
	Interval            int    `json:"interval,omitempty"`
	ManagedIdentityName string `json:"managedIdentityName,omitempty"`
}

func (b *BaseGitOpsProperties) GetInterval() time.Duration {
	return time.Duration(b.Interval) * time.Second
}
