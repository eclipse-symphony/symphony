//go:build darwin
// +build darwin

package rust

/*
#cgo LDFLAGS: -L./target/x86_64-apple-darwin/release -lrust_binding -lm -ldl -lpthread
*/
import "C"
