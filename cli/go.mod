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
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.50.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	helm.sh/helm/v3 v3.15.4 // indirect
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
	golang.org/x/sys v0.28.0 // indirect
	sigs.k8s.io/yaml v1.4.0
)
