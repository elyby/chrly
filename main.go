package main

import (
	"runtime"

	"elyby/minecraft-skinsystem/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
