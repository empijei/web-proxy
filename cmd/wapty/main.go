// Package main runs WAPTY.
package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/empijei/web-proxy/wapty"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, ".wapty")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		panic(err)
	}
	if err := wapty.Run(context.Background(), dir); err != nil {
		panic(err)
	}
}
