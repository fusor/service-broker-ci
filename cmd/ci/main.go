package main

import (
	"os"

	"github.com/fusor/service-broker-ci/pkg/ci"
)

func main() {
	c, err := ci.CreateCi()
	if err != nil {
		os.Exit(1)
	}
	c.Run()
}
