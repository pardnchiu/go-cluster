package main

import (
	"fmt"
	"os"

	"github.com/pardnchiu/pdcluster/internal/node"
)

func main() {
	health, err := node.CheckHealth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(health.([]byte)))
}
