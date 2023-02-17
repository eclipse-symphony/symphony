module github.com/azure/symphony/cli

go 1.19

replace github.com/azure/symphony/api => ../api

require github.com/spf13/cobra v1.6.1

require (
	golang.org/x/exp v0.0.0-20220929160808-de9c53c655b9 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	github.com/azure/symphony/api v0.0.0
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jedib0t/go-pretty/v6 v6.4.2
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0
)
