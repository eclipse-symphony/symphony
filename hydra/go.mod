module github.com/eclipse-symphony/symphony/hydra

go 1.19

replace github.com/eclipse-symphony/symphony/api => ../api

replace github.com/eclipse-symphony/symphony/coa => ../coa

require (
	github.com/eclipse-symphony/symphony/api v0.0.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/eclipse-symphony/symphony/coa v0.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/exp v0.0.0-20220929160808-de9c53c655b9 // indirect
	helm.sh/helm/v3 v3.10.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
