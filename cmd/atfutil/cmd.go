package main

import (
	"os"

	"atfutil/pkg/cli/atfutil"
)

func main() {
	if err := atfutil.Command().Execute(); err != nil {
		os.Exit(1)
	}
}
