package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-envdir /path/to/env/dir command arg1 arg2")
		os.Exit(1)
	}
	dir := os.Args[1]
	cmd := os.Args[2:]
	if len(cmd) == 0 {
		fmt.Println("No command specified")
		os.Exit(1)
	}
	env, err := ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading environment directory: %v\n", err)
		os.Exit(1)
	}
	os.Exit(RunCmd(cmd, env))
}
