package main

import (
	"os"

	"github.com/speedyhoon/gci/cmd/gci"
)

var (
	version = "0.3"
)

func main() {
	e := gci.NewExecutor(version)

	err := e.Execute()
	if err != nil {
		os.Exit(1)
	}
}
