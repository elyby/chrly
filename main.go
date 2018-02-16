package main

import (
	"runtime"

	"github.com/elyby/chrly/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
