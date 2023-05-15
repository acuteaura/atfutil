package main

import (
	"os"

	"ops-networking/pkg/cli/atfutil"
)

func main() {
	if err := atfutil.Command().Execute(); err != nil {
		os.Exit(1)
	}
}
