module github.com/eclipse-symphony/symphony/cli

go 1.22.4

toolchain go1.22.6

replace github.com/eclipse-symphony/symphony/api => ../api

replace github.com/eclipse-symphony/symphony/coa => ../coa

require github.com/spf13/cobra v1.8.1

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/eclipse-symphony/symphony/coa v0.0.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.50.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	helm.sh/helm/v3 v3.15.4 // indirect
	k8s.io/apimachinery v0.30.3 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/princjef/mageutil v1.0.0
)

require (
	github.com/eclipse-symphony/symphony/api v0.0.0
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.4.2
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.29.0 // indirect
	sigs.k8s.io/yaml v1.4.0
)
