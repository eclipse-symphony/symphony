//go:build darwin
// +build darwin

package rust

/*
#cgo LDFLAGS: -L./target/x86_64-apple-darwin/release -lsymphony -lm -ldl -lpthread
*/
import "C"
