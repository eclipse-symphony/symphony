//go:build linux
// +build linux

package main

func main() {
	if err := mainLogic(); err != nil {
		panic(err)
	}
}
