package helm

import "github.com/eclipse-symphony/symphony/packages/testutils/conditions/jq"

var (
	DeployedCondition = jq.Equality(".info.status", "deployed")
	FailedCondition   = jq.Equality(".info.status", "failed")
)
